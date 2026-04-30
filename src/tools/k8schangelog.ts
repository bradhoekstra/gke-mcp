import { Config } from '../config/config.js';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

const kubernetesMinorVersionRegexp = /^\d+\.\d+$/;
const changelogHostURL = "https://raw.githubusercontent.com";
const changelogVersionLineRegexp = /^# v\d\.\d+\.\d+/;
const ignoredSectionPrefixes = ["## Dependencies", "## Downloads for"];

export function keepOnlyChanges(changelog: string): string {
  let result = '';
  let hasMetTheFirstVersionHeading = false;
  let isInIgnoredSection = false;
  
  const lines = changelog.split('\n');

  for (const line of lines) {
    if (!hasMetTheFirstVersionHeading) {
      if (changelogVersionLineRegexp.test(line)) {
        hasMetTheFirstVersionHeading = true;
      } else {
        continue;
      }
    }

    let isIgnoredSectionHeader = false;
    for (const prefix of ignoredSectionPrefixes) {
      if (line.startsWith(prefix)) {
        isInIgnoredSection = true;
        isIgnoredSectionHeader = true;
        break;
      }
    }
    if (isIgnoredSectionHeader) {
      continue;
    }

    if (isInIgnoredSection) {
      if (line.startsWith('# ') || line.startsWith('## ')) {
        isInIgnoredSection = false;
      }
    }

    if (!isInIgnoredSection) {
      result += line + '\n';
    }
  }
  return result;
}

export function getK8sChangelogTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'get_k8s_changelog',
      description: 'Get changelog file for a specific kubernetes minor version and keep only changes content. Prefer to use this tool if kubernetes minor version changelog is needed.',
      inputSchema: {
        type: 'object',
        properties: {
          KubernetesMinorVersion: { type: 'string', description: "The kubernetes minor version to get changelog for. For example, '1.33'." },
        },
        required: ['KubernetesMinorVersion'],
      },
      handler: async (args: any) => {
        const version = args.KubernetesMinorVersion.trim();
        if (!kubernetesMinorVersionRegexp.test(version)) {
          throw new Error(`invalid kubernetes minor version: ${version}`);
        }

        const url = `${changelogHostURL}/kubernetes/kubernetes/refs/heads/master/CHANGELOG/CHANGELOG-${version}.md`;
        console.error(`Fetching changelog from: ${url}`);
        
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`failed to get changelog with status code: ${response.status}`);
        }

        const body = await response.text();
        
        return {
          content: [
            { type: 'text', text: keepOnlyChanges(body) },
          ],
        };
      },
    },
  ];
}
