# Overview

Followed tuto:
https://www.faizanbashir.me/guide-to-create-kubernetes-operator-with-golang

Prerequisites:
https://sdk.operatorframework.io/docs/installation/


## Done

mkdir veradco-operator && cd veradco-operator

../operator-sdk_linux_amd64 init --domain "veradco-operator.com" --repo "github.com/smart-duck/veradco-operator"

../operator-sdk_linux_amd64 create api --group apps --version v1alpha1 --kind VeracoPlugin --resource --controller

make docker-build docker-push IMG=veradco-operator:latest

## Docker

sudo docker build -t smartduck/veradco-operator:0.1beta1 .
\# sudo docker push localhost:5001/hello-app:1.0

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco-operator:0.1beta1

make [un]deploy IMG="smartduck/veradco-operator:0.1beta1"

namespace/veradco-operator-system created
customresourcedefinition.apiextensions.k8s.io/veracoplugins.apps.veradco-operator.com created
serviceaccount/veradco-operator-controller-manager created
role.rbac.authorization.k8s.io/veradco-operator-leader-election-role created
clusterrole.rbac.authorization.k8s.io/veradco-operator-manager-role created
clusterrole.rbac.authorization.k8s.io/veradco-operator-metrics-reader created
clusterrole.rbac.authorization.k8s.io/veradco-operator-proxy-role created
rolebinding.rbac.authorization.k8s.io/veradco-operator-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/veradco-operator-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/veradco-operator-proxy-rolebinding created
service/veradco-operator-controller-manager-metrics-service created
deployment.apps/veradco-operator-controller-manager created