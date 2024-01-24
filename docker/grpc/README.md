# How to

```
cd veradco
sudo docker build -t smartduck/veradco:0.1.4 -f Dockerfile.grpc .
sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco:0.1.4
```