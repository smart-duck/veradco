# veradco

## Overview

Veradco a.k.a. Versatile Admission Controller is an admission controller that is expandable via a plugin system.

 ## Plugin

Plugins are simple Golang plugins implementing the following interface:

```
type VeradcoPlugin interface {
	Init(params string)
	Info() string
}
```

### Loading

Plugins are loaded in a docker image that is loaded via an init container. Refer to examples to see how to do. There are other ways to do it. Given the size of a plugin, it is very unlikely that you are able to put it in a config map (limited to 1 MB).

## Configuration

The configuration defines the plugins to use and their configuration.

```
banner: "veradcoBanner"
plugins:
- name: "plug1"
  path: "/home/lobuntu/go/src/test_plugin/plug1/plug.so"
  params: "--verbose -n veradco"
- name: "plug2"
  path: "/home/lobuntu/go/src/test_plugin/plug2/plug.so"
  params: "init -n veradco"
- name: "plug_ext"
  path: "/home/lobuntu/go/src/test_plugin/plug_ext.so"
  params: "Plugin created by anyone"
```

## Test with kind

Local registry:
https://kind.sigs.k8s.io/docs/user/local-registry/

## Dev

git clone https://github.com/douglasmakey/admissioncontroller.git

See also kubernetes official doc ac...

## Issue Plugin failed

```
$ kk apply -f pods/01_fail_pod_creation_test.yaml 
I0901 16:06:37.666248       1 handlers.go:95] Webhook [/mutate/pods - CREATE] - Allowed: true
I0901 16:06:37.689117       1 create.go:41] Unable to load plugin: plugin.Open("/plugs/plug1/plug"): plugin was built with a different version of package github.com/gogo/protobuf/proto
2022/09/01 16:06:37 http: panic serving 10.244.0.1:18104: runtime error: invalid memory address or nil pointer dereference
goroutine 60 [running]:
net/http.(*conn).serve.func1()
        /usr/local/go/src/net/http/server.go:1850 +0xbf
panic({0xeeb460, 0x1598cb0})
        /usr/local/go/src/runtime/panic.go:890 +0x262
plugin.lookup(...)
        /usr/local/go/src/plugin/plugin_dlopen.go:138
plugin.(*Plugin).Lookup(...)
        /usr/local/go/src/plugin/plugin.go:40
github.com/smart-duck/veradco/pods.validateCreate.func1(0xc0002b4000)
        /app/pods/create.go:44 +0x13a
github.com/smart-duck/veradco.wrapperExecution(0x3?, 0xc0001009b0?)
        /app/hook.go:48 +0x28
github.com/smart-duck/veradco.(*Hook).Execute(0x7f2d286b3bd8?, 0xc0002ad000?)
        /app/hook.go:31 +0x45
github.com/smart-duck/veradco/http.(*admissionHandler).Serve.func1({0x10ef5a0, 0xc0000b8460}, 0xc0000c1000)
        /app/http/handlers.go:62 +0x3e5
net/http.HandlerFunc.ServeHTTP(0xc0000b8460?, {0x10ef5a0?, 0xc0000b8460?}, 0xfa267e?)
        /usr/local/go/src/net/http/server.go:2109 +0x2f
net/http.(*ServeMux).ServeHTTP(0xc00003a734?, {0x10ef5a0, 0xc0000b8460}, 0xc0000c1000)
        /usr/local/go/src/net/http/server.go:2487 +0x149
net/http.serverHandler.ServeHTTP({0x10ea8f0?}, {0x10ef5a0, 0xc0000b8460}, 0xc0000c1000)
        /usr/local/go/src/net/http/server.go:2947 +0x30c
net/http.(*conn).serve(0xc0000008c0, {0x10eff40, 0xc00008a300})
        /usr/local/go/src/net/http/server.go:1991 +0x607
created by net/http.(*Server).Serve
        /usr/local/go/src/net/http/server.go:3102 +0x4db
pod/webserver created
```

