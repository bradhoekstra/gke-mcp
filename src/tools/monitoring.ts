import { Config } from '../config/config.js';
import { MetricServiceClient } from '@google-cloud/monitoring';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export function getMonitoringTools(config: Config): ToolDefinition[] {
  const client = new MetricServiceClient();

  return [
    {
      name: 'list_monitored_resource_descriptors',
      description: 'List monitored resource descriptors(schema) related to GKE for this project. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
        },
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        if (!projectId) {
          throw new Error('project_id argument cannot be empty');
        }

        const name = `projects/${projectId}`;
        
        try {
          const [descriptors] = await client.listMonitoredResourceDescriptors({ name });
          
          let result = '';
          for (const d of descriptors) {
            result += JSON.stringify(d, null, 2) + '\n';
          }
          
          return {
            content: [{ type: 'text', text: result }],
          };
        } catch (error: any) {
          console.error(`Failed to list monitored resource descriptors: ${error}`);
          throw error;
        }
      },
    },
  ];
}
