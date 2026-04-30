import { Config } from '../config/config.js';
import { exec as execCallback } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';
const exec = promisify(execCallback);

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export function getClusterToolkitTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'cluster_toolkit_download',
      description: 'Cluster Toolkit, is open-source software offered by Google Cloud which simplifies the process for you to create Google Kubernetes Engine clusters and deploy high performance computing (HPC), artificial intelligence (AI), and machine learning (ML). It is designed to be highly customizable and extensible, and intends to address the deployment needs of a broad range of use cases. This tool will download the public git repository so that Cluster Toolkit can be used.',
      inputSchema: {
        type: 'object',
        properties: {
          download_directory: { type: 'string', description: "Download directory for the git repo. By default use the absolute path to the current working directory." },
        },
        required: ['download_directory'],
      },
      handler: async (args: any) => {
        if (!args.download_directory) {
          throw new Error('download_directory argument cannot be empty');
        }
        
        let downloadDir = args.download_directory;
        if (!downloadDir.endsWith('cluster-toolkit')) {
          downloadDir = path.join(downloadDir, 'cluster-toolkit');
        }
        
        try {
          const { stdout, stderr } = await exec(`git clone https://github.com/GoogleCloudPlatform/cluster-toolkit.git ${downloadDir}`);
          return {
            content: [
              { type: 'text', text: stdout || stderr || 'Successfully cloned cluster-toolkit' },
            ],
          };
        } catch (error: any) {
          console.error(`Failed to download Cluster Toolkit: ${error}`);
          throw error;
        }
      },
    },
  ];
}
