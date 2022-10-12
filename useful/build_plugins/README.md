# Build binaries with provided docker image

It is possible to build plugins by using the docker image used by the Veradco init container.

## Build a plugin

To build a plugin it is suitable to use the Veracdo docker image dedicated to build. It avoids compatibility issues between Veradco and the plugin (by example "plugin was built with a different version of package github.com/gogo/protobuf/proto").

```
docker run -v /home/lobuntu/go/src/veradco/useful/build_plugins/a_plugin:/go/src/build/plugin1 smartduck/veradco-golang-builder:0.1 /bin/sh -c "/veradco_scripts/build_a_plugin.sh \"/go/src/build/plugin1\""
```

Explanation:
- The end of the path you provided is used as plugin name: plugin1 in the example above.
- It is mandatory to respect this path: /go/src/build/[plugin_name].
- The binary built ([plugin_name].so) is in the source folder.

## Build all provided plugins + Veradco

From Veradco sources in the building image:
```
docker run -v /home/lobuntu/go/src/veradco/useful/build_plugins/veradco_and_built-ins:/release smartduck/veradco-golang-builder:0.1 /bin/sh -c "/veradco_scripts/build_all.sh"
```

From source on the host:
```
docker run --rm \
  --env TO_BUILD_FOLDER="/to_build" \
  --env TO_BUILD_CHMOD="1000:1000" \
  -v /home/lobuntu/go/src/veradco/useful/build_plugins/veradco_and_built-ins:/release \
  -v /home/lobuntu/go/src/veradco/veradco:/to_build/veradco \
  -v /home/lobuntu/go/src/veradco/built-in_plugins:/to_build/built-in_plugins \
  smartduck/veradco-golang-builder:0.1 /bin/sh -c "/veradco_scripts/build_all.sh"
```
If TO_BUILD_FOLDER environment variable is defined, then /go/src folder is emptied and content of TO_BUILD_FOLDER folder is copied in. It prevents modification of rights in sources and also update of go.mod and go.sum files.

If TO_BUILD_CHMOD is defined, at the end the following command is launched: chown -R $TO_BUILD_CHMOD /release/*.

From source on the host:
```
docker run --rm \
  --env TO_BUILD_FOLDER="/to_build" \
  --env TO_BUILD_CHMOD="1000:1000" \
  --env VERADCO_CONF="/conf/veradco_conf.yaml" \
  -v /home/lobuntu/go/src/veradco/useful/build_plugins/conf:/conf \
  -v /home/lobuntu/go/src/veradco/useful/build_plugins/veradco_and_built-ins:/release \
  -v /home/lobuntu/go/src/veradco/veradco:/to_build/veradco \
  -v /home/lobuntu/go/src/veradco/built-in_plugins:/to_build/built-in_plugins \
  smartduck/veradco-golang-builder:v0.1.0 /bin/sh -c "/veradco_scripts/build_all.sh"
```
In the above example, we provided a Veradco configuration and set an environment variable that define its location. Only the plugins defined in the configuration are built.

Notes:
- If binaries are already in the target folder, the script will do nothing: empty the folder before launching.
- Built binaries have root belonging. To prevent that use variable TO_BUILD_CHMOD.
- If you want to build only Veradco, use /veradco_scripts/build_veradco.sh.
- If you want to build only provided plugins, use /veradco_scripts/build_plugins.sh.

## Build a standalone image with some plugins and veradco

To do it, use the docker file Dockerfile.standalone.

```
docker build --no-cache -t smartduck/veradco-standalone:v0.1.0 -f ./docker/standalone/Dockerfile.standalone ./useful/build_plugins/veradco_and_built-ins/
```

Push to local registry for tests:
```
sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-standalone:v0.1.0
```
