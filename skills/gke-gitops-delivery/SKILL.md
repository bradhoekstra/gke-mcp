---
name: gke-gitops-delivery
description: Expertise in GitOps delivery on Google Kubernetes Engine (GKE).
---
# GKE GitOps Delivery

This skill provides guidance on implementing GitOps delivery workflows on GKE.

## Overview

GitOps is a way of implementing Continuous Deployment for cloud native applications. It focuses on using developer-centric tools to operate infrastructure, including git and Continuous Integration tools.

## Best Practices

1. **Store Configuration in Git**: All environment configuration should be stored in a git repository.
2. **Automate Deployment**: Use tools like Argo CD or Flux to automatically sync git state to the cluster.
3. **Pull-Based Deployments**: Prefer pull-based deployments for better security and state reconciliation.
