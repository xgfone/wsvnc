# wsvnc

This is an executable program, which may be as the server of [noVNC](https://github.com/novnc/noVNC) instead of [websockify](https://github.com/novnc/websockify).

### Build

```shell
$ dep ensure
$ go build
```

### Run

```shell
$ ./wsvnc
```

Notice: The current host must run a redis server listening on `127.0.0.1:6379`, or you can modify it by the cli option `redis`.

```shell
[root@localhost ~]# ./wsvnc -h
Usage of ./wsvnc:
  -addr string
        The listen address (default ":5900")
  -logfile string
        The path of the log file
  -loglevel string
        The level of the log, such as debug, info, etc (default "debug")
  -option string
        The token option name of the request URL (default "token")
  -path string
        The path of the request URL (default "/websockify")
  -redis string
        The redis connection url (default "redis://localhost:6379/0")
```
