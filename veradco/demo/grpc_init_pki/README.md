# How to

```
cd veradco
go build -o demo/grpc_init_pki/veradcod cmd/serverd/main.go
cd demo/grpc_init_pki/
sudo docker build -t smartduck/veradco:0.1.grpc -f Dockerfile .
sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco:0.1.grpc
k delete ns veradco
k apply -f veradco_ns.yaml
k apply -f veradco_conf.yaml
k apply -f cmwebhooks.yaml
~/go/bin/stern -n veradco veradco &
k apply -f deployment.yaml

# Test
k apply -f ../pods/02_success_pod_creation_test.yaml
k delete -f ../pods/02_success_pod_creation_test.yaml

# Debug SVC:
kubectl exec -it webserver3 -- /bin/sh
dummyplugin.default.svc.cluster.local:50051

```