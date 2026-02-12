---
title: Managed Kubernetes Cluster Configuration
---

# Managed Kubernetes Cluster Configuration

## Problem Description

Managed Kubernetes clusters like AKS (Azure Kubernetes Service), EKS (Amazon Elastic Kubernetes Service), etc., typically use `exec` plugins in their default kubeconfig to dynamically obtain authentication credentials. For example:

- **AKS** uses the `kubelogin` command
- **EKS** uses the `aws` CLI
- **GKE** uses the `gcloud` command
- **GitLab Agent** uses the `glab` command

This authentication method works well in local client environments, but can be challenging in server-side environments like Kube Sentinel because:

1. These CLI tools may not be installed on the server
2. Even if installed, the server environment may not have the corresponding authentication configuration
3. Managing different user credentials in multi-tenant scenarios is difficult

Kube Sentinel provides two ways to solve this: **Managed Authentication Support** (for AWS and GitLab) and **Service Account Tokens** (for all others).

## Managed Authentication Support [NEW]

Kube Sentinel natively supports authentication for specific managed Kubernetes providers by securely managing your credentials and injecting them into the CLI tools.

### AWS EKS Authentication

For EKS clusters, Kube Sentinel supports authentication via `aws` or `aws-iam-authenticator`.

1. **Configure AWS Credentials**: Navigate to **Settings > AWS Settings** and paste your AWS credentials file content (typically found at `~/.aws/credentials`).
2. **Add Cluster**: Import your EKS kubeconfig. Kube Sentinel will detect the `aws` exec command.
3. **Secure Injection**: The system automatically injects the `AWS_SHARED_CREDENTIALS_FILE` environment variable for your requests, ensuring you only use your own credentials.

### GitLab Agent Authentication

For clusters managed via GitLab Agent, Kube Sentinel supports authentication using the `glab` CLI.

1. **Configure GitLab Token**: Navigate to **Settings > GitLab Settings**, add your GitLab host (e.g., `gitlab.com`), and provide a Personal Access Token (PAT).
2. **Validate**: Click **Validate** to initialize your `glab` session.
3. **Add Cluster**: Import your cluster kubeconfig that uses `glab` for authentication.
4. **Context Management**: Kube Sentinel automatically manages the `GLAB_CONFIG_DIR` to use your validated session.

> [!NOTE]
> Support for other managed providers like **AKS (Azure)** and **GKE (Google)** is coming soon.

---

## Alternative: Using Service Account Token

If your provider is not yet natively supported, or if you prefer a more generic approach, you can create a dedicated Service Account for Kube Sentinel and use its token for authentication.

Kube Sentinel provides a helper script for creation:

```sh
wget https://raw.githubusercontent.com/pixelvide/kube-sentinel/refs/heads/main/scripts/generate-kube-sentinel-kubeconfig.sh -O generate-kube-sentinel-kubeconfig.sh
chmod +x generate-kube-sentinel-kubeconfig.sh
./generate-kube-sentinel-kubeconfig.sh
```

### Manual Steps:

1. **Create Service Account and RBAC permissions**:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-sentinel-admin
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-sentinel-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: kube-sentinel-admin
    namespace: kube-system
```

2. **Create Long-lived Token Secret (Kubernetes 1.24+)**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kube-sentinel-admin-token
  namespace: kube-system
  annotations:
    kubernetes.io/service-account.name: kube-sentinel-admin
type: kubernetes.io/service-account-token
```

3. **Get token and cluster information**:

```bash
# Get token
TOKEN=$(kubectl get secret kube-sentinel-admin-token -n kube-system -o jsonpath='{.data.token}' | base64 -d)

# Get CA certificate
CA_CERT=$(kubectl get secret kube-sentinel-admin-token -n kube-system -o jsonpath='{.data.ca\.crt}')

# Get API Server address
API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
```

4. **Generate kubeconfig**:

```bash
cat > kube-sentinel-kubeconfig.yaml <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: ${CA_CERT}
    server: ${API_SERVER}
  name: kube-sentinel-cluster
contexts:
- context:
    cluster: kube-sentinel-cluster
    user: kube-sentinel-admin
  name: kube-sentinel-context
current-context: kube-sentinel-context
users:
- name: kube-sentinel-admin
  user:
    token: ${TOKEN}
EOF
```

## Related Documentation

- [Kubernetes Service Account Tokens](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/)
- [AKS Authentication](https://learn.microsoft.com/en-us/azure/aks/control-kubeconfig-access)
- [EKS Authentication](https://docs.aws.amazon.com/eks/latest/userguide/cluster-auth.html)
- [GKE Authentication](https://cloud.google.com/kubernetes-engine/docs/how-to/api-server-authentication)
