# Build docker images

## Order

[sudo] ./build_golang_builder.sh [SERVER|ALL|NOTHING]

Then, [sudo] build_run.sh or [sudo] build_run_standalone.sh

## Scenarii

### Build a standalone image

A standalone image is intended for a Veradco deployment without init container. It embeds veradcod binary and all necessary plugin binaries (.so). The deployment is as lightweight as possible but in return this kind of deployment cannot do building on the fly.

This example build veradco, the harbor_proxy_cache_populator plugin and an external plugin from its base64 encoded code in the configuration file.
```
# Create a golang builder image without built binaries: so dedicated to building
sudo ./build_golang_builder.sh NOTHING
# Build a standalone image previously built image and the provided standalone/veradco_conf.yaml configuration.
sudo ./build_run_standalone.sh
```

Distroless one:
```
./build_run_standalone.sh "v0.1.0" "v0.1.0" "./Dockerfile.standalone.distroless"
```