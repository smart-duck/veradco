#!/bin/sh


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