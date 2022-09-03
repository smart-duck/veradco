# veradco Golang builder image

Create veradco Golang builder image. This image is used as an init container to load plugins.

veradcod and built-in plugins are built at image build.

```
sudo docker build -t smartduck/veradco-golang-builder:0.1 -f ./Dockerfile.golang_builder .
sudo veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-golang-builder:0.1
```

## To read

https://iximiuz.com/en/posts/kubernetes-api-go-types-and-common-machinery/