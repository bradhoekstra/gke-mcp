import { LlmAgent } from '@google/adk/dist/esm/agents/llm_agent.js';
import { FunctionTool } from '@google/adk/dist/esm/tools/function_tool.js';
import { Runner } from '@google/adk/dist/esm/runner/runner.js';
import { InMemorySessionService } from '@google/adk/dist/esm/sessions/in_memory_session_service.js';
import { z } from 'zod';
import { Config } from '../config/config.js';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { exec as execCallback } from 'child_process';
import { promisify } from 'util';
import { randomUUID } from 'crypto';

const exec = promisify(execCallback);

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const defaultModel = "gemini-2.5-pro";

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export class Agent {
    private adkRunner: Runner;
    private sessionService: InMemorySessionService;

    constructor(config: Config) {
        const instructionPath = path.join(__dirname, 'instruction.md');
        let instructionTemplate = '';
        try {
            instructionTemplate = fs.readFileSync(instructionPath, 'utf8');
        } catch (e) {
            console.warn(`Warning: Could not read instruction.md at ${instructionPath}`);
            instructionTemplate = "You are a helpful AI assistant specializing in Kubernetes.";
        }

        const giqTool = new FunctionTool({
            name: "giq_generate_manifest",
            description: "Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud",
            parameters: z.object({
                model: z.string().describe("The model to use."),
                model_server: z.string().describe("The model server to use."),
                accelerator: z.string().describe("The accelerator to use."),
                target_ntpot_milliseconds: z.string().optional().describe("The maximum normalized time per output token (NTPOT) in milliseconds."),
            }),
            execute: async (args: any) => {
                try {
                    const { stdout } = await exec(`gcloud container recommender generate-optimized-manifest --model=${args.model} --model-server=${args.model_server} --accelerator=${args.accelerator}`);
                    return stdout;
                } catch (error: any) {
                    console.error(`Failed to generate manifest via gcloud: ${error}`);
                    throw error;
                }
            }
        });

        const adkAgent = new LlmAgent({
            name: "manifest_agent",
            description: "Agent specialized in generating and validating Kubernetes manifests.",
            model: defaultModel,
            instruction: instructionTemplate,
            tools: [giqTool],
        });

        this.sessionService = new InMemorySessionService();

        this.adkRunner = new Runner({
            appName: "gke-mcp",
            agent: adkAgent,
            sessionService: this.sessionService,
        });
    }

    async run(prompt: string, sessionId: string): Promise<string> {
        try {
            await this.sessionService.getSession({
                appName: "gke-mcp",
                userId: "default-user",
                sessionId: sessionId,
            });
        } catch (e) {
            await this.sessionService.createSession({
                appName: "gke-mcp",
                userId: "default-user",
                sessionId: sessionId,
            });
        }

        const msg = {
            role: "user",
            parts: [{ text: prompt }],
        };

        const events = this.adkRunner.runAsync({
            userId: "default-user",
            sessionId: sessionId,
            newMessage: msg,
        });

        let builder = '';
        for await (const event of events) {
            if (event.content) {
                for (const part of event.content.parts) {
                    if (part.text) {
                        builder += part.text;
                    }
                }
            }
        }

        return builder;
    }
}

export function getManifestGenTools(config: Config): ToolDefinition[] {
    const agent = new Agent(config);

    return [
        {
            name: "generate_manifest",
            description: "Generates a Kubernetes manifest using Vertex AI based on a description.",
            inputSchema: {
                type: 'object',
                properties: {
                    prompt: { type: 'string', description: "The description of the manifest to generate. e.g. 'nginx deployment with 3 replicas'" },
                    session_id: { type: 'string', description: "Optional. A unique identifier to maintain conversation history across multiple tool calls. If not provided, a new random ID will be generated." },
                },
                required: ['prompt'],
            },
            handler: async (args: any) => {
                const sessID = args.session_id || randomUUID();
                const manifest = await agent.run(args.prompt, sessID);
                return {
                    content: [{ type: 'text', text: manifest }],
                };
            }
        }
    ];
}
