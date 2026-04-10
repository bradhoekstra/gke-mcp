---
name: gke-gitops-delivery
description: Workflows for implementing GitOps continuous delivery on GKE using Config Sync or ArgoCD.
---

# GKE GitOps Delivery Skill

This skill provides workflows for adopting GitOps practices for deploying applications and managing cluster configuration on GKE.

## Workflows

### 1. Implement Config Sync (Google Managed)

Config Sync is Google's managed GitOps solution.

**Prerequisites**: Config Sync must be enabled on the cluster.

**Enable Config Sync:**
```bash
gcloud container clusters update <cluster-name> \
    --enable-config-sync \
    --region <region>
```

**Configure RootSync:**
A `RootSync` object tells Config Sync where to pull cluster-wide configuration from.

**Example RootSync:**
```yaml
apiVersion: configsync.gke.io/v1beta1
kind: RootSync
metadata:
  name: root-sync
  namespace: config-management-system
spec:
  sourceFormat: unstructured
  git:
    repo: https://github.com/my-org/my-config-repo
    branch: main
    dir: /cluster-config
    auth: ssh
    secretRef:
      name: git-creds
```

### 2. Implement ArgoCD

ArgoCD is a popular open-source GitOps tool.

**Install ArgoCD:**
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

**Create an Application in ArgoCD:**
An Application resource defines the source repo and the target cluster/namespace.

**Example Application:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/my-org/my-app-repo
    targetRevision: HEAD
    path: manifests
  destination:
    server: https://kubernetes.default.svc
    namespace: my-app-ns
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 3. Multi-Cluster GitOps with Config Sync

Config Sync is ideal for managing multiple clusters from a single repository.

**ClusterSelectors**: You can use `ClusterSelector` to apply resources only to specific clusters based on labels.

**Example ClusterSelector:**
```yaml
apiVersion: configsync.gke.io/v1beta1
kind: ClusterSelector
metadata:
  name: prod-clusters
spec:
  selector:
    matchLabels:
      env: prod
```

Reference it in your resource manifests to restrict where they are applied.

## Best Practices

1. **Single Source of Truth**: The Git repository is the only source of truth. Avoid manual `kubectl` edits in production.
2. **Immutability**: Treat cluster state as immutable. If changes are needed, make them in Git and let the GitOps controller apply them.
3. **Pruning**: Enable pruning (deleting resources in K8s that are no longer in Git) to avoid resource drift.
4. **Automated Sync**: Use automated sync with self-healing to ensure the cluster always matches Git.
5. **Multi-Cluster Consistency**: Use GitOps to ensure consistent configuration across multiple clusters, reducing snowflake environments.

