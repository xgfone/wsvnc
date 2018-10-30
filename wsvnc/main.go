package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/xgfone/go-tools/lifecycle"
	"github.com/xgfone/go-tools/net2/http2"
	"github.com/xgfone/go-tools/signal2"
	"github.com/xgfone/miss"
	"github.com/xgfone/wsvnc"
)

var logger miss.Logger

var (
	logFile    string
	logLevel   string
	listenAddr string
	redisURL   string
	urlPath    string
	urlOption  string
	certFile   string
	keyFile    string
	connection string
)

func init() {
	flag.StringVar(&logFile, "logfile", "", "The path of the log file")
	flag.StringVar(&logLevel, "loglevel", "debug", "The level of the log, such as debug, info, etc")
	flag.StringVar(&listenAddr, "addr", ":5900", "The listen address")
	flag.StringVar(&redisURL, "redis", "redis://localhost:6379/0", "The redis connection url")
	flag.StringVar(&urlPath, "path", "/websockify", "The path of the request URL")
	flag.StringVar(&urlOption, "option", "token", "The token option name of the request URL")
	flag.StringVar(&certFile, "cert", "", "The path of the cert file")
	flag.StringVar(&keyFile, "key", "", "The path of the key file")
	flag.StringVar(&connection, "connection", "",
		"Enable the authentication token to get the connection number")
}

func main() {
	defer lifecycle.Stop()
	flag.Parse()

	go signal2.HandleSignal()

	// Handle the logging
	level := miss.NameToLevel(logLevel)
	var out io.Writer = os.Stdout
	if logFile != "" {
		file, err := miss.SizedRotatingFileWriter(logFile, 1024*1024*1024, 30)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		lifecycle.Register(func() { file.Close() })
		out = file
	}
	encConf := miss.EncoderConfig{IsLevel: true, IsTime: true}
	encoder := miss.KvTextEncoder(out, encConf)
	logger = miss.New(encoder).Level(level).Cxt("caller", miss.Caller())
	wsvnc.LOG = logger

	// Handle the redis client
	redisOpt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("can't parse redis URL", "url", redisURL, "err", err)
		return
	}
	redisClient := redis.NewClient(redisOpt)
	lifecycle.Register(func() { redisClient.Close() })

	wsconf := wsvnc.ProxyConfig{
		GetBackend: func(r *http.Request) string {
			if vs := r.URL.Query()[urlOption]; len(vs) > 0 {
				token, err := redisClient.Get(vs[0]).Result()
				if err == nil {
					return token
				}
				logger.Error("redis GET error", "err", err)
			}
			return ""
		},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	handler := wsvnc.NewWebsocketVncProxyHandler(wsconf)
	http.Handle(urlPath, handler)
	http.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		if connection != "" {
			if http2.GetQuery(r.URL.Query(), "token") != connection {
				http2.String(w, http.StatusForbidden, "authentication failed")
				return
			}
		}
		http2.String(w, http.StatusOK, "%d", handler.Connections())
	})

	tlsfiles := []string{}
	if certFile != "" && keyFile != "" {
		tlsfiles = []string{certFile, keyFile}
	} else if certFile != "" || keyFile != "" {
		logger.Warn("The cert and key file is incomplete and don't use TLS")
	}

	logger.Info("Listening", "addr", listenAddr, "tls", len(tlsfiles) != 0)
	if err := http2.ListenAndServe(listenAddr, nil, tlsfiles...); err != nil {
		logger.Fatal("ListenAndServe", "err", err)
	}
}
