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
)

func init() {
	flag.StringVar(&logFile, "logfile", "", "The path of the log file")
	flag.StringVar(&logLevel, "loglevel", "debug", "The level of the log, such as debug, info, etc")
	flag.StringVar(&listenAddr, "addr", ":5900", "The listen address")
	flag.StringVar(&redisURL, "redis", "redis://localhost:6379/0", "The redis connection url")
	flag.StringVar(&urlPath, "path", "/websockify", "The path of the request URL")
	flag.StringVar(&urlOption, "option", "token", "The token option name of the request URL")
}

func main() {
	defer lifecycle.Stop()
	flag.Parse()

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

	go signal2.HandleSignal()
	logger.Info("Listening", "addr", listenAddr)
	if err := http2.ListenAndServe(listenAddr, nil); err != nil {
		logger.Fatal("ListenAndServe", "err", err)
	}
}
