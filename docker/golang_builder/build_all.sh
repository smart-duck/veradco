#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

if [ ! -f /release/veradcod ]; then
  echo "BUILD veradco"
  /veradco_scripts/build_veradco.sh
fi

if [ ! -d /release/plugins/ ]; then
  echo "BUILD INTERNAL plugins"
  /veradco_scripts/build_plugins.sh
fi

echo "BUILD EXTERNAL plugins"
/veradco_scripts/build_external_plugins.sh

source $(dirname $(readlink -f $0))/end_any_script.source

# Copy generated plugins to /app
if [ -d "/app" ]; then
  echo "Copy veradcod to /app, also plugins folder"
  cp -fr /release/* /app/
  chmod +x /app/veradcod
fi