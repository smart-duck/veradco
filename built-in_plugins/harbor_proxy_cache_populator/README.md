# HarborProxyCachePopulator

## Overview

This plugin allows to implicitly pull docker images in harbor proxy caches. To do it, it implements the v2 registry specification to mimic docker pull.

Instead of proceeding to a docker pull (using Golang Docker API that needs Docker daemon), the plugin implements the v2 registry specification for various reasons:
- Kubernetes is deprecating support for Docker as a container runtime starting with Kubernetes version 1.20. Kubernetes is about to use directly the underlying engine containerd directly (Docker is just a "middle-man" between Kubernetes and containerd).
- Using Docker pull requires to use the Docker daemon of the node host which implies to run the Veradco pods in a privileged mode and the Docker container as root. One of the best practices while running Docker Container is to run processes with a non-root user. Moreover, Docker daemon is not necessarily installed in nodes with Kubernetes versions from 1.20.

This plugin is suitable to be used in the case you have a master container registry (typically Harbor) to centralize images and slave container registries (typically ECR) used as Application registries.

The plugin takes advantage of the registries proxy caches mechanisms (as implemented in Harbor).

## How it works

```
+-----+    +-----+       +---------+                +---------+             +-------------+ +-----+
| CD  |    | K8s |       | Veradco |                | Harbor  |             | AnyRegistry | | ECR |
+-----+    +-----+       +---------+                +---------+             +-------------+ +-----+
   |          |               |                          |                         |           |
   | Apply    |               |                          |                         |           |
   |--------->|               |                          |                         |           |
   |          |               |                          |                         |           |
   |          | Validate      |                          |                         |           |
   |          |-------------->|                          |                         |           |
   |          |               |                          |                         |           |
   |          |               | Parse ECR image URL      |                         |           |
   |          |               |--------------------      |                         |           |
   |          |               |                   |      |                         |           |
   |          |               |<-------------------      |                         |           |
   |          |               |                          |                         |           |
   |          |               | Get OCI manifest         |                         |           |
   |          |               |------------------------->|                         |           |
   |          |               |                          |                         |           |
   |          |               |                          | Get OCI manifest?       |           |
   |          |               |                          |------------------------>|           |
   |          |               |                          |                         |           |
   |          |               | Get image layers         |                         |           |
   |          |               |------------------------->|                         |           |
   |          |               |                          |                         |           |
   |          |               |                          | Get image layers?       |           |
   |          |               |                          |------------------------>|           |
   |          |               |                          |                         |           |
   |          |               |                          | Replicate               |           |
   |          |               |                          |------------------------------------>|
   |          |               |                          |                         |           |
```

Description of the operation:
1. A CD pipeline applies a specification (Pod, Deployment...) towards the Kubernetes cluster.
2. An admission review is sent to Veradco by the API server in order it validates the specification.
3. The plugin HarborProxyCachePopulator is executed by Veradco and accept the specificationimmediatly (not is purpose). Then it executes asynchronously. It parses the ECR image URL to retrieve needed values (for each container or init container).
4. The plugin HarborProxyCachePopulator requests the "OCI manifest" towards the Harbor proxy cache.
5. If Harbor does not have this manifest, then it requests it to the underlying registry (of the proxy cache). 
6. The plugin HarborProxyCachePopulator selects the right Architecture from the "OCI manifest" and then requests the Image layers manifest towards Harbor proxy cache.
7. If Harbor does not have this manifest, then it retrieves image layers from the underlying registry.
8. Later, Harbor replicates the container image to the slave registries (ECR).

Note: to avoid managing 2 times the same image, Veradco manages a list of already met images.

## Plugin configuration

Here is an example:
```
maxNumberOfParallelJobs: 2
proxyCaches:
- regexURL: "^.*amazonaws.com/(proxy_[^:/]+)/([^:]+):(.+$)"
  replacementOCI: "https://harbor.registry.mine.io/v2/$1/$2/manifests/$3"
  replacementArch: "https://harbor.registry.mine.io/v2/$1/$2/manifests/ARCHDIGEST"
  targetArch: "amd64"
  targetOS: "linux"
```

### maxNumberOfParallelJobs parameter

This parameter defines the number of images to manage in parallel. Concretely you can put a lot as what the plugin does is not really consuming.

### proxyCaches parameter

A list of proxy caches to handle:
- regexURL: a regex that allows to identify and parse the slave registry URL.
- replacementOCI: used to generate the URL to the "OCI manifest" in the master registry (Harbor) from the regexURL (ReplaceAllString of the Golang regexp package).
- replacementArch: used to generate the URL to the Image layers manifest in the master registry (Harbor) from the regexURL (ReplaceAllString of the Golang regexp package). If this parameter is set to empty string "", then only the first step is proceeded. It shall finish by ARCHDIGEST that is replaced by the digest of the target architecture image.
- targetArch: the target architecture.
- targetOS: the target OS.

## Kubernetes secret to provision

To access your master registry (Harbor), you need to create a User/Password secret and define environment varariable using it in your Veradco deployment like that:
```
apiVersion: v1
kind: Secret
metadata:
  name: harbor
  namespace: veradco
type: Opaque
stringData:
    hUSER: "robot_pull"
    hPW: "cEsTunVRaiSeCReT"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veradco
  namespace: veradco
spec:
  template:
    metadata:
      labels:
        app: veradco
    spec:
      ...
      containers:
      - name: veradco-server
        image: localhost:5001/smartduck/veradco:0.1
        env:
        - name: hUSER
          valueFrom:
            secretKeyRef:
              name: harbor
              key: hUser
        - name: hPW
          valueFrom:
            secretKeyRef:
              name: harbor
              key: hPW
```

## DEBUG mode

Debug mode can be use to test deployment of this plugin in your cluster. You just have to define the HARBORPCP_DEBUG environment variable as follow (the value does not matter):
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veradco
  namespace: veradco
spec:
  template:
    metadata:
      labels:
        app: veradco
    spec:
      initContainers:
      - name: veradco-plugins-init
        image: localhost:5001/smartduck/veradco-golang-builder:0.1
      containers:
      - name: veradco-server
        image: localhost:5001/smartduck/veradco:0.1
        env:
        - name: "HARBORPCP_DEBUG"
          value: "debug proxy cache populator"
        - name: hUSER
          value: "HARBORuser"
        - name: hPW
          value: "HARBORpw"
```

Note: In debug mode, it does not matter not to define environment variables used (or define them with dummy values) to authenticate to the master registry (typically Harbor).

## Check that the plugin builds

As explained in the Veradco documentation, you can proceed as follow:
```
go mod init github.com/smart-duck/veradco/harbor_proxy_cache_populator
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```

## Example of use

In the following example, a pod containing 4 containers is applied. The plugin configuration accepts only to manage 2 images in parallel. Plugin is in DEBUG mode: it simulates image pull with a random time between 5 and 15 seconds.
```
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.134880       1 conf.go:255] >> Number of plugins selected: 1
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135318       1 conf.go:289] >> Plugin HarborProxyCachePopulator execution summary: Execute plugin HarborProxyCachePopulator
veradco-d959655c6-vmr8t veradco-server Check that image alpine:3.16 is in the proxy cache
veradco-d959655c6-vmr8t veradco-server Check that image alpine:3.15 is in the proxy cache
veradco-d959655c6-vmr8t veradco-server Check that image alpine:3.14 is in the proxy cache
veradco-d959655c6-vmr8t veradco-server Check that image alpine:3.13 is in the proxy cache
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135376       1 conf.go:291] >> Plugin HarborProxyCachePopulator is in dry run mode. Nothing to do!
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135742       1 plug.go:266] >>>>>> simulate pullImageFromProxyCache to pull alpine:3.13 - Wait 12 seconds
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135820       1 plug.go:266] >>>>>> simulate pullImageFromProxyCache to pull alpine:3.16 - Wait 5 seconds
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135845       1 plug.go:169] Wait for a slot is freed to pull alpine:3.15
veradco-d959655c6-vmr8t veradco-server I0926 12:51:32.135867       1 plug.go:169] Wait for a slot is freed to pull alpine:3.14
veradco-d959655c6-vmr8t veradco-server I0926 12:51:37.141212       1 plug.go:169] Wait for a slot is freed to pull alpine:3.14
veradco-d959655c6-vmr8t veradco-server I0926 12:51:37.141240       1 plug.go:169] Wait for a slot is freed to pull alpine:3.15
veradco-d959655c6-vmr8t veradco-server I0926 12:51:42.142272       1 plug.go:169] Wait for a slot is freed to pull alpine:3.14
veradco-d959655c6-vmr8t veradco-server I0926 12:51:42.142273       1 plug.go:266] >>>>>> simulate pullImageFromProxyCache to pull alpine:3.15 - Wait 5 seconds
veradco-d959655c6-vmr8t veradco-server I0926 12:51:47.146286       1 plug.go:266] >>>>>> simulate pullImageFromProxyCache to pull alpine:3.14 - Wait 6 seconds
```
