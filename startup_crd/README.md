# to test out

configure k8s context
sh ./startup_crd/controller/deploy.sh
kubectl logs deployment/ns-controller-startup-nginx --all-pods=true -f
in another terminal
kubectl apply -f ./startup_crd/healthcheck.yaml
