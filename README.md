# wsvnc

It supports a HTTP Handler to implement the VNC proxy over websocket.

you can use it easily as following:

```go
tokens := map[string]string {
	"token1": "host1:port1",
	"token2": "host2:port2",
	// ...
}
wsconf := wsvnc.ProxyConfig{
	GetBackend: func(r *http.Request) string {
		if vs := r.URL.Query()["token"]; len(vs) > 0 {
			return tokens[vs[0]]
		}
		return ""
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
handler := wsvnc.NewWebsocketVncProxyHandler(wsconf)
http.Handle("/websockify", handler)
http.ListenAndServe(":5900", nil)
```

Then, you can use [noVNC](https://github.com/novnc/noVNC) by the url `http://127.0.0.1:5900/websockify?token=token1` to connect to "host1:port1" over websocket.

**NOTICE:** The sub-package [wsvnc](https://github.com/xgfone/wsvnc/tree/master/wsvnc) implements the function above, but using the redis to store the mapping between `TOKEN` and `HOST:PORT`.
