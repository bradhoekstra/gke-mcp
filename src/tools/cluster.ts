import { Config } from '../config/config.js';
import { ClusterManagerClient } from '@google-cloud/container';
import * as k8s from '@kubernetes/client-node';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { exec as execCallback } from 'child_process';
import { promisify } from 'util';
const exec = promisify(execCallback);

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export function getClusterTools(config: Config): ToolDefinition[] {
  const client = new ClusterManagerClient();

  const getNodeSosReportWithPod = async (args: any, remoteTmpDir: string, podName: string): Promise<string> => {
    const overrides = {
      spec: {
        nodeName: args.node,
        hostNetwork: true,
        hostPID: true,
        hostIPC: true,
        containers: [
          {
            name: 'main',
            image: 'gke.gcr.io/debian-base',
            command: ['/bin/sleep', '99999'],
            volumeMounts: [
              {
                mountPath: '/host',
                name: 'root',
              },
            ],
          },
        ],
        volumes: [
          {
            name: 'root',
            hostPath: {
              path: '/',
              type: 'Directory',
            },
          },
        ],
        securityContext: {
          runAsUser: 0,
        },
        nodeSelector: {
          'kubernetes.io/hostname': args.node,
        },
      },
    };

    const overridesStr = JSON.stringify(overrides);
    
    try {
      await exec(`kubectl run ${podName} --image=gke.gcr.io/debian-base --restart=Never --overrides='${overridesStr}'`);
      
      // Wait for pod to be ready
      await exec(`kubectl wait --for=condition=Ready pod/${podName} --timeout=60s`);
      
      // Run sos report
      const execScript = `apt update && apt install -y sosreport && mkdir -p /host${remoteTmpDir} && sos report --sysroot=/host --all-logs --batch --tmp-dir=/host${remoteTmpDir}`;
      const { stdout } = await exec(`kubectl exec ${podName} -- sh -c "${execScript}"`);
      
      // Parse output for filename
      const re = new RegExp(`(/host)?${remoteTmpDir}/[^\\s]+\\.tar\\.(xz|gz)`);
      const match = stdout.match(re);
      if (!match) {
        throw new Error(`Could not find sos report filename in output: ${stdout}`);
      }
      
      let remotePath = match[0];
      if (!remotePath.startsWith('/host')) {
        remotePath = '/host' + remotePath;
      }
      
      const localFilename = `sosreport-${args.node}-${new Date().toISOString().replace(/[:.]/g, '-')}.tar.xz`;
      const localPath = path.join(args.destination, localFilename);
      
      // Copy file
      await exec(`kubectl exec ${podName} -- cat ${remotePath} > ${localPath}`);
      
      return localPath;
    } finally {
      // Cleanup pod
      await exec(`kubectl delete pod ${podName} --wait=false --grace-period=0 --force`).catch(() => {});
    }
  };

  const getNodeSosReportWithSSH = async (args: any): Promise<string> => {
    // Find zone
    const { stdout: zoneOut } = await exec(`gcloud compute instances list --filter="name=${args.node}" --format="value(zone)"`);
    const zone = zoneOut.trim();
    if (!zone) {
      throw new Error(`Could not find zone for node ${args.node}`);
    }
    
    // Generate report
    const { stdout } = await exec(`gcloud compute ssh --zone ${zone} ${args.node} --command "sudo sos report --all-logs --batch --tmp-dir=/var"`);
    
    // Parse filename
    const re = /\/var\/sosreport-[^\s]+\.tar\.xz/;
    const match = stdout.match(re);
    if (!match) {
      throw new Error(`Could not find sos report filename in ssh output: ${stdout}`);
    }
    const remotePath = match[0];
    
    // Chown
    await exec(`gcloud compute ssh --zone ${zone} ${args.node} --command "sudo chown $USER ${remotePath}"`);
    
    const localFilename = `sosreport-${args.node}-${new Date().toISOString().replace(/[:.]/g, '-')}.tar.xz`;
    const localPath = path.join(args.destination, localFilename);
    
    // SCP
    await exec(`gcloud compute scp --zone ${zone} ${args.node}:${remotePath} ${localPath}`);
    
    // Cleanup remote
    await exec(`gcloud compute ssh --zone ${zone} ${args.node} --command "sudo rm ${remotePath}"`).catch(() => {});
    
    return localPath;
  };

  return [
    {
      name: 'list_clusters',
      description: 'List GKE clusters. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          location: { type: 'string', description: "GKE cluster location. Leave this empty if the user doesn't doesn't provide it." },
        },
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const location = args.location || '-';

        const parent = `projects/${projectId}/locations/${location}`;
        const [response] = await client.listClusters({ parent });

        return {
          content: [
            { type: 'text', text: `Found ${response.clusters?.length || 0} clusters in project ${projectId}:` },
            { type: 'text', text: JSON.stringify(response, null, 2) },
          ],
        };
      },
    },
    {
      name: 'get_cluster',
      description: 'Get / describe a GKE cluster. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          location: { type: 'string', description: "GKE cluster location. Leave this empty if the user doesn't doesn't provide it." },
          name: { type: 'string', description: "GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name." },
        },
        required: ['name'],
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const location = args.location || config.defaultLocation;
        const name = args.name;

        if (!name) {
          throw new Error('name argument cannot be empty');
        }

        const clusterName = `projects/${projectId}/locations/${location}/clusters/${name}`;
        const [response] = await client.getCluster({ name: clusterName });

        return {
          content: [
            { type: 'text', text: JSON.stringify(response, null, 2) },
          ],
        };
      },
    },
    {
      name: 'create_cluster',
      description: 'Create a GKE cluster. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          location: { type: 'string', description: "GKE cluster location. Leave this empty if the user doesn't doesn't provide it." },
          cluster: { type: 'object', description: "GKE cluster configuration." },
        },
        required: ['cluster'],
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const location = args.location || config.defaultLocation;
        const cluster = args.cluster;

        const parent = `projects/${projectId}/locations/${location}`;
        const [response] = await client.createCluster({ parent, cluster });

        return {
          content: [
            { type: 'text', text: JSON.stringify(response, null, 2) },
          ],
        };
      },
    },
    {
      name: 'get_kubeconfig',
      description: 'Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config.',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          location: { type: 'string', description: "GKE cluster location. Leave this empty if the user doesn't provide it." },
          name: { type: 'string', description: "GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name." },
        },
        required: ['name'],
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const location = args.location || config.defaultLocation;
        const name = args.name;

        if (!name) {
          throw new Error('name argument cannot be empty');
        }

        const clusterName = `projects/${projectId}/locations/${location}/clusters/${name}`;
        const [resp] = await client.getCluster({ name: clusterName });

        const clusterCaCertificate = resp.masterAuth?.clusterCaCertificate;
        let endpoint = resp.endpoint;

        if (!clusterCaCertificate) {
          throw new Error(`clusterCaCertificate not found for cluster ${name}`);
        }
        if (!endpoint) {
          throw new Error(`endpoint not found for cluster ${name}`);
        }

        if (!endpoint.startsWith('https://')) {
          endpoint = 'https://' + endpoint;
        }

        const newClusterName = `gke_${projectId}_${location}_${name}`;

        const kc = new k8s.KubeConfig();
        try {
          kc.loadFromDefault();
        } catch (e) {
          // If no default config exists, start with a blank one
        }

        kc.addCluster({
          name: newClusterName,
          server: endpoint,
          caData: clusterCaCertificate,
          skipTLSVerify: false,
        });

        kc.addUser({
          name: newClusterName,
          exec: {
            apiVersion: 'client.authentication.k8s.io/v1beta1',
            command: 'gke-gcloud-auth-plugin',
            provideClusterInfo: true,
          }
        });

        kc.addContext({
          name: newClusterName,
          cluster: newClusterName,
          user: newClusterName,
        });

        kc.setCurrentContext(newClusterName);

        const kubeconfigPath = path.join(os.homedir(), '.kube', 'config');
        fs.mkdirSync(path.dirname(kubeconfigPath), { recursive: true });
        fs.writeFileSync(kubeconfigPath, kc.exportConfig());

        return {
          content: [
            { type: 'text', text: `Kubeconfig for cluster ${name} (Project: ${projectId}, Location: ${location}) successfully appended/updated. Current context set to ${newClusterName}.` },
          ],
        };
      },
    },
    {
      name: 'get_node_sos_report',
      description: 'Generate and download an SOS report from a GKE node. Can use \'pod\', \'ssh\' or \'any\' methods. Defaults to \'any\' (pod with fallback to ssh). Use \'ssh\' if node is API-unhealthy.',
      inputSchema: {
        type: 'object',
        properties: {
          node: { type: 'string', description: "GKE node name to collect SOS report from." },
          destination: { type: 'string', description: "Local directory to download the SOS report to. Defaults to /tmp/sos-report if not specified." },
          method: { type: 'string', description: "Method to get sos report. Can be 'pod', 'ssh' or 'any'. Defaults to 'any'. When the node is unhealthy from api server, use ssh only." },
          timeout: { type: 'integer', description: "Timeout in seconds for the report collection. Defaults to 180 (3 minutes)." },
        },
        required: ['node'],
      },
      handler: async (args: any) => {
        if (!args.node) {
          throw new Error('node argument cannot be empty');
        }
        if (!/^[a-z0-9][a-z0-9-.]*[a-z0-9]$/.test(args.node)) {
          throw new Error(`invalid node name: ${args.node}`);
        }

        args.destination = args.destination || '/tmp/sos-report';
        args.method = args.method || 'any';
        args.timeout = args.timeout || 180;

        fs.mkdirSync(args.destination, { recursive: true });

        // Check if node is healthy
        let isHealthy = false;
        try {
          const { stdout } = await exec(`kubectl get node ${args.node} -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'`);
          if (stdout.includes('True')) {
            isHealthy = true;
          }
        } catch (e) {
          // Ignore error, assume unhealthy or inaccessible
        }

        if (!isHealthy) {
          args.method = 'ssh';
        }

        const remoteTmpDir = `/tmp/sos-${args.node}-${Date.now()}`;
        const podName = `sos-debug-${Date.now()}`;

        if (args.method === 'pod' || args.method === 'any') {
          try {
            const localPath = await getNodeSosReportWithPod(args, remoteTmpDir, podName);
            return {
              content: [{ type: 'text', text: `SOS report successfully generated and downloaded to: ${localPath}` }],
            };
          } catch (e) {
            if (args.method === 'pod') {
              throw e;
            }
            // Fallback to SSH if method is 'any'
            console.error(`Pod method failed, falling back to SSH: ${e}`);
          }
        }

        try {
          const localPath = await getNodeSosReportWithSSH(args);
          return {
            content: [{ type: 'text', text: `SOS report successfully generated (via SSH) and downloaded to: ${localPath}` }],
          };
        } catch (e) {
          throw new Error(`Failed to get sos report: ${e}`);
        }
      },
    },
  ];
}
