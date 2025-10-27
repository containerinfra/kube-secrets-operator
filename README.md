# kube-secrets-operator

The Kubernetes Secrets Operator is designed to assist in generating secrets for Kubernetes environments, particularly those driven by GitOps methodologies and seeking reusable Kubernetes resources, like those used with FluxCD.

This tool offers a middle ground for teams that consider an entire Key Management System (KMS) and Secret Management setup to be overly complicated or expensive for the given environment. It acts as an solution between SealedSecrets (manual work required) and cloud-based KMS services (process overhead), such as Vault, making it well-suited for testing purposes, CI/CD operations, and temporary setups.

## Alternative Solutions

There are other alternatives available, such as [mittwald/kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator). While these are viable options, we encountered a specific requirement where `Secret` values needed to be embedded within a configuration file that also need to be a `Secret`.

## Deployment

```yaml
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: OCIRepository
metadata:
  name: kube-secrets-operator-crds
  namespace: kube-system
spec:
  interval: 160m
  url: oci://ghrc.io/containerinfra/charts/kube-secrets-operator-crds
  ref:
    semver: ">= 0.0.0"
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: kube-secrets-operator-crds
  namespace: kube-system
spec:
  chartRef:
    kind: OCIRepository
    name: kube-secrets-operator-crds
    namespace: kube-system
  interval: 1h
  values: {}
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: OCIRepository
metadata:
  name: kube-secrets-operator
  namespace: kube-system
spec:
  interval: 160m
  url: oci://ghrc.io/containerinfra/charts/kube-secrets-operator
  ref:
    semver: ">= 0.0.0"
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: kube-secrets-operator
  namespace: kube-system
spec:
  chartRef:
    kind: OCIRepository
    name: kube-secrets-operator
    namespace: kube-system
  interval: 1h
  install:
    crds: Skip
  values:
    enableServiceMonitor: false
    replicaCount: 1
    affinity:
      nodeAffinity:
          # prefer scheduling on control-plane machines
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              preference:
                matchExpressions:
                  - key: node-role.kubernetes.io/control-plane
                    operator: Exists
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: node-role.kubernetes.io/control-plane
        effect: NoSchedule
```
