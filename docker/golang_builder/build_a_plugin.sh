#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

set -e

# build_folder="/go/src/build/plugin"
build_folder="$1"

# Remove trailing slash if any
build_folder=$(echo $build_folder | sed -E 's:(^.+[^/])/?$:\1:g')

plugin_name=$(echo $build_folder | grep -o -E "[^/]+/?$")

cd $build_folder

if [ ! -f "go.mod" ] || [ "$PRESERVE_GO_MOD" != "true" ]; then
  set +e
  rm go.mod go.sum
  set -e
  go mod init "github.com/smart-duck/veradco/veradco/$plugin_name"
  go mod edit -replace github.com/smart-duck/veradco/veradco=../../veradco
  go mod tidy
fi

go build -buildmode=plugin -o "$build_folder/$plugin_name.so" plug.go

source $(dirname $(readlink -f $0))/end_any_script.source