# EnforceLabels plugin

## Overview

It is a validating plugin that enforces labels and annotations thanks to a configuration as follow:
```
annotations: 
  owner: ^to.+
labels: 
  nodegp: ^ng-.+
```

## Build

go build -o docker/enforcelabels main.go

cd docker;sudo docker build -t smartduck/enforcelabels-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/enforcelabels-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_enforcelabels.yaml

~/go/bin/stern -n default enforcelabelsplugin &
