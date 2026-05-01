---
name: gke-reliability
description: Workflows for ensuring high availability and reliability of GKE workloads.
---

# GKE Reliability Skill

This skill provides workflows for configuring your GKE cluster and workloads for high availability and reliability.

## Workflows

### 1. Verify Cluster High Availability

Check if the cluster is regional or has multi-zonal node pools.

**MCP Tool:**

Use the `get_cluster` tool.

Arguments:

- `project_id`: `<project_id>`
- `location`: `<location>` (Region or Zone)
- `cluster_name`: `<cluster_name>`

If the response `location` is a region (e.g., `us-central1`), the control plane is regional.
If `nodePools.locations` (or `locations` in the response) has multiple entries, nodes are spread across multiple zones.

### 2. Configure Pod Disruption Budgets (PDB)

PDBs ensure that a minimum number of pods are available during voluntary disruptions (like node upgrades).

**MCP Tool:**

Use the `get_k8s_resource` tool to check existing PDBs.

Arguments:

- `project_id`: `<project_id>`
- `location`: `<location>`
- `cluster_name`: `<cluster_name>`
- `resourceType`: `poddisruptionbudgets`
- `namespace`: `<namespace>`

**Example Manifest:**

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: my-app-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: my-app
```

### 3. Configure Health Probes

Ensure all production containers have Liveness, Readiness, and optionally Startup probes.

- **Readiness Probe**: Determines when a container is ready to start accepting traffic.
- **Liveness Probe**: Determines when to restart a container.
- **Startup Probe**: Disables liveness and readiness checks until the app has started up.

**MCP Tool:**

Use the `get_k8s_resource` tool to get the deployment configuration and inspect probes.

Arguments:

- `project_id`: `<project_id>`
- `location`: `<location>`
- `cluster_name`: `<cluster_name>`
- `resourceType`: `deployments`
- `name`: `<app-name>`
- `namespace`: `<namespace>`
- `outputFormat`: `YAML`

### 4. Graceful Shutdown

Ensure applications handle `SIGTERM` signals gracefully and have an appropriate `terminationGracePeriodSeconds` set (default is 30s).

### 5. Topology Spread Constraints

Ensure pods are spread across zones or nodes to avoid correlated failures.

**Example Manifest excerpt:**

```yaml
spec:
  topologySpreadConstraints:
    - maxSkew: 1
      topologyKey: topology.kubernetes.io/zone
      whenUnsatisfiable: DoNotSchedule # or ScheduleAnyway
      labelSelector:
        matchLabels:
          app: my-app
```

### 6. Maintenance Windows and Exclusions

Configure when GKE can perform automated upgrades to avoid peak hours.

**Command:**

```bash
gcloud container clusters update <cluster-name> \
    --region <region> \
    --maintenance-window-start <start-time> \
    --maintenance-window-recurrence "FREQ=DAILY"
```

**Alternative (MCP Tool):**

You can use the `update_cluster` tool, but the payload requires a specific JSON structure for `desiredMaintenancePolicy`. If unsure, the `gcloud` command above is a safe fallback.

## Best Practices

1. **Regional Clusters**: Always use regional clusters for production workloads to survive zone failures.
2. **Probes for All Containers**: Every container in a production pod should have at least a readiness probe.
3. **PDBs for Critical Apps**: Use PDBs to prevent downtime during automated node upgrades.
4. **Zone Spreading**: Always use `topologySpreadConstraints` to ensure pods are distributed across zones, even in regional clusters.
5. **Schedule Maintenance**: Set maintenance windows to ensure upgrades happen during low-traffic periods.
