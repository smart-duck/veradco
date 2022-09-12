#!/bin/bash

BUILD_ARG=ALL

[[ "$1" != "" ]] && BUILD_ARG=$1

cd /home/lobuntu/go/src/veradco
sudo docker build --build-arg BUILD=$BUILD_ARG -t smartduck/veradco-golang-builder:0.1 -f ./Dockerfile.golang_builder .
sudo veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-golang-builder:0.1
cd veradco/demo
sudo docker build -t smartduck/veradco:0.1 -f ../Dockerfile.golang_builder ..
sudo local_registry/push_local_image_to_local_registry.sh smartduck/veradco:0.1