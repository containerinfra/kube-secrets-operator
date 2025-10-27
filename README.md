# kube-secrets-operator

The Kubernetes Secrets Operator is designed to assist in generating secrets for Kubernetes environments, particularly those driven by GitOps methodologies and seeking reusable Kubernetes resources, like those used with FluxCD.

This tool offers a middle ground for teams that consider an entire Key Management System (KMS) and Secret Management setup to be overly complicated or expensive for the given environment. It acts as an solution between SealedSecrets (manual work required) and cloud-based KMS services (process overhead), such as Vault, making it well-suited for testing purposes, CI/CD operations, and temporary setups.

## Alternative Solutions

There are other alternatives available, such as [mittwald/kubernetes-secret-generator](https://github.com/mittwald/kubernetes-secret-generator). While these are viable options, we encountered a specific requirement where `Secret` values needed to be embedded within a configuration file that also need to be a `Secret`.
