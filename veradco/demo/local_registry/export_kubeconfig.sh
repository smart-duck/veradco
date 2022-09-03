#!/bin/bash

sudo chmod +r /root/.kube/config
sudo cp /root/.kube/config /tmp
export KUBECONFIG="/tmp/config"