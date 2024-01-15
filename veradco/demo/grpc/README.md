# How to

```
cd veradco
go build -o demo/grpc/veradcod cmd/serverd/main.go
cd demo/grpc/
sudo docker build -t smartduck/veradco:0.1.grpc -f Dockerfile .
sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco:0.1.grpcsh smartduck/veradco:0.1.grpc
k delete ns veradco
k apply -f veradco_ns.yaml
k apply -f veradco_conf.yaml
k create secret tls veradco-tls -n veradco \
    --cert "certs/admission-tls.crt" \
    --key "certs/admission-tls.key"
~/go/bin/stern -n veradco veradco &
k apply -f deployment.yaml
CA_BUNDLE=$(cat certs/ca.crt | base64 | tr -d '\n')
sed -e 's@${CA_BUNDLE}@'"$CA_BUNDLE"'@g' <"webhooks.yaml" | k apply -f -

# Test
k apply -f ../pods/02_success_pod_creation_test.yaml
k delete -f ../pods/02_success_pod_creation_test.yaml

# Debug SVC:
kubectl exec -it webserver3 -- /bin/sh
dummyplugin.default.svc.cluster.local:50051

```