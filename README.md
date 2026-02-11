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

# Raspberry PI 5 setup
- ssh
    - VSCode
        - delete the previous one
        - create a new one
        - even though you had the same configuration, it wouldn't work, you need to recreate for some reason
- CLI
```bash
echo "alias ll='ls --all -l'" >> ~/.bashrc
sudo apt update && sudo apt upgrade -y
# install docker
sudo apt install docker.io
sudo chown root:$USER /var/run/docker.sock
# install docker-compose
sudo apt install docker-compose
# install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
# install k9s
curl -sS https://webinstall.dev/k9s | bash
### reboot machine
docker run -d -p 80:80 nginx # router configured in the way to aim to internal network port 80 and translate it to external ip on port 80 - http://173.174.98.86/
### make accessible from outside by exposing port in home router
### set domain to public ip
### sometimes domain isn't reachable from local network due to local network can't reach public ip created in local network, but tp link resolves it automatically
```
- to run simple server with SSL
```bash
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
- setup kubernetes cluster
```bash
```
