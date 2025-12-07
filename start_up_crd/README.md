# Start up CRD

To run things up:
```sh
# configure k8s context and use it
kubectl apply -f ./start_up_crd/crd.yaml
sh ./start_up_crd/controller/deploy.sh
kubectl logs deployment/ns-controller-dep --all-pods=true -f
# open another terminal
kubectl apply -f ./start_up_crd/nginx_start.yaml
curl 35.188.72.228:30008
```

Generates:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  finalizers:
  - true.test/finalizer
  labels:
    manged: cc
  name: nginx-deployment-from-crd-cc-nginx-init-app
  namespace: ns-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - image: nginx:latest
        name: nginx
        ports:
        - containerPort: 80
          protocol: TCP
---
apiVersion: v1
kind: Service
metadata:
  finalizers:
  - true.test/finalizer
  labels:
    manged: cc
  name: my-nginx-service-from-crd-cc-nginx-init-app
  namespace: ns-namespace
spec:
  ports:
  - nodePort: < --- >
    port: 3200
    protocol: TCP
    targetPort: 80
  selector:
    app: nginx
  type: NodePort
```
