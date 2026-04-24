---
name: k8s-event-responder-agent
description: 'Respond to Kubernetes events, perform automated diagnostic analysis, and provide summaries for GKE cluster operators. Use when: (1) Investigating pod failures, (2) Analyzing K8s event logs, or (3) Triaging cluster alerts.'
---

# K8s Event Responder Agent

## Overview

This skill enables an agent to monitor, analyze, and respond to Kubernetes events in GKE clusters.

## Agent Launch

To start the continuous event monitor, spawn an isolated sub-agent using `sessions_spawn` (e.g. `runtime="subagent"`, `mode="session"`). Instruct the sub-agent to execute `scripts/event_monitor.py`. Because the event monitor is a long-running process, the sub-agent will need to run it in the background and parse its output interactively while it is running. The sub-agent should proactively message the user about new events as they occur. When launching the agent, ask the user to pick a delivery method for these messages (e.g., routing back to the main session via `sessions_send`, or sending to a specific channel).

## Workflow

When an event is detected:
1. **Analyze**: Use [analyze_event.py](scripts/analyze_event.py) to fetch context (logs, recent changes, status).
2. **Interpret**: Consult [k8s-events.md](references/k8s-events.md) to classify the event urgency and suggested remediation.
3. **Report**: Summarize findings and recommended actions for the operator.

## How to use

- "Analyze the latest crashloopbackoff event in the `prod` namespace."
- "What's going on with this OOMKill event?"
- "Summarize recent error events in the `web-frontend` deployment."
