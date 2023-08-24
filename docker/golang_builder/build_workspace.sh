#!/bin/sh

# This script builds veradco and the addons that are in the configuration following these rules:
# - if code is filled, plugin is built from code as an external plugin
# - plugin is searched in built-un_plugins folder
# - plugin is searched in external_plugins

set -ev

[ -z "$VERADCO_CONF" ] && VERADCO_CONF="/conf/veradco.yaml"

[ -z "$BUILTIN_PATH" ] && BUILTIN_PATH="/go/src/built-in_plugins"

[ -z "$EXTERNAL_PATH" ] && EXTERNAL_PATH="/go/src/external_plugins"

mkdir -p $EXTERNAL_PATH

PLUGIN_PREFIX="built-in-"

# This script needs a veradco conf unless it is used for CI build!
[ -f "$VERADCO_CONF" ] || exit 1

[ -d "$BUILTIN_PATH" ] || exit 2

PLUGINS_LIB_PATH="/app/plugins"

mkdir -p "$PLUGINS_LIB_PATH"

source $(dirname $(readlink -f $0))/start_any_script.source

echo "Prepare veradco"

cd /go/src

go work init

cd /go/src/veradco
rm go.mod go.sum || true

go mod init github.com/smart-duck/veradco
go mod tidy

go work use .

echo "Prepare plugins"
plugins_to_compile=""

for plugin in $(cat $VERADCO_CONF | yq '.plugins[].name'); do
  name=$plugin
  path=$(cat $VERADCO_CONF | yq ".plugins[] | select(.name==\"$plugin\") | .path")
  plug_go=$(cat $VERADCO_CONF | yq ".plugins[] | select(.name==\"$plugin\") | .code")

  if [ "$plug_go" != "null" ]; then
    # It is an external plugin: build from code
    echo "Prepare plugin $name from code"
    id_plugin="$(date '+%Y%m%d%H%M%S%N')"
    folder=$EXTERNAL_PATH/$id_plugin
    mkdir -p $folder
    cd $folder
    echo $plug_go | base64 -d > $folder/plug.go
    go mod init "github.com/smart-duck/veradco/$id_plugin"
    go mod edit -replace github.com/smart-duck/veradco=../../veradco
    go mod tidy

    go work use .

    plugins_to_compile="$plugins_to_compile $folder:$path"
  else
    # search it in built-in plugin
    plugin_src_path="$BUILTIN_PATH/$name/plug.go"
    echo "Prepare built-in plugin $name from path $plugin_src_path"
    if [ ! -f "$plugin_src_path" ]; then
      echo "File $plugin_src_path does not exist"
      exit 1
    fi
    
    cd $BUILTIN_PATH

    folder="$BUILTIN_PATH/$name/"

    cd "$folder"

    plugin_name="${PLUGIN_PREFIX}$name"

    echo "Prepare built-in plugin $name"

    rm go.mod go.sum || true
    go mod init "github.com/smart-duck/veradco/$plugin_name"

    go mod edit -replace github.com/smart-duck/veradco=../../veradco
    go mod tidy

    go work use .

    plugins_to_compile="$plugins_to_compile $folder:$path"
  fi
done

go work edit -print

go work sync

echo "Build plugins"

for item in $plugins_to_compile; do
  echo $item
  folder=$(echo $item | grep -o -E "^[^:]+")
  path=$(echo $item | grep -o -E "[^:]+$")
  echo "Build plugin $path"
  cd $folder
  go build -buildmode=plugin -o $path plug.go
  ls -l $path
done

echo "Build veradco"
cd /go/src/veradco
go build -o /app/veradcod cmd/serverd/main.go

ls -l /app/veradcod

# folder="$BUILTIN_PATH/implicit_proxy_cache_populator/"

# cd "$folder"
# plugin_name=${PLUGIN_PREFIX}$(echo -n "$folder" | grep -o -E "[^/]+/$" | grep -o -E "^[^/]+")

# echo "Building plugin $plugin_name"
# rm go.mod go.sum || true
# go mod init "github.com/smart-duck/veradco/$plugin_name"
# go mod edit -replace github.com/smart-duck/veradco=../../veradco
# go mod tidy

# go work use .

# go work edit -print

# echo "$plugin_name go.mod:"
# cat go.mod
# echo "$plugin_name go.sum:"
# cat go.sum || true

# echo "go work sync"
# go work sync

# echo "veradco go.mod:"
# cat /go/src/veradco/go.mod
# echo "veradco go.sum:"
# cat /go/src/veradco/go.sum || true
# echo "$plugin_name go.mod:"
# cat go.mod
# echo "$plugin_name go.sum:"
# cat go.sum || true

# echo "build plugin"
# go build -buildmode=plugin -o "$PLUGINS_LIB_PATH/$plugin_name.so" plug.go

# echo "build veradco"
# cd /go/src/veradco
# go build -o /app/veradcod cmd/serverd/main.go

source $(dirname $(readlink -f $0))/end_any_script.source









# echo "veradco go.mod:"
# cat go.mod
# echo "veradco go.sum:"
# cat go.sum || true

# mkdir -p /app

# go build -o /app/veradcod cmd/serverd/main.go

# source $(dirname $(readlink -f $0))/end_any_script.source






# if [ ! -f /app/veradcod ]; then
#   echo "BUILD veradco"
#   /veradco_scripts/build_veradco.sh
# fi

# if [ ! -d /app/plugins/ ]; then
#   echo "BUILD INTERNAL plugins"
#   /veradco_scripts/build_plugins.sh
# fi

# echo "BUILD EXTERNAL plugins"
# /veradco_scripts/build_external_plugins.sh

# source $(dirname $(readlink -f $0))/end_any_script.source

# # Copy generated plugins to /app
# if [ -f "/app/veradcod" ]; then
#   # echo "Copy veradcod to /app, also plugins folder"
#   # cp -fr /release/* /app/
#   chmod +x /app/veradcod
# fi