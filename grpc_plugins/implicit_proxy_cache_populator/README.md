# ImplicitProxyCachePopulator

## Overview

This plugin allows to implicitly replicate docker images from public registries to the private ECR. To do it, it uses the regclient module that implements docker v2 specification.

## How it works

```
+-----+    +-----+       +---------+    +---------+                +-----------------+ +-------------+ +-----------------+
| CD  |    | K8s |       | Veradco |    | Plugin  |                | PublicRegistry  | | PrivateECR  | | PrivateRegistry |
+-----+    +-----+       +---------+    +---------+                +-----------------+ +-------------+ +-----------------+
   |          |               |              |                              |                 |                 |
   | Apply    |               |              |                              |                 |                 |
   |--------->|               |              |                              |                 |                 |
   |          |               |              |                              |                 |                 |
   |          | Validate      |              |                              |                 |                 |
   |          |-------------->|              |                              |                 |                 |
   |          |               |              |                              |                 |                 |
   |          |               | Execute      |                              |                 |                 |
   |          |               |------------->|                              |                 |                 |
   |          |               |              |                              |                 |                 |
   |          |               |              | Check image pattern          |                 |                 |
   |          |               |              |--------------------          |                 |                 |
   |          |               |              |                   |          |                 |                 |
   |          |               |              |<-------------------          |                 |                 |
   |          |               |              |                              |                 |                 |
   |          |               |              | Check image exists           |                 |                 |
   |          |               |              |----------------------------------------------->|                 |
   |          |               |              |                              |                 |                 |
   |          |               |              | Pull layers                  |                 |                 |
   |          |               |              |----------------------------->|                 |                 |
   |          |               |              |                              |                 |                 |
   |          |               |              | Push layers                  |                 |                 |
   |          |               |              |----------------------------------------------------------------->|
   |          |               |              |                              |                 |                 |
```

Note: to avoid managing 2 times the same image, Veradco manages a list of already met images.

## Plugin integration and configuration

Here is an example:
```yaml
plugins:
- name: "implicit_proxy_cache_populator"
  endpoints: "/validate/pods"
  path: "/app/plugins/built-in-implicit_proxy_cache_populator.so"
  kinds: "(?i)^Pod$"
  operations: "CREATE|UPDATE"
  namespaces: ".*"
  dryRun: false
  configuration: |
    maxNumberOfParallelJobs: 2
    proxyCaches:
    - patternEcrProxyCache: "(^[0-9]+\\.dkr\\.ecr\\.[^.]+\\.amazonaws.com)/(proxy_)(docker\\.io|docker\\.cloudsmith\\.io|docker\\.elastic\\.io|gcr\\.io|ghcr\\.io|k8s\\.gcr\\.io|public\\.ecr\\.aws|quay\\.io|registry\\.k8s\\.io|registry\\.opensource\\.zalan\\.do|us-docker\\.pkg\\.dev|xpkg\\.upbound\\.io)/([^:]+):?(.*)$"
      platform: "linux/amd64"
      region: "eu-central-1"
  scope: "Validating"
```

## Check that the plugin builds

As explained in the Veradco documentation, you can proceed as follow:
```sh
go mod init github.com/smart-duck/veradco/implicit_proxy_cache_populator
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```

Note: as it is a built-in plugin, you don't need to check that it builds.

## Package

```
go build -o docker/implicit_proxy_cache_populator main.go

cd docker;sudo docker build -t smartduck/implicit_proxy_cache_populator-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/implicit_proxy_cache_populator-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_implicit_proxy_cache_populator.yaml

~/go/bin/stern -n default implicit_proxy_cache_populator &
```
