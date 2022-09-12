#!/bin/bash

kubectl create ns ds-ns
kubectl apply -f deployments/01_namespaces.yaml

while true; do \

  kubectl delete -f pods/01_fail_pod_creation_test.yaml
  kubectl delete -f pods/02_fail_pod_creation_test.yaml
  kubectl delete -f pods/02_success_pod_creation_test.yaml
  kubectl delete -f pods/03_success_pod_creation_test_special.yaml
  kubectl delete -f deployments/02_fail_deployment_creation.yaml
  kubectl delete -f deployments/03_sucess_deployment_creation.yaml
  kubectl delete -f daemonsets/daemonset.yaml

  kubectl apply -f pods/01_fail_pod_creation_test.yaml
  kubectl apply -f pods/02_fail_pod_creation_test.yaml
  kubectl apply -f pods/02_success_pod_creation_test.yaml
  kubectl apply -f pods/03_success_pod_creation_test_special.yaml
  kubectl apply -f deployments/02_fail_deployment_creation.yaml
  kubectl apply -f deployments/03_sucess_deployment_creation.yaml
  kubectl apply -f daemonsets/daemonset.yaml
done