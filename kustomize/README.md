# Step 1

Generate PKI with dedicated script generate-pki.sh

# Step 2

Generate yaml and apply it.

```
kustomize build ../../kustomize/demo/ | kubectl apply -f -
kustomize build ../../kustomize/demo/ | kubectl delete -f -
```