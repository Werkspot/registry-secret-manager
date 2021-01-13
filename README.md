# Container Registry Secret Manager

Manages the creation and distribution of credentials for container registries.

## Installation

```
$ git clone https://github.com/Werkspot/registry-secret-manager
$ helm upgrade registry-secret-manager --namespace registry-secret-manager --values helm/values.yaml registry-secret-manager/helm
```

## TODO

- [x] Add support for DockerHub and ECR registries
- [x] Listen for new ServiceAccounts creation via a webhook
- [x] Reconcile ServiceAccounts (create Secrets and inject its name in `ImagePullSecrets`)
- [x] Reconcile Secrets (renew ECR tokens every 3 hours)
- [ ] Optimize ECR token usage (now each request/reconcile performs a new login)
- [x] Make DockerHub and ECR registries optional
- [ ] Use the same logging client for Controller-Runtime, Kubernetes Client, Webhook and Reconcilers
- [ ] Make the Helm Chart available somewhere
