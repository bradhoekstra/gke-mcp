import { Config } from '../config/config.js';
import { exec as execCallback } from 'child_process';
import { promisify } from 'util';
const exec = promisify(execCallback);

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export function getGiqTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'giq_generate_manifest',
      description: 'Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud',
      inputSchema: {
        type: 'object',
        properties: {
          model: { type: 'string', description: "The model to use. Get the list of valid models from 'gcloud container ai profiles models list' if the user doesn't provide it." },
          model_server: { type: 'string', description: "The model server to use. Get the list of valid model servers from 'gcloud container ai profiles list ...' if the user doesn't provide it." },
          accelerator: { type: 'string', description: "The accelerator to use. Get the list of valid accelerators from 'gcloud container ai profiles list ...' if the user doesn't provide it." },
          target_ntpot_milliseconds: { type: 'string', description: "The maximum normalized time per output token (NTPOT) in milliseconds." },
        },
        required: ['model', 'model_server', 'accelerator'],
      },
      handler: async (args: any) => {
        try {
          // Fallback to gcloud command
          const { stdout } = await exec(`gcloud container recommender generate-optimized-manifest --model=${args.model} --model-server=${args.model_server} --accelerator=${args.accelerator}`);
          return {
            content: [{ type: 'text', text: stdout }],
          };
        } catch (error: any) {
          console.error(`Failed to generate manifest via gcloud: ${error}`);
          throw error;
        }
      },
    },
    {
      name: 'giq_fetch_models',
      description: 'Fetch available models for GKE Inference Quickstart.',
      inputSchema: {
        type: 'object',
        properties: {},
      },
      handler: async (args: any) => {
        try {
          const { stdout } = await exec('gcloud container ai profiles models list --format="value(model)"');
          return {
            content: [
              { type: 'text', text: stdout.trim() },
            ],
          };
        } catch (error: any) {
          console.error(`Failed to fetch models: ${error}`);
          throw error;
        }
      },
    },
  ];
}
