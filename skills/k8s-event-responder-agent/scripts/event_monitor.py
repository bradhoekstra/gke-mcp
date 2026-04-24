#!/usr/bin/env python3
"""Module for monitoring Kubernetes events continuously."""

import subprocess
import json

def monitor_events(namespace="default"):
    """Watch for specific Kubernetes events and trigger diagnostics."""
    print(f"Starting event monitor in namespace: {namespace}...")
    cmd = ["kubectl", "get", "events", "-n", namespace, "--watch", "--output-watch-events", "-o", "json"]

    with subprocess.Popen(cmd, stdout=subprocess.PIPE, text=True) as process:
        try:
            for line in process.stdout:
                line = line.strip()
                if not line:
                    continue
                watch_event = json.loads(line)
                event = watch_event.get("object", {})
                if event.get("type") == "Warning":
                    reason = event.get("reason")
                    message = event.get("message")
                    involved_object = event.get("involvedObject", {}).get("name")
                    print(f"EVENT_DETECTED:{reason}:{involved_object}:{message}")
                    # analysis = subprocess.check_output(["python3", "/home/bhoekstra_google_com/openclaw/skills/k8s-event-responder-agent/scripts/analyze_event.py", involved_object, namespace], text=True)
                    # print(f"ANALYSIS_RESULT:{analysis.strip()}")
        except KeyboardInterrupt:
            process.terminate()
            print("Monitor stopped.")

if __name__ == "__main__":
    monitor_events()
