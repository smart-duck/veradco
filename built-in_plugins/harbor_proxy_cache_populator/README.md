# Overview of the plugin HarborProxyCachePopulator

This plugin allows to implicitly pull docker images in harbor proxy caches. To do it, it implements the v2 registry specification to mimic docker pull.

Instead of proceeding to a docker pull (using Golang Docker API), the plugin implements the v2 registry specification for various reasons:
- Kubernetes is deprecating support for Docker as a container runtime starting with Kubernetes version 1.20. Kubernetes is about to use directly the underlying engine containerd directly (Docker is just a "middle-man" between Kubernetes and containerd).
- Using Docker pull requires to use the Docker daemon of the node host which implies to run the Veradco pods in a privileged mode and the Docker container as root. One of the best practices while running Docker Container is to run processes with a non-root user. Moreover, Docker daemon is not necessarily installed with Kubernetes versions from 1.20.

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
6. The plugin HarborProxyCachePopulator selects the right Architecture from the OCI manifest and then requests the Image layers manifest towards Harbor proxy cache.
7. If Harbor does not have this manifest, then it retrieves image layers from the underlying registry.
8. Later, Harbor replicates the container image to the slave registries (ECR).

## Check that the plugin builds

As explained in the Veradco documentation, you can proceed as follow:
```
go mod init github.com/smart-duck/veradco/harbor_proxy_cache_populator
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```
