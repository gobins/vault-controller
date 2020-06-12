# vault-controller
A K8s controller to manage Hashicorp Vault configuration using CRDs.

## Deploy
```
kubectl apply -f https://raw.githubusercontent.com/gobins/vault-controller/master/config/deploy.yaml
```

### Configuration
To enable the controller to talk to vault API, create a configmap.
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  namespace: vault-controller-system
data:
  address: http://10.244.0.6:8200
  token: root
```
### SysAuth
```
apiVersion: vault.gobins.github.io/v1
kind: SysAuth
metadata:
  name: sysauth-sample
  namespace: vault-controller-system
spec:
  path: "testapprole"
  description: "testing"
  type: "approle"
```

### Policy
```
apiVersion: vault.gobins.github.io/v1
kind: Policy
metadata:
  name: policy-sample
  namespace: vault-controller-system
spec:
  name: testpolicy
  rules: |
    path "user-kv/data/{{identity.entity.name}}/*" {
        capabilities = [ "create", "update", "read", "delete", "list" ]
    }
    path "user-kv/metadata" {
      capabilities = ["list"]
    }
```

### Todo
- [ ] Add other authentication for vault client
- [ ] Add webhook for validation
- [ ] Add CRDs for auth methods(Approle, AWS, Tokens, Google Cloud)