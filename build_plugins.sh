#!/bin/sh

set -e

DEFAULT_PLUGINS_PATH="/go/src/built-in_plugins"

[ -z "$PLUGINS_PATH" ] && PLUGINS_PATH=$DEFAULT_PLUGINS_PATH

PLUGIN_PREFIX="built-in-"

echo "$PLUGINS_PATH" | grep "$DEFAULT_PLUGINS_PATH" || PLUGIN_PREFIX="ext-"

cd $PLUGINS_PATH

PLUGINS_LIB_PATH="/release/plugins"

mkdir -p "$PLUGINS_LIB_PATH"

for folder in $(ls -d $PLUGINS_PATH/*/); do
  cd "$folder"
  plugin_name=${PLUGIN_PREFIX}$(echo -n "$folder" | grep -o -E "[^/]+/$" | grep -o -E "^[^/]+")
  echo "Building plugin $plugin_name"
  set +e
  rm go.mod go.sum
  set -e
  go mod init "github.com/smart-duck/veradco/$plugin_name"
  go mod edit -replace github.com/smart-duck/veradco=../../veradco
  go mod tidy
  go build -buildmode=plugin -o "$PLUGINS_LIB_PATH/$plugin_name.so" plug.go
  cd ..
done

echo "List of built plugins:"
ls $PLUGINS_LIB_PATH