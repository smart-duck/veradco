#!/bin/bash

echo "==> Tag image  with local registry path localhost:5001/$1"
docker tag $1 localhost:5001/$1
echo "==> List this image"
docker images | grep $(echo $1 | cut -d':' -f 1)
echo "==> Push"
docker push localhost:5001/$1