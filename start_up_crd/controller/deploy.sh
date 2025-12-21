#!/bin/bash
set -e

echo -n "Please ensure to set up right kubernetes context (y/n): "
read response
if [ "$response" != "y" ]; then
    exit 1
fi

gcloud auth configure-docker us-central1-docker.pkg.dev
dir="$(cd "$(dirname "$0")"; pwd)"
docker build -t us-east1-docker.pkg.dev/lrncrd/test/controller-ns:latest $dir
docker push us-east1-docker.pkg.dev/lrncrd/test/controller-ns:latest
kubectl apply -f $dir/k8s_crd_controller_deploy.yaml
