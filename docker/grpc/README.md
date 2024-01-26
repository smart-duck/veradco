# How to

```
cd veradco
sudo docker build -t smartduck/veradco:v0.2.0 -f Dockerfile.grpc .
sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco:v0.2.0
```