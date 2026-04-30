import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { Config } from '../config/config.js';
import { getClusterTools, ToolDefinition } from './cluster.js';
import { getClusterToolkitTools } from './clustertoolkit.js';
import { getDeployTools } from './deploy.js';
import { getGkeReleaseNotesTools } from './gkereleasenotes.js';
import { getK8sChangelogTools } from './k8schangelog.js';
import { getGiqTools } from './giq.js';
import { getLoggingTools } from './logging.js';
import { getMonitoringTools } from './monitoring.js';
import { getRecommendationTools } from './recommendation.js';
import { getManifestGenTools } from '../agents/manifestgen.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

export function installTools(server: Server, config: Config) {
  const tools: ToolDefinition[] = [
    ...getClusterTools(config),
    ...getClusterToolkitTools(config),
    ...getDeployTools(config),
    ...getGkeReleaseNotesTools(config),
    ...getK8sChangelogTools(config),
    ...getGiqTools(config),
    ...getLoggingTools(config),
    ...getMonitoringTools(config),
    ...getRecommendationTools(config),
    ...getManifestGenTools(config),
  ];

  server.setRequestHandler(ListToolsRequestSchema, async () => {
    return {
      tools: tools.map(t => ({
        name: t.name,
        description: t.description,
        inputSchema: t.inputSchema,
      })),
    };
  });

  server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const tool = tools.find(t => t.name === request.params.name);
    if (!tool) {
      throw new Error(`Tool not found: ${request.params.name}`);
    }
    
    try {
      return await tool.handler(request.params.arguments);
    } catch (error: any) {
      return {
        isError: true,
        content: [{ type: 'text', text: error.message || String(error) }],
      };
    }
  });
}
