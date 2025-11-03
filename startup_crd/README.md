# Start up CRD

To run things up:
```sh
# configure k8s context and use it
sh ./startup_crd/controller/deploy.sh
kubectl logs deployment/ns-controller-startup-nginx --all-pods=true -f
# open another terminal
kubectl apply -f ./startup_crd/healthcheck.yaml
```
