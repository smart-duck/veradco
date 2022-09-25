#!/bin/bash

script_abs_path="$(dirname $(readlink -f $0))"

kubectl create ns ds-ns
kubectl apply -f $script_abs_path/deployments/01_namespaces.yaml

while true; do \

  kubectl delete -f $script_abs_path/pods/01_fail_pod_creation_test.yaml
  kubectl delete -f $script_abs_path/pods/02_fail_pod_creation_test.yaml
  kubectl delete -f $script_abs_path/pods/02_success_pod_creation_test.yaml
  kubectl delete -f $script_abs_path/pods/03_success_pod_creation_test_special.yaml
  kubectl delete -f $script_abs_path/pods/04_multiple_alpine.yaml
  kubectl delete -f $script_abs_path/deployments/02_fail_deployment_creation.yaml
  kubectl delete -f $script_abs_path/deployments/03_sucess_deployment_creation.yaml
  kubectl delete -f $script_abs_path/daemonsets/daemonset.yaml

  kubectl apply -f $script_abs_path/pods/01_fail_pod_creation_test.yaml
  kubectl apply -f $script_abs_path/pods/02_fail_pod_creation_test.yaml
  kubectl apply -f $script_abs_path/pods/02_success_pod_creation_test.yaml
  kubectl apply -f $script_abs_path/pods/03_success_pod_creation_test_special.yaml
  kubectl apply -f $script_abs_path/pods/04_multiple_alpine.yaml
  kubectl apply -f $script_abs_path/deployments/02_fail_deployment_creation.yaml
  kubectl apply -f $script_abs_path/deployments/03_sucess_deployment_creation.yaml
  kubectl apply -f $script_abs_path/daemonsets/daemonset.yaml
done