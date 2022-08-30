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