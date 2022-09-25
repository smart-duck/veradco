# Overview

This plugin allows to implicitly pull docker images in harbor proxy caches. To do it, it implements the docker registry specification to mimic docker pull.

This plugin is suitable to be used in the case you have a master container registry (typically Harbor) to centralize images and slave container registries (typically ECR) used as Application registries.

The plugin take advantage of the registries proxy caches mechanisms (as implemented in Harbor).

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
1. A CD pipeline apply a specification (Pod, Deployment...) towards the Kubernetes cluster.
2. An admission review is sent to Veradco by the API server in order it validates the specification.
3. The plugin HarborProxyCachePopulator (asynchronous) is executed by Veradco: it parses the ECR image URL to retrieve needed values.
4. The plugin HarborProxyCachePopulator requests the OCI manifest towards Harbor proxy cache.
5. If Harbor does not have this manifest, then it queries it to the underlying registry (of the proxy cache). 
6. The plugin HarborProxyCachePopulator select the right  Architecture from the OCI manifest and then requests the Image layers manifest towards Harbor proxy cache.
7. If Harbor does not have this manifest, then it retrieves image layers from the underlying registry.
8. Later, Harbor replicates the container image to the slave registry (ECR).

## Check that the plugin builds

```
go mod init github.com/smart-duck/veradco/harbor_proxy_cache_populator
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```

# TODO

- Debug mode with long sleep: DONE
- DEBUG env var: DONE
- Manage dryRun in plugin: DONE
- Queue with channels