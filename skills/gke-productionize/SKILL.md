---
name: gke-productionize
description: Assists in preparing applications and clusters on GKE for production.
---

# GKE Productionize Skill

This skill acts as a high-level orchestrator for preparing a GKE cluster and its workloads for production readiness. It covers discovery, assessment, and references specialized skills for detailed implementation.

## Scope

This skill is adaptable to:

- A single application (already on Kubernetes or not).
- A set of applications.
- A target cluster.

## Workflow

### 0. Pre-Kubernetes Phase (If applicable)

If the application has not been deployed to GKE or Kubernetes yet, and may lack a container image or YAML files, follow these steps:

#### App Assessment

- Identify the language, framework, and dependencies.
- Determine configuration methods (e.g., environment variables).
- Identify stateful needs (databases, file storage).
- Determine port mapping and protocol (HTTP, TCP, etc.).

#### Containerization

- Create a `Dockerfile` suitable for the application.
- Recommend multi-stage builds for smaller, more secure images.
- Ensure the app logs to `stdout`/`stderr`.

#### Image Management

- Build the container image.
- Push the image to Google Artifact Registry.

#### Manifest Generation

- Generate a basic `Deployment` manifest with resource requests/limits and health probes.
- Generate a `Service` manifest (ClusterIP or LoadBalancer as needed).

#### Initial Deployment

- Apply the manifests to the target GKE cluster.
- Verify the app is running and accessible.

Once deployed, proceed to the standard **Discovery Phase** and **Production Readiness Checklist**.

### 1. Discovery Phase

Before making recommendations, you must discover the current state of the environment.

#### Cluster Discovery

Run these commands to understand the cluster setup:

- Check cluster details: `gcloud container clusters describe <cluster-name> --zone <zone> --project <project>` (or `--region`).
- Check for Autopilot vs Standard: Look for `autopilot: true` in the describe output.
- Check release channel: Look for `releaseChannel`.

#### Workload Discovery

If a specific application is targeted, discover its configuration:

- Get deployment/statefulset details: `kubectl get deployment <app-name> -n <namespace> -o yaml`
- Check for resource requests and limits.
- Check for liveness, readiness, and startup probes.
- Check for HPA: `kubectl get hpa -n <namespace>`
- Check for PDB: `kubectl get pdb -n <namespace>`
- Check for NetworkPolicies: `kubectl get networkpolicy -n <namespace>`

### 2. Production Readiness Checklist

Go through these areas and assess readiness.

#### A. Scalability & Resource Management

- **Resource Requests & Limits**: Ensure all production containers have explicit CPU and memory requests and limits.
- **Horizontal Pod Autoscaling (HPA)**: Recommend HPA for workloads that experience variable load.
- **Vertical Pod Autoscaling (VPA)**: Recommend VPA for workloads that are hard to size or for optimizing requests.
- **Cluster Autoscaler**: Ensure Cluster Autoscaler is enabled (automatic in Autopilot).
- **ComputeClasses (Autopilot)**: Recommend appropriate ComputeClasses if applicable.

> [!NOTE]
> For detailed implementation of scaling strategies, refer to the [gke-workload-scaling](../gke-workload-scaling/SKILL.md) skill.

#### B. Observability

- **Logging**: Verify Cloud Logging is enabled. Check if workloads are logging to stdout/stderr.
- **Monitoring**: Verify Cloud Monitoring is enabled.
- **Managed Service for Prometheus**: Recommend setting up Managed Service for Prometheus for application metrics.
- **Dashboards & Alerts**: Suggest creating dashboards for key metrics (CPU, Memory, Latency) and setting up alerts.

#### C. Reliability

- **Multi-Zonal/Regional**: Recommend regional clusters for high availability.
- **Pod Disruption Budgets (PDB)**: Ensure production workloads have PDBs configured to prevent downtime during maintenance.
- **Health Probes**: Ensure Liveness and Readiness probes are configured correctly. Startup probes should be used for slow-starting apps.
- **Termination Grace Period**: Ensure appropriate `terminationGracePeriodSeconds` is set for graceful shutdown.

#### D. Security

- **Workload Identity**: Recommend using Workload Identity instead of service account keys for accessing Google Cloud APIs.
- **Namespace Isolation**: Ensure workloads are separated by namespaces.
- **Network Policies**: Recommend using Network Policies to restrict traffic between pods.
- **Pod Security Admission**: Ensure appropriate security standards are enforced (e.g., baseline or restricted).
- **Secrets Management**: Recommend using **Google Cloud Secret Manager** integrated via the Secret Provider Class (CSI driver) or External Secrets Operator.
- **Cluster Network Security**: Recommend **Private Clusters**, **Master Authorized Networks**, and **Cloud NAT** for outbound internet access.
- **Vulnerability Scanning**: Recommend enabling vulnerability scanning in Artifact Registry and policies to block high-risk images.

> [!NOTE]
> For detailed implementation of workload security and network policies, refer to the [gke-workload-security](../gke-workload-security/SKILL.md) skill.

#### E. Backup & Disaster Recovery

- **Backup for GKE**: Recommend enabling and configuring Backup for GKE for stateful workloads.
- **Disaster Recovery**: Ensure backups are tested and recovery procedures are documented.

#### F. Cost Optimization

- **Autopilot / Rightsizing**: Leverage Autopilot for automatic rightsizing or use VPA for Standard clusters.
- **Spot VMs**: Consider Spot VMs for fault-tolerant or non-critical workloads.
- **Resource Profiling**: Regularly review resource requests and limits to avoid over-provisioning.

#### G. Edge Security & Ingress

- **Gateway API / Ingress**: Recommend Gateway API or GKE Ingress for exposing services.
- **Cloud Armor**: Recommend using Cloud Armor for WAF and DDoS protection.
- **SSL Certificates**: Use Google-managed SSL certificates for HTTPS.

#### H. Deployment & GitOps

- **GitOps Practice**: Discourage direct `kubectl apply` for production. Recommend GitOps tools like **Config Sync** or ArgoCD.


## Adaptability Guidelines

- **Single App**: Focus on Health Probes, HPA, Resource Limits, PDB, and Workload Identity for that specific app.
- **Cluster Wide**: Focus on Cluster Autoscaler, Multi-zonal setup, Release Channels, Maintenance Windows, and default Network Policies.
- **Interactive Approach**: Always ask the user for confirmation or missing info (e.g., "What are the typical traffic patterns for this app?" to recommend HPA settings).
