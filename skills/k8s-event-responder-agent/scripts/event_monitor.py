#!/usr/bin/env python3
import time
import subprocess
import json

# Example event monitor script for Kubernetes
# This script watches for specific events and triggers diagnostics.

def monitor_events(namespace="default"):
    print(f"Starting event monitor in namespace: {namespace}...")
    
    # Watch events using kubectl
    # Note: This is a long-running process
    cmd = ["kubectl", "get", "events", "-n", namespace, "--watch", "-o", "json"]
    
    process = subprocess.Popen(cmd, stdout=subprocess.PIPE, text=True)
    
    try:
        for line in process.stdout:
            event = json.loads(line)
            # Example logic: filter for failures
            if event.get("type") == "Warning":
                reason = event.get("reason")
                message = event.get("message")
                involved_object = event.get("involvedObject", {}).get("name")
                
                print(f"Event Detected: {reason} on {involved_object}")
                print(f"Message: {message}")
                
                # Here, you would trigger the analyze-event.py script
                # subprocess.run(["./scripts/analyze-event.py", reason, namespace])
                
    except KeyboardInterrupt:
        process.terminate()
        print("Monitor stopped.")

if __name__ == "__main__":
    monitor_events()
