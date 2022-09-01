#!/bin/bash

# KUBECTL="sudo kubectl --context kind-kind"

KUBECTL="kubectl"

[ -z "$KUBECTL_ALIAS" ] || KUBECTL="$KUBECTL_ALIAS"

echo "Delete configuration ConfigMap"
$KUBECTL delete -f veradco_conf.yaml

echo "Delete k8s Secret"
$KUBECTL delete secret veradco-tls -n veradco

echo "Deleting previous k8s admission deployment"
$KUBECTL delete -f deployment.yaml

echo "Deleting previous k8s webhooks for demo"
$KUBECTL delete -f webhooks.yaml
