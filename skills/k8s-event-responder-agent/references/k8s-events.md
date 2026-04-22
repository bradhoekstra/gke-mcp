# Kubernetes Events Reference

This guide helps the agent classify GKE events and recommend appropriate remediation.

## Common Error Patterns

### 1. CrashLoopBackOff
- **Cause**: Application crashing on start, missing config, bad env vars, or resource exhaustion.
- **Agent Action**: 
    - Fetch logs from the last restart (`kubectl logs <pod> --previous`).
    - Inspect pod description for environment variable errors or readiness probe failures.

### 2. OOMKilled
- **Cause**: Application exceeded memory limits.
- **Agent Action**:
    - Check container memory usage vs limit.
    - Review logs for memory leak indicators.
    - Recommend increasing memory limit or investigating memory usage.

### 3. ImagePullBackOff
- **Cause**: Invalid image name, tag, or registry authentication issue.
- **Agent Action**:
    - Verify image name and tag existence.
    - Check for ImagePullSecrets configuration.

### 4. FailedScheduling
- **Cause**: Insufficient cluster resources (CPU/Mem/GPU) or node selector mismatch.
- **Agent Action**:
    - Inspect cluster capacity vs pod requirements.
    - Check for node affinity constraints.

## Urgency Matrix

| Event Type | Priority | Action Type |
| :--- | :--- | :--- |
| OOMKilled | High | Immediate Investigation |
| CrashLoopBackOff | High | Immediate Investigation |
| FailedScheduling | Medium | Cluster Scaling / Capacity |
| ImagePullBackOff | Medium | Deployment Configuration |
