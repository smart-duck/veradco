#!/bin/bash

echo "Delete previous deployment if exists"

$( dirname "$0" )/delete_all.sh  > /dev/null 2>&1

sleep 4s

# KUBECTL="sudo kubectl --context kind-kind"

KUBECTL="kubectl"

[ -z "$KUBECTL_ALIAS" ] || KUBECTL="$KUBECTL_ALIAS"

echo "Creating certificates"
rm -Rf certs/
mkdir -p certs
openssl req -nodes -new -x509 -keyout certs/ca.key -out certs/ca.crt -subj "/CN=Admission Controller Demo"
openssl genrsa -out certs/admission-tls.key 2048
# openssl req -new -key certs/admission-tls.key -subj "/CN=veradco.veradco.svc" -addext "subjectAltName = DNS.1:veradco.veradco.svc" | openssl x509 -req -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/admission-tls.crt

openssl req -new -key certs/admission-tls.key -subj "/CN=veradco.veradco.svc" -config admission-cert.conf | openssl x509 -req -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/admission-tls.crt -extensions v3_req -extfile admission-cert.conf

echo "Creating namespace veradco"
$KUBECTL apply -f veradco_ns.yaml

echo "Creating configuration ConfigMap"
$KUBECTL apply -f veradco_conf.yaml

echo "Creating k8s Secret"
$KUBECTL create secret tls veradco-tls -n veradco \
    --cert "certs/admission-tls.crt" \
    --key "certs/admission-tls.key"

echo "Creating k8s admission deployment"
$KUBECTL apply -f deployment.yaml

echo "Creating k8s webhooks for demo"
CA_BUNDLE=$(cat certs/ca.crt | base64 | tr -d '\n')
sed -e 's@${CA_BUNDLE}@'"$CA_BUNDLE"'@g' <"webhooks.yaml" | $KUBECTL apply -f -