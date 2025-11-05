# Start up CRD

To run things up:
```sh
# configure k8s context and use it
kubectl apply -f ./message_crd/crd.yaml
sh ./message_crd/controller/deploy.sh
kubectl logs deployment/msg-controller-dep --all-pods=true -f
# open another terminal
kubectl apply -f ./message_crd/message.yaml
```
