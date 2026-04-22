---
name: k8s-event-responder-agent
description: 'Respond to Kubernetes events, perform automated diagnostic analysis, and provide summaries for GKE cluster operators. Use when: (1) Investigating pod failures, (2) Analyzing K8s event logs, or (3) Triaging cluster alerts.'
---

# K8s Event Responder Agent

## Overview

This skill enables an agent to monitor, analyze, and respond to Kubernetes events in GKE clusters.

## Workflow

When an event is detected:
1. **Analyze**: Use [analyze-event.py](scripts/analyze-event.py) to fetch context (logs, recent changes, status).
2. **Interpret**: Consult [k8s-events.md](references/k8s-events.md) to classify the event urgency and suggested remediation.
3. **Report**: Summarize findings and recommended actions for the operator.

## How to use

- "Analyze the latest crashloopbackoff event in the `prod` namespace."
- "What's going on with this OOMKill event?"
- "Summarize recent error events in the `web-frontend` deployment."
