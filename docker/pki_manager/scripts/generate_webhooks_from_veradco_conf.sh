#!/bin/sh

set -e

YQ_CMD="yq_linux_amd64"

chmod +x $YQ_CMD

if [[ "$VERADCO_CONF" != "" && -f $VERADCO_CONF ]]; then
  pathVeradcoConf=$VERADCO_CONF
  allEndpoints="/validate/pods /mutate/pods /validate/deployments /mutate/deployments /validate/daemonsets /mutate/daemonsets /validate/statefulsets /mutate/statefulsets /validate/others /mutate/others"
  listOfEndpoints=""
  for endpoints in $(cat $pathVeradcoConf | /$YQ_CMD ".plugins[].endpoints"); do
    for availEpt in $allEndpoints; do
      found="true"
      echo $availEpt | grep -E -q $endpoints || found="false"
      if [[ "$found" == "true" ]]; then
        # echo "FOUNDFOUNDFOUNDFOUNDFOUNDFOUNDFOUNDFOUND1"
        found="false"
        echo $listOfEndpoints | grep -E -q $availEpt || found="true"
        if [[ "$found" == "true" ]]; then
          # echo "FOUNDFOUNDFOUNDFOUNDFOUNDFOUNDFOUNDFOUND2"
          # echo "Add endpoint $availEpt"
          listOfEndpoints="$listOfEndpoints$availEpt "
        fi
      fi
    done
  done

  echo "listOfEndpoints=$listOfEndpoints"

  for ep in $listOfEndpoints; do
    echo "Add endpoint $ep"
# Add endpoint /validate/pods
# Add endpoint /mutate/pods
    nameWhook=$(echo $ep | sed "s#/##g")-veradco

    isValidatingWh="true"
    echo $ep | grep -i -q "/validate/" || isValidatingWh="false"

    if [[ "$isValidatingWh" == "true" ]]; then

# cat <<EOF | kubectl apply -f -
cat <<EOF | kubectl apply -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: $nameWhook
webhooks:
  - name: $nameWhook.default.svc
    sideEffects: None
    admissionReviewVersions: ["v1"]
    clientConfig:
      service:
        name: veradco
        namespace: veradco
        path: "$ep"
      caBundle: "${CA_BUNDLE}"
    rules:
      - operations: ["CREATE", "UPDATE", "DELETE", "CONNECT"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Ignore
EOF

    else

cat <<EOF | kubectl apply -f -
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: $nameWhook
webhooks:
  - name: $nameWhook.default.svc
    sideEffects: None
    admissionReviewVersions: ["v1"]
    clientConfig:
      service:
        name: veradco
        namespace: veradco
        path: "$ep"
      caBundle: "${CA_BUNDLE}"
    rules:
      - operations: ["CREATE", "UPDATE", "DELETE", "CONNECT"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Ignore
EOF

    fi

  done
else
  echo "Veradco conf not found"
  exit 1
fi