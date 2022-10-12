#!/bin/bash

CURRDIR=$(dirname $(readlink -f $0))

BUILD_ARG=ALL

BUILD_VERSION="v0.1.0"

[[ "$1" != "" ]] && BUILD_ARG=$1

[[ "$2" != "" ]] && BUILD_VERSION=$2

cd $CURRDIR/golang_builder
docker build --build-arg BUILD=$BUILD_ARG -t smartduck/veradco-golang-builder:$BUILD_VERSION -f ./Dockerfile.golang_builder $CURRDIR/..

# Put it in local registry
# sudo veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-golang-builder:$BUILD_VERSION
