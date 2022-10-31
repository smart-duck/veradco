# veradco

## Overview

Veradco a.k.a. Versatile Admission Controller is an admission controller that is expandable via a plugin system. It handles Mutating and Validating webhooks that you can extend by developing your own plugins or by using some third-party ones or the ones that are built-in.

[Kubernetes admission controllers](https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers)

> In a nutshell, Kubernetes admission controllers are plugins that govern and enforce how the cluster is used. They can 
> be thought of as a gatekeeper that intercept (authenticated) API requests and may change the request object or deny the request altogether.
> The admission control process has two phases: the mutating phase is executed first, followed by the validating phase.

Kubernetes admission Controller Phases:

![](https://d33wubrfki0l68.cloudfront.net/af21ecd38ec67b3d81c1b762221b4ac777fcf02d/7c60e/images/blog/2019-03-21-a-guide-to-kubernetes-admission-controllers/admission-controller-phases.png)

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
        |                                | Excecute scoped plugins 
        |                                |------------------------------
        |                                |                             | 
        |                                |<----------------------------- 
        |                                |                            
        |   Plugins responses aggregated |                            
        |<-------------------------------|                            
        |                                |
```

Description of the operation:
1. An admission review is sent to Veradco by the API server in order it validates or mutates a Kubernetes specification.
2. Veradco execute the plugins that are in the scope of the admission review. It sums up the pluings responses.
3. Veradco responds to the API server on behalf of executed plugins.

More fully, this sequence diagram shows that Veradco acts at mutate and at validate phase:
```
+---------------+                     +---------+                   
| KubernetesAPI |                     | Veradco |                   
+---------------+                     +---------+                   
        |                                  |                        
        | Mutate                           |                        
        |--------------------------------->|                        
        |                                  |                        
        |                                  | Execute scoped plugins 
        |                                  |----------------------- 
        |                                  |                      | 
        |                                  |<---------------------- 
        |                                  |                        
        |       Aggregated plugins patches |                        
        |<---------------------------------|                        
        |                                  |                        
        | Validate                         |                        
        |--------------------------------->|                        
        |                                  |                        
        |                                  | Execute scoped plugins 
        |                                  |----------------------- 
        |                                  |                      | 
        |                                  |<---------------------- 
        |                                  |                        
        |      Summed up plugins allowance |                        
        |<---------------------------------|                        
        |                                  |                        
```

Note: If a plugin rejects the request, then its response is sent to the Kubernetes API.

## Repository structure

The repository is made of 3 main folders:
- veradco: the Golang code of the Veradco Admission Controller.
- built-in_plugins: a collection of plugins provided with Veradco. Some plugins are simple example while some others can be useful in a real Kubernetes cluster. Each plugin is in a subfolder and has a documentation in the README.md file.
- Kustomize: some Kustomize overlays to install Veradco in a Kubernetes cluster. You can create your own Kustomize overlay from one of the provided one to deploy Veradco in your cluster in a way suitable to your environment.
- docker: Several Dockerfile files and linked generation scripts. This collection of Dockerfiles covers the following use cases:
  - Build binaries including plugins
  - Create a Veradco standalone image (does not build on the fly)
  - Create a build on the fly image for the init container and a run image for the container

## Install Veradco

### Nominal installation

The nominal installation install Veradco container together with its init container. It is then able to build plugins (and even veradcod) on the fly.

To install Veradco, use the Kustomize configuration provided or create your own.

Prior to proceed to the Kustomize, you should have initialized the PKI linked with the service using the provided script generate-pki.sh . The script generates a CA and the certificate/key of the Veradco service.

By example, if you want to install the default installation, run:
```sh
kubectl apply -k kustomize/base
```

To create the deployment specification with kustomize command and deploy it with kubectl command:
```sh
kustomize build ../../kustomize/base | kubectl apply -f -
```

Note: It is up to you to create the webhooks via Kustomize with your own overlay or separately.

### Custom installation

It is obviously possible to create your own installation. Typically you can create an image containing all the binaries (veradcod and plugins) in your CI and use it in your CD. In this case you don't need to have the init container (heavy) but you lost the ability to build on the fly. To build the binaries you can use the init container image which is suitable for both CD and CI use: refer to the README of the folder useful/build_plugins. 

## Docker images

Veradco is made of 2 Docker images:
- An init container that is responsible to provide to Veradco container its binary and the required plugins. In its nominal version, it builds on the fly the required built-in plugins and the external ones. It then shares via a shared volume the veradcod binary and the built plugins. This image can also be used as a CI one to build the binaries.
- A lightweight container that runs veradcod (the Veradco server and its monitoring server). Nominally, all binaries including plugins and veradcod are provided by the init container via a shared volume.

## Veradco endpoints

### List of provided endpoints

Veradco is provided with 11 endpoints:
- /healthz: serves as kubelet's livenessProbe hook to monitor health of the Veradco server
- /validate/pods: a validating webhook endpoint specialized for pods. When used, a core v1 Pod API object is directly passed to the scoped plugins.
- /mutate/pods: a mutating webhook endpoint specialized for pods. When used, a core v1 Pod API object is directly passed to the scoped plugins.
- /validate/deployments: a validating webhook endpoint specialized for deployments. When used, a specialized API object is directly passed to the scoped plugins.
- /mutate/deployments: a mutating webhook endpoint specialized for pods. When used, a specialized API object is directly passed to the scoped plugins.
- /validate/daemonsets: a validating webhook endpoint specialized for daemonsets. When used, a specialized API object is directly passed to the scoped plugins.
- /mutate/daemonsets: a mutating webhook endpoint specialized for daemonsets. When used, a specialized API object is directly passed to the scoped plugins.
- /validate/statefulsets: a validating webhook endpoint specialized for statefulsets. When used, a specialized API object is directly passed to the scoped plugins.
- /mutate/statefulsets: a mutating webhook endpoint specialized for statefulsets. When used, a specialized API object is directly passed to the scoped plugins.
- /validate/others: a validating webhook endpoint non-specialized that provides to the scoped plugins a generic meta v1 Kubernetes API object "PartialObjectMetadata". The "PartialObjectMetadata" object is intended to be used by the plugin to determine if it is in its scope and then it is up to the plugin to unmarshal the specialized object from the provided admission request.
- /mutate/others: a mutating webhook endpoint non-specialized that provides to the scoped plugins a generic meta v1 Kubernetes API object "PartialObjectMetadata". The "PartialObjectMetadata" object is intended to be used by the plugin to determine if it is in its scope and then it is up to the plugin to unmarshal the specialized object from the provided admission request.

Note: the "other" endpoints can seem complicated to use. Refer to examples to speed up understanding.

### How to define Veradco Webhooks

You have to deploy the webhooks in accordance with the endpoints you want to use.

Basically, if you want to use the /validate/pods endpoint, then you have to define a webhook with a rule filtering pods resources as follow:
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: veradco-pod-validation
webhooks:
  - name: veradco-pod-validation.default.svc
    sideEffects: None
    admissionReviewVersions: ["v1"]
    clientConfig:
      service:
        name: veradco
        namespace: veradco
        path: "/validate/pods"
      caBundle: "$(CA_BUNDLE)"
    namespaceSelector:
      matchExpressions:
        - key: kubernetes.io/metadata.name
          operator: NotIn
          values: ["veradco"]
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
        scope: "Namespaced"
    failurePolicy: Ignore
```

Notes:
- the namespaceSelector used allows not to apply this webhook to the veradco namespace resources. The Kubernetes API server sets this label on all namespaces to the name of the namespace.
- As it uses the /validate/pods endpoint, it is a ValidatingWebhookConfiguration resource.

For the other specialized endpoints you have to do in the same way.

Others endpoints (/validate/others, /mutate/others) are more generic and the relative webhooks have to be defined with a rule filtering resources other than the ones handled by the specialized endpoints.

Here is an example of a validating webhook that uses the /validate/others endpoint:
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: veradco-other-validation
webhooks:
  - name: veradco-other-validation.default.svc
    sideEffects: None
    admissionReviewVersions: ["v1"]
    clientConfig:
      service:
        name: veradco
        namespace: veradco
        path: "/validate/others"
      caBundle: "${CA_BUNDLE}"
    rules:
      - operations: ["CREATE", "DELETE", "UPDATE", "CONNECT"]
        apiGroups: ["*"]
        apiVersions: ["*"]
        resources: ["secrets", "namespaces"]
    failurePolicy: Ignore
```

## Plugin

This part descibes what is a plugin. Each plugin is splitted in 2 parts:
- The plugin code that shall implement an interface in order Veradco is able to handle it.
- To use a plugin, it shall be declared in the configuration. A plugin configuration defines its scope, its configuration, if it shall be run in dry mode and optionally the code to build it for the external ones.

### Plugins and endpoints

There is a bijection between plugins and the Golang Kubernetes object that is passed to the Execute function of a plugin. So, depending on the code of the plugin, it can be executed several times. To avoid that, it is suitable to define the applicable endpoints in the plugin configuration (parameter endpoints). By example, the Generic plugin is designed to work with all endpoints. If you want it to work with the endpoint "others" which is its native one, you need to define the endpoints parameter to "/validate/others" in the plugin configuration.

Note: in the Veradco configuration, a plugin can be declared several times with different configurations.

### Interface to implement

A plugin is a piece of Golang code that implements the following interface:
```go
type VeradcoPlugin interface {
  Init(configFile string) error
  Execute(kobj runtime.Object, operation string, dryRun bool, r *admission.AdmissionRequest) (*admissioncontroller.Result, error)
  Summary() string
}
```

## Configuration

The configuration defines the plugins to use and their configuration. Here is an example of a ConfigMap.

The configuation embeds some filtering fields so that your plugin is called or not. It's up to you in the code of your plugin to do the rest. It is as flexible as a Go programming code using Kubernetes API is. Filtering fields most of the time work as regular expressions.

```yaml
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
  endpoints: "/validate/pods"
```

### Field name

To identify the plugin.

### Field path

The path of the plugin file (.so). If the plugin needs to be built (refer to code field below), it is MANDATORY that is path is /app/external_plugins. It will be built by the Init Container at webhook startup.  
For a built-in plugin the path is /app/plugins.  
For a plugin you built yourself the path is as you want. We advise you to build your plugin by using the init container docker image to avoid infrequent issues with Golang plugins compatibility.

#### Built-in plugins

Here is a list of built-in plugins. It could be out-of-date. To have the up-to-date list of built-in plugins, refer directly to the code (built-in_plugins folder). Some plugins are simple example while some others can be useful in a real Kubernetes cluster. Each plugin is in a subfolder and has a documentation in the README.md file.

- built-in-add_dummy_sidecar
- built-in-basic
- built-in-enforce_labels
- built-in-forbid_tag
- built-in-generic
- built-in-harbor_proxy_cache_populator
- built-in-plug1
- built-in-registry_cache

### code

This field is optional.

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

This field is optional.

Filter only on resources having some labels.  
value is a regular expressions.

### annotations

This field is optional.

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

### endpoints

This field is optional. If not defined, all activated endpoints are likely to be used for the plugin.

A regular expression that defines the endpoints that execute the plugin.  
"/validate/pods|/validate/others" is suitable for the 2 endpoints /validate/pods et /validate/others. In its code the plugin can restrict suitable endpoints. It can be designed for a particular endpoint: by example a plugin checking the validity of a container image can restrict its usage to the endpoint /validate/pods.

## How to develop a plugin

Developing a plugin is quiet simple. The most direct path is to take inspiration from the examples provided.

A plugin is generally made of a single file called plug.go. For external plugins (passed as a base 64 block), it is obviously mandatory that it is made of a single file.

### Build the plugin

Build the plugin is useful only to check that the plugin builds. It is up to Veradco init container to build the plugins.

To check that your plugin builds, you can proceed as follow:
```sh
go mod init github.com/smart-duck/veradco/my-plugin
# Optionally if you pulled the veradco code, you can point to it.
go mod edit -replace github.com/smart-duck/veradco=[Veradco repository]/veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```

## Regular expressions handling

Regular expressions are handled by Verado thanks to the golang package regexp. But, Veradco introduces a special wild card that is used in the cases it is relevant:
- regular expression act as a reverse pattern if it is prefixed by (!\~). By example, "(!\~)(?i)test" matches that the value does not contain "test" whatever the case is.

## Issue handling

If you are facing Veradco issues, refer to the logs of its pods.

Sometime, Veradco can run in degraded mode. By example, in case it does not succeed to load all the plugins that are declared in the configuration:
```sh
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.266961       1 main.go:28] >>>>>> Starting veradco
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.267183       1 server.go:42] >>>> NewServer
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.267943       1 server.go:51] >> Configuration /conf/veradco.yaml successfully loaded
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.268201       1 conf.go:66] >>>> Loading plugins
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.268391       1 conf.go:68] >> Loading plugin  Generic
veradco-b69b8d445-krm2x veradco-server E0930 08:24:48.268631       1 conf.go:73] Unable to load plugin Generic: plugin.Open("/app/plugins/built-in-generic.so"): realpath failed
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.268893       1 conf.go:68] >> Loading plugin  HarborProxyCachePopulator
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.300927       1 conf.go:95] >> Init plugin HarborProxyCachePopulator
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.301115       1 conf.go:114] >> 1 plugins loaded over 2
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.301138       1 server.go:63] >> 1 plugins successfully loaded
veradco-b69b8d445-krm2x veradco-server I0930 08:24:48.301466       1 main.go:47] >> Server running on port: 8443
```

## Example of Veradco logs

The following logs show the starting of Veradco with only one external plugin in its configuration (HarborProxyCachePopulator). We can see the building of the plugin by the init container and then the starting of Veradco server:
```
+ veradco-d959655c6-crwrd › veradco-plugins-init
veradco-d959655c6-crwrd veradco-plugins-init BUILD INTERNAL plugins
veradco-d959655c6-crwrd veradco-plugins-init /go/src/built-in_plugins
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-add_dummy_sidecar
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-basic
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-enforce_labels
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-forbid_tag
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-generic
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-harbor_proxy_cache_populator
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-plug1
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-registry_cache
veradco-d959655c6-crwrd veradco-plugins-init NO NEED TO BUILD plugin built-in-user_plugin
veradco-d959655c6-crwrd veradco-plugins-init List of built plugins:
veradco-d959655c6-crwrd veradco-plugins-init BUILD EXTERNAL plugins
veradco-d959655c6-crwrd veradco-plugins-init Copy veradcod to /app, also plugins folder
veradco-d959655c6-crwrd veradco-plugins-init File /app/external_plugins/harbor_proxy_cache_populator.so does not exist. Build plugin...
veradco-d959655c6-crwrd veradco-plugins-init go: creating new go.mod: module github.com/smart-duck/veradco/20220926122555
veradco-d959655c6-crwrd veradco-plugins-init go: to add module requirements and sums:
veradco-d959655c6-crwrd veradco-plugins-init 	go mod tidy
veradco-d959655c6-crwrd veradco-plugins-init go: finding module for package gopkg.in/yaml.v3
veradco-d959655c6-crwrd veradco-plugins-init go: finding module for package k8s.io/api/admission/v1
veradco-d959655c6-crwrd veradco-plugins-init go: finding module for package k8s.io/api/core/v1
veradco-d959655c6-crwrd veradco-plugins-init go: finding module for package k8s.io/apimachinery/pkg/runtime
veradco-d959655c6-crwrd veradco-plugins-init go: finding module for package k8s.io/klog/v2
veradco-d959655c6-crwrd veradco-plugins-init go: found github.com/smart-duck/veradco in github.com/smart-duck/veradco v0.0.0-00010101000000-000000000000
veradco-d959655c6-crwrd veradco-plugins-init go: found github.com/smart-duck/veradco/kres in github.com/smart-duck/veradco v0.0.0-00010101000000-000000000000
veradco-d959655c6-crwrd veradco-plugins-init go: found gopkg.in/yaml.v3 in gopkg.in/yaml.v3 v3.0.1
veradco-d959655c6-crwrd veradco-plugins-init go: found k8s.io/api/admission/v1 in k8s.io/api v0.25.2
veradco-d959655c6-crwrd veradco-plugins-init go: found k8s.io/api/core/v1 in k8s.io/api v0.25.2
veradco-d959655c6-crwrd veradco-plugins-init go: found k8s.io/apimachinery/pkg/runtime in k8s.io/apimachinery v0.25.2
veradco-d959655c6-crwrd veradco-plugins-init go: found k8s.io/klog/v2 in k8s.io/klog/v2 v2.80.1
veradco-d959655c6-crwrd veradco-plugins-init List of external plugins:
veradco-d959655c6-crwrd veradco-plugins-init harbor_proxy_cache_populator.so
veradco-d959655c6-crwrd veradco-plugins-init app content:
veradco-d959655c6-crwrd veradco-plugins-init /app:
veradco-d959655c6-crwrd veradco-plugins-init total 31804
veradco-d959655c6-crwrd veradco-plugins-init drwxr-xr-x    2 root     root          4096 Sep 26 12:26 external_plugins
veradco-d959655c6-crwrd veradco-plugins-init drwxr-xr-x    2 root     root          4096 Sep 26 12:25 plugins
veradco-d959655c6-crwrd veradco-plugins-init -rwxr-xr-x    1 root     root      32555368 Sep 26 12:25 veradcod
veradco-d959655c6-crwrd veradco-plugins-init 
veradco-d959655c6-crwrd veradco-plugins-init /app/external_plugins:
veradco-d959655c6-crwrd veradco-plugins-init total 27420
veradco-d959655c6-crwrd veradco-plugins-init -rw-r--r--    1 root     root      28075320 Sep 26 12:26 harbor_proxy_cache_populator.so
veradco-d959655c6-crwrd veradco-plugins-init 
veradco-d959655c6-crwrd veradco-plugins-init /app/plugins:
veradco-d959655c6-crwrd veradco-plugins-init total 0
- veradco-d959655c6-crwrd › veradco-plugins-init
+ veradco-d959655c6-crwrd › veradco-server
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.677444       1 main.go:28] >>>>>> Starting veradco
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.677724       1 server.go:29] >>>> NewServer
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.678412       1 server.go:38] >> Configuration /conf/veradco.yaml successfully loaded
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.678435       1 conf.go:65] >>>> Loading plugins
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.678447       1 conf.go:67] >> Loading plugin  HarborProxyCachePopulator
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.691763       1 conf.go:94] >> Init plugin HarborProxyCachePopulator
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.692165       1 conf.go:113] >> 1 plugins loaded over 1
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.692245       1 server.go:50] >> 1 plugins successfully loaded
veradco-d959655c6-crwrd veradco-server I0926 12:26:37.692432       1 main.go:47] >> Server running on port: 8443
```

