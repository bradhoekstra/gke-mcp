---
name: gke-cost-optimization
description: Guidance on optimizing costs for Google Kubernetes Engine (GKE) clusters.
---

# GKE Cost Optimization

This skill provides guidance on optimizing costs for Google Kubernetes Engine (GKE) clusters.

## Overview

Cost optimization in GKE involves tracking costs, setting limits to prevent waste, and rightsizing workloads to match actual usage.

## Workflows

### 1. Diagnosis: Identifying Cost Drivers

Before optimizing, understand where costs are coming from.

- **Check Recommendations**: Use the `list_recommendations` MCP tool to find cost optimization opportunities.
- **Analyze Usage**: Use Cloud Billing reports with GKE Cost Allocation enabled to group costs by namespace and labels.
- **Check for Waste**: Look for idle resources, over-provisioned Pods, and unused persistent volumes.

### 2. Enable GKE Cost Allocation

GKE cost allocation allows you to see the cost of your GKE resources in Cloud Billing, broken down by namespace and cluster labels.

**Command:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-cost-allocation \
    --region <region>
```

_Note: You can also use the `update_cluster` MCP tool with appropriate JSON payload._

### 3. Configure Resource Quotas

Resource quotas restrict the total resource consumption in a namespace, preventing any single tenant from consuming all cluster resources.

**Example ResourceQuota Manifest:**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: my-namespace
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 16Gi
    limits.cpu: "8"
    limits.memory: 32Gi
```

### 4. Rightsizing Strategies

Rightsizing involves adjusting the requested resources of your workloads to match their actual utilization.

- **Use VPA in Recommender Mode**: Let VPA observe usage and recommend CPU and memory requests.
- **Use Autopilot**: Autopilot automatically rightsizes Pods based on requested resources and manages node provisioning efficiently.

### 5. Advanced Spot VM usage with ComputeClasses

For Standard clusters, use `ComputeClass` to manage Spot VMs with fallback to on-demand VMs and active migration.

- **Spot Fallback**: Configure priorities in `ComputeClass` to prefer Spot VMs but fallback to On-Demand if Spot is unavailable.
- **Active Migration**: Set `spec.activeMigration.optimizeRulePriority: true` to move workloads back to Spot VMs when available.

Refer to `gke-compute-class-creator` skill for details.

## Best Practices

1. **Enable Cost Allocation**: Always enable GKE cost allocation to understand where your money is going.
2. **Use Resource Quotas**: Enforce resource quotas in multi-tenant clusters to prevent cost runaways.
3. **Leverage Spot VMs**: Use Spot VMs for fault-tolerant, stateless workloads. Use `ComputeClass` for better management.
4. **Automate Scaling**: Use Cluster Autoscaler and HPA/VPA or switch to Autopilot.
5. **Review Recommendations**: Regularly check GKE cost optimization recommendations using `list_recommendations`.
