# GitHub Deploy

## Create secret
```
kubectl create secret generic github-deploy --from-literal=webhook-secret=mysecret --from-literal=github-token=token
```
