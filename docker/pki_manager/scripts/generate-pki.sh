#!/bin/sh

set -e

CURRDIR=$(dirname $(readlink -f $0))

echo "Creating certificates"
rm -Rf /etc/certs/ || true
mkdir -p /etc/certs/
openssl req -nodes -new -x509 -days 3600 -keyout /etc/certs/ca.key -out /etc/certs/ca.crt -subj "/CN=Veradco Admission Controller CA"
openssl genrsa -out /etc/certs/tls.key 2048
openssl req -new -key /etc/certs/tls.key -subj "/CN=veradco.veradco.svc" -config $CURRDIR/veradco-cert.conf | openssl x509 -req -CA /etc/certs/ca.crt -CAkey /etc/certs/ca.key -CAcreateserial -days 3600 -out /etc/certs/tls.crt -extensions v3_req -extfile $CURRDIR/veradco-cert.conf

export CA_BUNDLE=$(cat /etc/certs/ca.crt | base64 | tr -d '\n')

# cp -f $CURRDIR/vars.env.tmpl /etc/certs/vars.env

rm /etc/certs/ca.key

ls /etc/certs/

echo "CA_BUNDLE=$CA_BUNDLE" | tee -a /etc/certs/vars.env

echo "Update webhooks file"

source /etc/certs/vars.env

echo "Delete potential remaining webhooks:"
remainings=$(kubectl get validatingwebhookconfiguration.admissionregistration.k8s.io --selector=createByVeradco==true  | cut -d" " -f1)
for webhook in $remainings; do
  wname=$(echo $webhook | cut -d" " -f1)
  if [[ "$wname" != "NAME" ]]; then
    echo "Remove webhook $wname"
    kubectl delete validatingwebhookconfiguration "$wname"
  fi
done

remainings=$(kubectl get mutatingwebhookconfiguration.admissionregistration.k8s.io --selector=createByVeradco==true  | cut -d" " -f1)
for webhook in $remainings; do
  wname=$(echo $webhook | cut -d" " -f1)
  if [[ "$wname" != "NAME" ]]; then
    echo "Remove webhook $wname"
    kubectl delete mutatingwebhookconfiguration "$wname"
  fi
done

echo "Apply webhooks:"

if [[ -f /conf/webhooks.yaml ]]; then

  sed -E "s#(^\s*caBundle:\s*)(.+$)#\\1${CA_BUNDLE}#g" /conf/webhooks.yaml > /tmp/webhooks.yaml

  echo "Webhooks:"
  cat /tmp/webhooks.yaml

  CREATE_WEBHOOK=$(kubectl apply -f /tmp/webhooks.yaml)

  NO_WEBHOOK_CREATED="true"

  for webhook in $CREATE_WEBHOOK; do
    kindname=$(echo $webhook | cut -d" " -f1)
    found="true"
    echo $kindname | grep -q -E "^[^/]+/[^/]+$" || found="false"
    if [[ "$found" == "true" ]]; then
      echo "Add label to recognize $kindname"
      kubectl label --overwrite $kindname createByVeradco=true
      NO_WEBHOOK_CREATED="false"
    fi
  done

  if [[ "$NO_WEBHOOK_CREATED" == "true" ]]; then
    echo "No webhook created. Script terminated with error 2"
    exit 2
  fi

  # COUNTER=0
  # while true; do
  #   CREATE_WEBHOOK=$(kubectl apply -f /tmp/webhooks.yaml)
  #   if [[ "$?" == "0" || "$COUNTER" == "100" ]]; then
  #     if [[ "$?" == "0" ]]; then
  #       for webhook in $CREATE_WEBHOOK; do
  #         kindname=$(echo $webhook | cut -d" " -f1)
  #         echo $kindname | grep -q -E "^[^/]+/[^/]+$"
  #         if [[ "$?" == "0" ]]; then 
  #           echo "Add label to recognize $kindname"
  #           kubectl label --overwrite $kindname createByVeradco=true
  #         fi
  #       done
  #     fi
  #     break
  #   fi
  #   COUNTER=$(( COUNTER + 1 ))
  #   echo "TRY AGAIN $COUNTER"
  # done
# elif [[ -f /conf_veradco/veradco.yaml ]]; then
#   export VERADCO_CONF="/conf_veradco/veradco.yaml"
#   source generate_webhooks_from_veradco_conf.sh
else
  # echo "No conf found to create webhooks (neither custom conf nor veradco conf)"
  echo "No conf found to create webhooks. Script terminated with error 1"
  exit 1
fi
