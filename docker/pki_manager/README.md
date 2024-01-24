# PKI manager

## Overview

Manages Veradco service PKI via an init container.

## Package

```
sudo docker build -t smartduck/veradco_pki_manager:0.1.4 -f ./Dockerfile .

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco_pki_manager:0.1.4
```
