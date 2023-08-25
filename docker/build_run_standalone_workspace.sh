#!/bin/bash

set -e

CURRDIR=$(dirname $(readlink -f $0))

STANDALONE_DIR="$CURRDIR/standalone"

RELEASE_DIR=$STANDALONE_DIR/release

BUILD_VERSION="v0.1.0"

[[ "$1" != "" ]] && BUILD_VERSION=$1

BUILD_VERSION_BUILDER=$BUILD_VERSION

[[ "$2" != "" ]] && BUILD_VERSION_BUILDER=$2

DOCKERFILE_STANDALONE="./Dockerfile.standalone"

[[ "$3" != "" ]] && DOCKERFILE_STANDALONE=$3

VERADCO_CONF="/conf/veradco_conf.yaml"
VERADCO_CONF_DIR=$STANDALONE_DIR
if [[ "$4" != "" ]]; then
  VERADCO_CONF="/conf/$(basename $4)"
  VERADCO_CONF_DIR=$(dirname $4)
fi

sudo rm -Rf $RELEASE_DIR || true

mkdir -p $RELEASE_DIR

# First of all build binaries
docker run --rm \
  --env TO_BUILD_FOLDER="/to_build" \
  --env TO_BUILD_CHMOD="1000:1000" \
  --env VERADCO_CONF=$VERADCO_CONF \
  -v $VERADCO_CONF_DIR:/conf \
  -v $RELEASE_DIR:/app \
  -v $CURRDIR/../veradco:/to_build/veradco \
  -v $CURRDIR/../built-in_plugins:/to_build/built-in_plugins \
  smartduck/veradco-golang-builder:$BUILD_VERSION_BUILDER /bin/sh -c "/veradco_scripts/build_workspace.sh"

cd $STANDALONE_DIR
docker build --no-cache -t smartduck/veradco-standalone:$BUILD_VERSION -f ./Dockerfile.standalone $RELEASE_DIR/


# Put it in local registry
# sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-standalone:$BUILD_VERSION
