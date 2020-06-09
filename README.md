# vault-controller
A K8s controller to manage Hashicorp Vault configuration using CRDs.


### Configuration
To enable the controller to talk to vault API, create a configmap.
```
apiVersion: vault.gobins.github.io/v1
kind: Policy
metadata:
  name: policy-sample
  namespace: vault-controller-system
spec:
  # Add fields here
  name: testpolicy
  rules: |
    # Grant permissions on user specific path
    path "user-kv/data/{{identity.entity.name}}/*" {
        capabilities = [ "create", "update", "read", "delete", "list" ]
    }

    # For Web UI usage
    path "user-kv/metadata" {
      capabilities = ["list"]
    }
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