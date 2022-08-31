#!/bin/bash

KUBECTL="sudo kubectl --context kind-kind"

echo "Delete k8s Secret"
$KUBECTL delete secret veradco-tls -n veradco

echo "Deleting previous k8s admission deployment"
$KUBECTL delete -f deployment.yaml

echo "Deleting previous k8s webhooks for demo"
$KUBECTL delete -f webhooks.yaml
