# wsvnc

This is an executable program.

### Build

```shell
$ dep ensure
$ go build
```

### Run

```shell
$ ./wsvnc
```

Notice: The current host must run a redis server listening on `127.0.0.1:5900`, or you can modify it by the cli option `redis`.

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
        The option name of the request URL (default "token")
  -path string
        The path of the request URL (default "/websockify")
  -redis string
        The redis connection url (default "redis://localhost:6379/0")
```
