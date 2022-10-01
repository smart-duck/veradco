#!/bin/sh

# Special case copy to go src
if [ ! -z "$TO_BUILD_FOLDER" ]; then
  echo "Remove all from /go/src"
  rm -Rf /go/src/*
  echo "Copy content $TO_BUILD_FOLDER to /go/src"
  cp -R $TO_BUILD_FOLDER/* /go/src/
fi

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

# Special chown
if [ ! -z "$TO_BUILD_CHMOD" ]; then
  echo "Change owner of built binaries to $TO_BUILD_CHMOD"
  chown -R $TO_BUILD_CHMOD /release/*
fi