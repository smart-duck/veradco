#!/bin/bash

CURRDIR=$(dirname $(readlink -f $0))

BUILD_VERSION="v0.1.0"

[[ "$1" != "" ]] && BUILD_VERSION=$1

cd $CURRDIR/run
sudo docker build -t smartduck/veradco:$BUILD_VERSION -f ./Dockerfile.golang_builder .

# Put it in local registry
sudo local_registry/push_local_image_to_local_registry.sh smartduck/veradco:$BUILD_VERSION
