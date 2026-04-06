# crd-lrn
TODO: aggr_crd, notification_crd

# GCP setup
- general
```bash
gcloud auth login
gcloud services enable cloudbuild.googleapis.com
gcloud config set project lrncrd-481920
# gcloud projects get-iam-policy lrncrd-481920 --flatten="bindings[].members" --filter="bindings.members:user:fcdd227@gmail.com" --format="table(bindings.role)"
# gcloud projects add-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/cloudbuild.builds.editor"
# gcloud projects add-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/artifactregistry.writer"
# gcloud projects remove-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/cloudbuild.builds.editor"
# gcloud projects remove-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/artifactregistry.writer"
```
- k9s
```bash
curl -sS https://webinstall.dev/k9s | bash
# restart profile
```
- node, go, kubectl, gcloud, helm, ...
    - installed by default

# Raspberry PI setup
- ssh
    - VSCode
        - delete `~/.ssh/known_hosts` and `~/.ssh/known_hosts.old`
        - update `~/.ssh/config` by adding a new host
        - even though you had the same configuration
- CLI
```bash
sudo apt update && sudo apt upgrade -y
echo "alias ll='ls --all -l'" >> ~/.bashrc
# install docker
sudo apt install docker.io git -y
sudo chown root:$USER /var/run/docker.sock
# install docker-compose
sudo apt install docker-compose
# pull kubernetes packages (install only kubectl)
curl -fsSL "https://pkgs.k8s.io/core:/stable:/$(curl -L -s https://dl.k8s.io/release/stable.txt | sed 's/\.[0-9]*$//')/deb/Release.key" | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/$(curl -L -s https://dl.k8s.io/release/stable.txt | sed 's/\.[0-9]*$//')/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubectl
# install k9s
curl -sS https://webinstall.dev/k9s | bash
sudo reboot # reboot machine
docker run -d -p 80:80 nginx # router configured in the way to aim to internal network port 80 and translate it to external ip on port 80 - http://173.174.98.86/
### make accessible from outside by exposing port in home router
### set domain to public ip
### sometimes domain isn't reachable from local network due to local network can't reach public ip created in local network, but tp link router resolves it automatically
```
- to run simple server with SSL
```bash
cd ~
mkdir project
cd project
mkdir caddy_with_ssl
cd caddy_with_ssl
echo 'general-solution.com {
  root * /usr/share/caddy
  file_server
}' > Caddyfile
echo '<!DOCTYPE html>
<html>
<head>
  <title>My Caddy Site</title>
</head>
<body>
  <h1>Hello from Caddy 🚀</h1>
  <p>It works well!</p>
</body>
</html>' > index.html
docker run -d --name caddy -p 80:80 -p 443:443 -v "./index.html:/usr/share/caddy/index.html" -v "./Caddyfile:/etc/caddy/Caddyfile" -v caddy_data:/data -v caddy_config:/config caddy
```
```bash
cd ~
mkdir project
cd project
mkdir nginx_with_ssl
cd nginx_with_ssl
echo 'server {
    listen 80;
    server_name general-solution.com www.general-solution.com;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
}' > nonssl-default.conf
echo 'server {
    listen 80;
    server_name general-solution.com www.general-solution.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name general-solution.com www.general-solution.com;

    ssl_certificate /etc/letsencrypt/live/general-solution.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/general-solution.com/privkey.pem;

    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }
}' > default.conf
echo '<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<body>
    <h1>Running...</h1>
    <h2>Working...</h2>
    <h2>Test</h2>
</body>
</html>' > index.html
echo 'version: "3.9"
services:
  nginx-nonssl:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./index.html:/usr/share/nginx/html/index.html
      - ./nonssl-default.conf:/etc/nginx/conf.d/default.conf
      - ./www:/var/www/certbot
  certbot-init:
    image: certbot/certbot
    entrypoint: sh -c "certbot certonly --webroot -w /var/www/certbot -d general-solution.com --email you@email.com --agree-tos --no-eff-email --keep-until-expiring"
    volumes:
      - ./certbot:/etc/letsencrypt
      - ./www:/var/www/certbot
    depends_on:
      - nginx-nonssl
' > docker-compose-cert-gen.yaml
echo 'version: "3.9"
services:
  nginx:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./index.html:/usr/share/nginx/html/index.html
      - ./default.conf:/etc/nginx/conf.d/default.conf
      - ./www:/var/www/certbot
      - ./certbot:/etc/letsencrypt
  certbot:
    image: certbot/certbot
    entrypoint: >
      sh -c "trap exit TERM;
      while :; do
        sleep 10d & wait $${!};
        certbot renew;
      done"
    volumes:
      - ./certbot:/etc/letsencrypt
      - ./www:/var/www/certbot
    depends_on:
      - nginx
' > docker-compose.yaml
docker-compose -f ./docker-compose-cert-gen.yaml up -d
sleep 30s
docker-compose -f ./docker-compose-cert-gen.yaml stop
docker-compose up -d
```
```bash
cd ~
mkdir project
cd project
mkdir traefik_with_ssl
cd traefik_with_ssl
echo 'server {
    listen 80;
    root /usr/share/nginx/html;
    location / {
        try_files /index.html =200;
    }
}
' > default.conf
echo 'api:
  insecure: true
entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"
providers:
  docker: {}
certificatesResolvers:
  letsencrypt:
    acme:
      email: "you@email.com"
      storage: "/letsencrypt/acme.json"
      tlsChallenge: {}
' > traefik.yml
echo 'version: "3.9"
services:
  traefik:
    image: traefik:v3.6
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./traefik.yml:/etc/traefik/traefik.yml:ro
      - ./letsencrypt:/letsencrypt
  app:
    image: nginx:latest
    volumes:
      - ./default.conf:/etc/nginx/conf.d/default.conf
    labels:
      - "traefik.http.routers.app.rule=Host(`general-solution.com`)"
      - "traefik.http.routers.app.entrypoints=websecure"
      - "traefik.http.routers.app.tls.certresolver=letsencrypt"
' > docker-compose.yaml
docker-compose up -d
```
```bash
cd ~
mkdir project
cd project
mkdir envoy_with_ssl
cd envoy_with_ssl
```
- setup kubernetes cluster
```bash
sudo apt remove zram-tools -y
sudo apt remove systemd-zram-generator -y
sudo swapoff -a
echo -n ' cgroup_enable=cpuset cgroup_enable=memory cgroup_memory=1' | sudo tee -a /boot/firmware/cmdline.txt
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
sudo reboot
# ----------------------- kind cli -----------------------
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-amd64
[ $(uname -m) = aarch64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-arm64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
echo 'kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: dev-cluster
nodes:
  - role: control-plane
    kubeadmConfigPatches:
    - |
      kind: InitConfiguration
      nodeRegistration:
        kubeletExtraArgs:
          node-labels: "ingress-ready=true"
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
      protocol: TCP
    - containerPort: 443
      hostPort: 443
      protocol: TCP
  - role: worker
  - role: worker
' > kind-config.yaml
kind create cluster --config kind-config.yaml
# ----------------------- k3d cli -----------------------
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create dev-cluster --servers 1 --agents 2 --wait
# ------- k3s cli (for 2 machines in local network) -------
# on first machine
curl -sfL https://get.k3s.io | sh -
echo 'tls-san:
  - kuber.general-solution.com' > /etc/rancher/k3s/config.yaml
sudo systemctl restart k3s
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
kubectl get pods
ip a | grep 192 # local ip
sudo cat /var/lib/rancher/k3s/server/node-token # token
# on second machine
curl -sfL https://get.k3s.io | K3S_URL=https://<control-plane-ip>:6443 K3S_TOKEN=<tone> sh -
mkdir ~/.kube
# copy ~/.kube/config from first machine to ~/.kube/config in second machine
kubectl apply -f https://raw.githubusercontent.com/ahd3r/crd-lrn/refs/heads/main/start_up/nginx.yaml
echo "apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: general-solution
spec:
  ingressClassName: traefik
  rules:
    - host: general-solution.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 3200" | kubectl apply -f -
curl general-solution.com
# to stop cluster
# on first machine
/usr/local/bin/k3s-killall.sh # stops all related processes
/usr/local/bin/k3s-uninstall.sh # removes all related binaries
# on second machine
/usr/local/bin/k3s-killall.sh # stops all related processes
/usr/local/bin/k3s-agent-uninstall.sh # removes all related binaries
# ------------------------ kubeadm ------------------------
# kubernetes package already pulled (line ~39)
sudo apt-get install -y kubeadm kubelet
sudo systemctl enable kubelet
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
mkdir -p ~/.kube
sudo cp -i /etc/kubernetes/admin.conf ~/.kube/config
sudo chown $(id -u):$(id -g) ~/.kube/config
# to stop cluster
sudo kubeadm reset -f
sudo rm -rf /etc/cni/net.d
sudo rm -rf ~/.kube
# ----------------------- minikube -----------------------
```
- install opencode
