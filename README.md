# crd-lrn
TODO: aggr_crd, notification_crd

# GCP setup
- general
```bash
gcloud auth login
gcloud services enable cloudbuild.googleapis.com
gcloud config set project lrncrd-481920
gcloud projects get-iam-policy lrncrd-481920 --flatten="bindings[].members" --filter="bindings.members:user:fcdd227@gmail.com" --format="table(bindings.role)"
gcloud projects add-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/cloudbuild.builds.editor"
gcloud projects add-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/artifactregistry.writer"
gcloud projects remove-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/cloudbuild.builds.editor"
gcloud projects remove-iam-policy-binding lrncrd-481920 --member="user:fcdd227@gmail.com" --role="roles/artifactregistry.writer"
```
- k9s
```bash
curl -sS https://webinstall.dev/k9s | bash
# restart profile
```
- node, go, kubectl, gcloud, helm, ...
    - installed by default
