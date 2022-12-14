# Launch kind with local registry

https://kind.sigs.k8s.io/docs/user/local-registry/

```
sudo local_registry/create_kind_with_local_registry.sh
kk cluster-info
```

```
sudo kind delete cluster
sudo kind get clusters
```

## Export KUBECONFIG

```
sudo chmod +r /root/.kube/config
sudo cp /root/.kube/config /tmp
export KUBECONFIG="/tmp/config"
```


# Push an image

```
sudo docker pull gcr.io/google-samples/hello-app:1.0
sudo docker tag gcr.io/google-samples/hello-app:1.0 localhost:5001/hello-app:1.0
sudo docker images
sudo docker push localhost:5001/hello-app:1.0
kk create deployment hello-server --image=localhost:5001/hello-app:1.0
kk logs hello-server-6fc844b47f-f7jcx
```

# Veradco

```
sudo docker tag veradco/dummy:0.1 localhost:5001/veradco/dummy:0.1
sudo docker images
sudo docker push localhost:5001/veradco/dummy:0.1
kk create ns ns-test
kk apply -f tests/local_registry/pod_veradco_dummy.yaml 
kk get po
kk describe po veradcodummy -n ns-test
kk logs veradcodummy -n ns-test
```