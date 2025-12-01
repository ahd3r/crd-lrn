# Start up CRD

To run things up:
```sh
# configure k8s context and use it
kubectl apply -f ./start_up_crd/crd.yaml
sh ./start_up_crd/controller/deploy.sh
kubectl logs deployment/ns-controller-dep --all-pods=true -f
# open another terminal
kubectl apply -f ./start_up_crd/nginx_start.yaml
curl 0.0.0.0:30008
```

Generates:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment-from-crd-cc
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: my-nginx-service-from-crd-cc
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 3200
      targetPort: 80
      nodePort: <--->
  type: NodePort
```
