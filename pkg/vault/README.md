# Vault

## Setting up K8S authentication

To set up authentication between Vault & Kubernetes, you need:
    
- [Kubectl](https://docs.docker.com/desktop/kubernetes/#enable-kubernetes)
- [Hashicorp vault server](https://www.vaultproject.io/downloads) - a dev server can be started with `vault server -dev`

If you already have a vault server running somewhere (e.g dev-util or mini-vms), you can use the cli in the downloaded vault.
Export the following env variables:

```shell
export VAULT_ADDR=http://localhost:8200 // whereever vault is
export VAULT_TOKEN=replace_with_token_value
```

Create a service account, secret & ClusterRoleBinding

```shell
cat <<EOF | kubectl create -f -
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vault-auth
---
apiVersion: v1
kind: Secret
metadata:
  name: vault-auth
  annotations:
    kubernetes.io/service-account.name: vault-auth
type: kubernetes.io/service-account-token
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: vault-auth
    namespace: default
EOF
```

Enable k8s authentication in vault
```shell
vault auth enable kubernetes
```
(if you're using the dev-util version, this is already enabled)

Save the JWT, Cert & Host URL & configure the auth method
```shell
TOKEN_REVIEW_JWT=$(kubectl get secret vault-auth -o go-template='{{ .data.token }}' | base64 --decode)
KUBE_CA_CERT=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 --decode)
KUBE_HOST=$(kubectl config view --raw --minify --flatten --output='jsonpath={.clusters[].cluster.server}')

vault write auth/kubernetes/config token_reviewer_jwt="$TOKEN_REVIEW_JWT" kubernetes_host="$KUBE_HOST" kubernetes_ca_cert="$KUBE_CA_CERT" disable_local_ca_jwt="true"
```

Create a named role - `devweb-app` & associate it with the account

```shell
vault write auth/kubernetes/role/devweb-app \
  bound_service_account_names=vault-auth \
  bound_service_account_namespaces=default \
  ttl=24h
```

You will be able to authenticate and get a client token via:
```shell
vault write auth/kubernetes/login role=devweb-app jwt=$TOKEN_REVIEW_JWT
```

Useful Links:
- https://www.vaultproject.io/docs/auth/kubernetes
- https://support.hashicorp.com/hc/en-us/articles/4404389946387-Kubernetes-auth-method-Permission-Denied-error#:~:text=This%20error%20message%20is%20usually,auth%20is%20not%20configured%20properly
- 