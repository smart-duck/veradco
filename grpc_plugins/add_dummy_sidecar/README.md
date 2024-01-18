# AddSidecar plugin

## Overview

This plugin is an example one. It is a mutating plugin that inject a side car and an annotation to a pod.

## Package

```
go build -o docker/adddummysidecar main.go

cd docker;sudo docker build -t smartduck/adddummysidecar-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/adddummysidecar-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_adddummysidecar.yaml

~/go/bin/stern -n default adddummysidecarplugin &
```