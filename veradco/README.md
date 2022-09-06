# veradco

## Overview

Veradco a.k.a. Versatile Admission Controller is an admission controller that is expandable via a plugin system. It handles Mutating and Validating webhooks that you can extend by developing your own plugins or by using third-party webhooks or the ones that are built-in.

With Veradco, you take advantage of the full power of Mutating and Validating webhooks in a simple and flexible way. You only need to write the functional part. Plugin are written in golang, can be packaged in a ConfigMap and are built on the fly by the provided init container.

To help you develop your plugins, examples are provided in the veradco repositories. They cover many use cases.

## Cons

### Plugins are big

Same size than veradco itself.

```
/ # ls -lhRt /release/
/release/:
total 25M    
drwxr-xr-x    2 root     root        4.0K Sep  4 11:20 plugins
-rwxr-xr-x    1 root     root       25.1M Sep  4 11:19 veradcod

/release/plugins:
total 90M    
-rw-r--r--    1 root     root       23.7M Sep  4 11:20 built-in-user_plugin.so
-rw-r--r--    1 root     root       24.3M Sep  4 11:20 built-in-plug1.so
-rw-r--r--    1 root     root       18.2M Sep  4 11:20 built-in-generic.so
-rw-r--r--    1 root     root       23.7M Sep  4 11:20 built-in-basic.so
```

 ## Plugin

A plugin is a piece of Go code that implements the following interface:

```
type VeradcoPlugin interface {
        Init(configFile string) error
        Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
        Summary() string
}
```

### Loading

Plugins are loaded thanks to a ConfigMap. Refer to examples to see how to do. There are other ways to do it.

## Configuration

The configuration defines the plugins to use and their configuration. Here is an example of a ConfigMap.

The configuation embeds some basic filtering fields so that your plugin is called or not. It's up to you in the code of your plugin to do the rest. It is as flexible as a programming language is. Filtering fileds work as regular expressions.

TODO update

```
plugins:
- name: "Basic"
  # The path of the plugin.
  # If the plugin needs to be build from the parameter plug.go (below), it is MANDATORY that is path is /app/external_plugins. It will be built by the Init Container.
  # For a built-in plugin the path is /app/plugins
  # For a plugin you built yourself the path is as you want. We advise you to build your plugin by using the init container docker image to avoid infrequent issues with Golang plugins compatibility. TODO: documentation.
  path: "/app/plugins/built-in-basic.so"
  # It is possible to make your regular expression act as a reverse pattern. To do so, just prefix it with (!~)
  # Except that, it works as defined in the Golang regexp package.
  # By example, "(!~)(?i)test" matches that the value does not contain "test" whatever the case is.
  # A regular expression to define the resources on which the plugin is called
  resources: "Pod"
  # A regular expression to define the operations on which the plugin is called
  # Example: "CREATE|UPDATE"
  # It's up to the plugin to manage different operations in its code
  operations: "CREATE|UPDATE"
  # A regular expression to define the namespaces in the scope of the plugin
  # Example: "kube-system|default"
  namespaces: ".*"
  # Filter only on resources having some labels.
  # Both key and value are regular expressions
  # This parameter is used together with the one of the AdmissionRequest: dryRun = dryRun || AdmissionRequest.DryRun
  dryRun: false
  # The configuration of the plugin. Passed to the plugin via the Init function of the plugin.
  configuration: |
    Basic configuration
  # A regular expression that define the scope of the plugin.
  # "Validating|Mutating" fits to both scopes
  scope: "Validating"
```

## Setup go environment

```
go mod init github.com/smart-duck/veradco
go mod tidy
```

## veradco logs

```
kk logs $(kk get po -n veradco | grep veradco | grep -o -E "veradco-[0-9a-f]+[^ ]+") --follow -n veradco &
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

# TODO

To add to plugin/Execute:
- DryRun, conf/Execute/AdmissionRequest https://pkg.go.dev/k8s.io/kubernetes/pkg/apis/admission#AdmissionRequest
- Operation CREATE DELETE UPDATE CONNECT AdmissionRequest

# Check there is no build error

```
go build -o /dev/null ../cmd/serverd/main.go
```

