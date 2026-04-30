import { Config } from '../config/config.js';
import { RecommenderClient } from '@google-cloud/recommender';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export function getRecommendationTools(config: Config): ToolDefinition[] {
  const client = new RecommenderClient();

  return [
    {
      name: 'list_recommendations',
      description: 'List recommendations for GKE. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          location: { type: 'string', description: "GKE cluster location. Leave this empty if the user doesn't doesn't provide it." },
        },
        required: ['location'],
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        if (!projectId) {
          throw new Error('project_id argument cannot be empty');
        }
        if (!args.location) {
          throw new Error('location argument not set');
        }

        const parent = `projects/${projectId}/locations/${args.location}/recommenders/google.container.DiagnosisRecommender`;
        
        try {
          const [recommendations] = await client.listRecommendations({ parent });
          
          let result = '';
          for (const r of recommendations) {
            result += JSON.stringify(r, null, 2) + '\n';
          }
          
          return {
            content: [{ type: 'text', text: result }],
          };
        } catch (error: any) {
          console.error(`Failed to list recommendations: ${error}`);
          throw error;
        }
      },
    },
  ];
}
