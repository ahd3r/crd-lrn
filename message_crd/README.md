# Start up CRD

To run things up:
```sh
# configure k8s context and use it
kubectl apply -f ./message_crd/crd.yaml # creates CRD within cluster
sh ./message_crd/controller/deploy.sh # deploy custom controller into cluster (will work only if runs within the cluster)
kubectl logs deployment/msg-controller-dep --all-pods=true -f # watch logs of the custom controller
# open another terminal
kubectl apply -f ./message_crd/message.yaml # create a resource that trigger controller
```
