#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

set -e

shallBuild () {
  SHALL_BUILD="NO"

  [ -f $VERADCO_CONF ] || SHALL_BUILD="YES"

  if [ "$SHALL_BUILD" != "YES" ]; then
    PLUGPATH=$(echo "$1" | sed "s#/release/#/app/#")
    IS_IN_CONF=$(cat $VERADCO_CONF | yq ".plugins[] | select(.path==\"$PLUGPATH\") | .name")
    SHALL_BUILD="YES"
    if [ "$IS_IN_CONF" = "" ]; then
      PLUGPATH="$1"
      IS_IN_CONF=$(cat $VERADCO_CONF | yq ".plugins[] | select(.path==\"$PLUGPATH\") | .name")
      [ "$IS_IN_CONF" != "" ] || SHALL_BUILD="NO"
    fi
  fi
}

DEFAULT_PLUGINS_PATH="/go/src/built-in_plugins"

[ -z "$PLUGINS_PATH" ] && PLUGINS_PATH=$DEFAULT_PLUGINS_PATH

PLUGIN_PREFIX="built-in-"

echo "$PLUGINS_PATH" | grep "$DEFAULT_PLUGINS_PATH" || PLUGIN_PREFIX="ext-"

cd $PLUGINS_PATH

PLUGINS_LIB_PATH="/app/plugins"

[ -z "$VERADCO_CONF" ] && VERADCO_CONF="/conf/veradco.yaml"

mkdir -p "$PLUGINS_LIB_PATH"

for folder in $(ls -d $PLUGINS_PATH/*/); do
  cd "$folder"
  plugin_name=${PLUGIN_PREFIX}$(echo -n "$folder" | grep -o -E "[^/]+/$" | grep -o -E "^[^/]+")

  shallBuild "$PLUGINS_LIB_PATH/$plugin_name.so"

  if [ "$SHALL_BUILD" = "YES" ]; then
    echo "Building plugin $plugin_name"
    set +e
    rm go.mod go.sum
    set -e
    go mod init "github.com/smart-duck/veradco/veradco/$plugin_name"
    go mod edit -replace github.com/smart-duck/veradco/veradco=../../veradco
    go mod tidy
    echo "$plugin_name go.mod:"
    cat go.mod
    echo "$plugin_name go.sum:"
    cat go.sum || true
    go build -buildmode=plugin -o "$PLUGINS_LIB_PATH/$plugin_name.so" plug.go
  else
    echo "NO NEED TO BUILD plugin $plugin_name"
  fi
  cd ..
done

echo "List of built plugins:"
ls $PLUGINS_LIB_PATH

source $(dirname $(readlink -f $0))/end_any_script.source