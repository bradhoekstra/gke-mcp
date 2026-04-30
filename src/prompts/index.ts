import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { Config } from '../config/config.js';
import { getCostPrompts, PromptDefinition } from './cost.js';
import { getDeployPrompts } from './deploy.js';
import { getUpgradeRiskReportPrompts } from './upgraderiskreport.js';
import { getUpgradesBestPracticesRiskReportPrompts } from './upgradesbestpracticesriskreport.js';
import {
  ListPromptsRequestSchema,
  GetPromptRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

export function installPrompts(server: Server, config: Config) {
  const prompts: PromptDefinition[] = [
    ...getCostPrompts(config),
    ...getDeployPrompts(config),
    ...getUpgradeRiskReportPrompts(config),
    ...getUpgradesBestPracticesRiskReportPrompts(config),
  ];

  server.setRequestHandler(ListPromptsRequestSchema, async () => {
    return {
      prompts: prompts.map(p => ({
        name: p.name,
        description: p.description,
        arguments: p.arguments,
      })),
    };
  });

  server.setRequestHandler(GetPromptRequestSchema, async (request) => {
    const prompt = prompts.find(p => p.name === request.params.name);
    if (!prompt) {
      throw new Error(`Prompt not found: ${request.params.name}`);
    }
    
    try {
      return await prompt.handler(request.params.arguments);
    } catch (error: any) {
      throw new Error(`Failed to get prompt: ${error.message || error}`);
    }
  });
}
