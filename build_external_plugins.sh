#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

echo "Copy veradcod to /app, also plugins folder"
cp -fr /release/* /app/

chmod +x /app/veradcod

# set -x

set -e

[ -z "$VERADCO_CONF" ] && VERADCO_CONF="/conf/veradco.yaml"

# This script needs a veradco conf unless it is used for CI build!
[ -f "$VERADCO_CONF" ] || exit 0

build_folder="/go/src/ext_plugins"

mkdir -p "$build_folder"

external_plugins_folder="/app/external_plugins"

mkdir -p "$external_plugins_folder"

for plugin in $(cat $VERADCO_CONF | yq '.plugins[].name'); do
  name=$plugin
  path=$(cat $VERADCO_CONF | yq ".plugins[] | select(.name==\"$plugin\") | .path")
  plug_go=$(cat $VERADCO_CONF | yq ".plugins[] | select(.name==\"$plugin\") | .code")
  # echo "plug_go=$plug_go"
  if [ -f "$path" ]; then
    echo "File $path exists."
  else 
    echo "File $path does not exist. Build plugin..."
	# id_plugin="$(uuidgen)"
	id_plugin="$(date '+%Y%m%d%H%M%S%N')"
	plugin_folder="$build_folder/$id_plugin"
	mkdir -p "$plugin_folder"
	go_file="$plugin_folder/plug.go"
	echo $plug_go | base64 -d > $go_file
	cd "$plugin_folder"
	go mod init "github.com/smart-duck/veradco/$id_plugin"
	go mod edit -replace github.com/smart-duck/veradco=../../veradco
	go mod tidy
	go build -buildmode=plugin -o "$path" plug.go
  fi
done

echo "List of external plugins:"
ls $external_plugins_folder

echo "app content:"
ls -lRt /app

# sleep 1000s

source $(dirname $(readlink -f $0))/end_any_script.source