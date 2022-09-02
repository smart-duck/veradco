# Overview

TODO

## Bulk

```
cd /home/lobuntu/go/src/veradco
sudo docker build -t smartduck/veradco-golang-builder:0.1 -f ./Dockerfile.golang_builder .
sudo veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-golang-builder:0.1
cd veradco/demo
sudo docker build -t smartduck/veradco:0.1 -f ../Dockerfile.golang_builder ..
sudo local_registry/push_local_image_to_local_registry.sh smartduck/veradco:0.1
export KUBECTL_ALIAS="sudo kubectl --context kind-kind"
./deploy.sh 
kk apply -f deployments/02_fail_deployment_creation.yaml 
kk apply -f deployments/02_fail_deployment_creation.yaml 
kk apply -f deployments/03_sucess_deployment_creation.yaml 
kk delete -f deployments/03_sucess_deployment_creation.yaml 
kk apply -f pods/01_fail_pod_creation_test.yaml 
kk apply -f pods/02_success_pod_creation_test.yaml 
kk delete -f pods/02_success_pod_creation_test.yaml 
kk logs veradco-5476ddd449-hhwwt --follow
kk logs kube-apiserver-kind-control-plane --follow -n kube-system | grep -i admission
./delete_all.sh 
./deploy.sh 
kk apply -f deployments/02_fail_deployment_creation.yaml 
kk apply -f pods/01_fail_pod_creation_test.yaml 


sudo docker images | grep $(echo douglasmakey/admissioncontroller:0.1 | cut -d':' -f 1)
 1092  sudo tests/local_registry/push_local_image_to_local_registry.sh douglasmakey/admissioncontroller:0.1
 1093  sudo tests/local_registry/push_local_image_to_local_registry.sh douglasmakey/admissioncontroller:0.2

 