# veradco

## Overview

Veradco a.k.a. Versatile Admission Controller is an admission controller that is expandable via a plugin system. It handles Mutating and Validating webhooks that you can extend by developing your own plugins or by using some third-party ones or the ones that are built-in.

With Veradco, you take advantage of the full power of Mutating and Validating webhooks in a simple and flexible way. You only need to write the functional part. Plugin are written in golang, can be packaged in a ConfigMap and are built on the fly by the provided init container. A big advantage is that you don't need to learn a new programming/configuration language and so, you are not stuck in a cramped and finite universe.

You also take advantage of the built-in monitoring that gives you statistics about plugins such as call frequency or execution time. These metrics are prefixed by veradco. You can scrape them towards Prometheus via a ServiceMonitor.

To help you develop your plugins, examples are provided in the veradco repository (built-in ones). They cover many use cases.

## Basic operation of Veradco

Basically, Veradco works as follow:
```
+---------------+                   +---------+                       
| K8sAPIserver  |                   | Veradco |                       
+---------------+                   +---------+                       
        |                                |                            
        | Validate or Mutate             |                            
        |------------------------------->|                            
        |                                |                            
        |                                | Excecute concerned plugins 
        |                                |--------------------------- 
        |                                |                          | 
        |                                |<-------------------------- 
        |                                |                            
        |     Responds to the API server |                            
        |<-------------------------------|                            
        |                                |
```

Description of the operation:
1. An admission review is sent to Veradco by the API server in order it validates or mutates a Kubernetes specification.
2. Veradco execute the plugins that are in the scope of the admission review. It sums up the pluings responses.
3. Veradco responds to the API server on behalf of executed plugins.

## Repository structure

The repository is made of 3 main folders:
- veradco: the Golang code of the Veradco Admission Controller
- built-in_plugins: a collection of plugins provided with Veradco. Some plugins are simple example while some others can be useful in a real Kubernetes cluster. Each plugin is in a subfolder and has a documentation in the README.md file
- Kustomize: some Kustomize overlays to install Veradco in a Kubernetes cluster. You can create your own Kustomize overlay from one of the provided one to deploy Veradco in your cluster in a way suitable to your environment.

## Docker images

Veradco is made of 2 Docker images:
- An init container that is responsible to provide to Veradco container its binary and the required plugins. In its nominal version, it builds on the fly the required built-in plugins and the external ones. It then shares via a shared volume the veradcod binary and the built plugins.
- A lightweight container that runs veradcod (the Veradco server and its monitoring server).

## Veradco endpoints

Veradco is provided with 11 endpoints:
- /healthz: serves as kubelet's livenessProbe hook to monitor health of the Veradco server
- /validate/pods: a validating webhook endpoint specialized for pods. When used, a core v1 Pod API object is directly passed to the scoped plugins.
- /mutate/pods: a mutating webhook endpoint specialized for pods. When used, a core v1 Pod API object is directly passed to the scoped plugins.
- /validate/deployments: a validating webhook endpoint specialized for deployments.
- /mutate/deployments: a mutating webhook endpoint specialized for pods.
- /validate/daemonsets: a validating webhook endpoint specialized for daemonsets.
- /mutate/daemonsets: a mutating webhook endpoint specialized for daemonsets.
- /validate/statefulsets: a validating webhook endpoint specialized for statefulsets.
- /mutate/statefulsets: a mutating webhook endpoint specialized for statefulsets.
- /validate/others: a validating webhook endpoint non-specialized that provides to the scoped plugins a generic meta v1 Kubernetes API object "PartialObjectMetadata". The "PartialObjectMetadata" object is intended to be used by the plugin to determine if it is in its scope and then it is up to the plugin to unmarshal the specialized object from the provided admission request.
- /mutate/others: a mutating webhook endpoint non-specialized that provides to the scoped plugins a generic meta v1 Kubernetes API object "PartialObjectMetadata". The "PartialObjectMetadata" object is intended to be used by the plugin to determine if it is in its scope and then it is up to the plugin to unmarshal the specialized object from the provided admission request.

Note: the "other" endpoints can seem complicated to use. Refer to examples to speed up understanding.

## Plugin

This part descibes what is a plugin. Each plugins is splitted in 2 parts:
- A plugin shall implement an interface in order Veradco is able to hanle it.
- To use a plugin, it shall be declared in the configuration. A plugin configuration defines its scope.

### Interface to implement

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

The configuation embeds some basic filtering fields so that your plugin is called or not. It's up to you in the code of your plugin to do the rest. It is as flexible as a Go programming code using Kubernetes API is. Filtering fields most of the time work as regular expressions.

```
plugins:
- name: "extplug1"
  path: "/app/external_plugins/extplug1.so"
  code: |
    cGFja2FnZSBtYWluCgppbXBvcnQgKAoJLy8gbWV0YSAiazhzLmlvL2FwaW1hY2hpbmVyeS9wa2cv
    ...ZXJhZGNvUGx1Z2luIFBsdWcx
  kinds: "^Pod$"
  operations: "CREATE"
  namespaces: ".*"
  labels:
  - key: owner
    value: me
  annotations:
  - key: owner
    value: me
  dryRun: false
  configuration: |
    This plugin does not have configuration
    That's like that!
  scope: "Validating"
```

### name

To identify the plugin.

### path

The path of the plugin file (.so). If the plugin needs to be built (refer to code field below), it is MANDATORY that is path is /app/external_plugins. It will be built by the Init Container at webhook startup.  
For a built-in plugin the path is /app/plugins.  
For a plugin you built yourself the path is as you want. We advise you to build your plugin by using the init container docker image to avoid infrequent issues with Golang plugins compatibility.

### code

The code field contains the code of the plugin converted in base 64. If the plugin has to be built, it shall be packaged in a single file. The base 64 block is decoded in a single Go file. The plugin is compiled with this code only if the provided path does not point to an existing file.

### kind

A regular expression to define the kinds on which the plugin is called.  
Example: "^Pod$"

### operations

A regular expression to define the operations on which the plugin is called.  
Example: "CREATE|UPDATE"  
It's up to the plugin to manage supported operations in its code.

### namespaces

A regular expression to define the namespaces in the scope of the plugin.  
Example: "kube-system|default"

### labels

Filter only on resources having some labels.  
value is a regular expressions.

### annotations

Filter only on resources having some annotations.  
value is a regular expressions.

### dryRun

This boolean parameter is self explanatory and managed upstream at veradco level. If the plugin does things outside of the scope of the webhooks, it shall be managed in its code. 

### configuration

The configuration of the plugin. Passed to the plugin via the Init function of the plugin. The format of the configuration is up to the plugin.

### scope

A regular expression that defines the scope of the plugin.  
There are 2 scopes: Validating and Mutating.  
"Validating|Mutating" is suitable for both scopes.

## How to develop a plugin

Developing a plugin is quiet simple. The most direct path is to take inspiration from the examples provided.

A plugin is generally made of a single file called plug.go. For external plugins (passed as a base 64 block), it is obviously mandatory that it is made of a single file.

### Build the plugin

Build the plugin is useful only to check that the plugin builds. It is up to Veradco init container to build the plugins.

To check that your plugin builds, you can proceed as follow:
```
go mod init github.com/smart-duck/veradco/my-plugin
# Optionally if you pulled the veradco code, you can point to it.
go mod edit -replace github.com/smart-duck/veradco=[Veradco repository]/veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```