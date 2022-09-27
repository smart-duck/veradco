#!/bin/bash

CURRDIR=$(dirname $(readlink -f $0))

echo "Creating certificates"
rm -Rf $CURRDIR/base/certs/
mkdir -p $CURRDIR/base/certs/
openssl req -nodes -new -x509 -days 3600 -keyout $CURRDIR/base/certs/ca.key -out $CURRDIR/base/certs/ca.crt -subj "/CN=Veradco Admission Controller CA"
openssl genrsa -out $CURRDIR/base/certs/tls.key 2048
openssl req -new -key $CURRDIR/base/certs/tls.key -subj "/CN=veradco.veradco.svc" -config $CURRDIR/veradco-cert.conf | openssl x509 -req -CA $CURRDIR/base/certs/ca.crt -CAkey $CURRDIR/base/certs/ca.key -CAcreateserial -days 3600 -out $CURRDIR/base/certs/tls.crt -extensions v3_req -extfile $CURRDIR/veradco-cert.conf

export CA_BUNDLE=$(cat $CURRDIR/base/certs/ca.crt | base64 | tr -d '\n')

cp -f $CURRDIR/base/vars.env.tmpl $CURRDIR/base/certs/vars.env

echo "CA_BUNDLE=$CA_BUNDLE" | tee -a $CURRDIR/base/certs/vars.env