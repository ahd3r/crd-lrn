#!/bin/bash
set -e

echo -n "Please ensure to set up right kubernetes context (y/n): "
read response
if [ "$response" != "y" ]; then
    exit 1
fi

gcloud config set project lrncrd-481920
gcloud auth configure-docker us-central1-docker.pkg.dev
dir="$(cd "$(dirname "$0")"; pwd)"
gcloud builds submit $dir --tag us-central1-docker.pkg.dev/lrncrd-481920/ctrls/controller-ns:latest
kubectl apply -f $dir/k8s_crd_controller_deploy.yaml
