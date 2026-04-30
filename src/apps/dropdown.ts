import { Config } from '../config/config.js';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const resourceURI = "ui://dropdown/index.html";
const mimeType = "text/html;profile=mcp-app";

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

export interface ResourceDefinition {
  uri: string;
  name: string;
  mimeType: string;
  handler: () => Promise<any>;
}

export function getDropdownTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'dropdown',
      description: `Renders an interactive UI dropdown for the user to select an item from a list.
Use this tool when you need the user to choose one option from a set of available resources (e.g., clusters, regions, namespaces).
You MUST provide a valid array of 1 or more options. 

Timing: Call this tool immediately before you need the user's input to proceed. Do not ask the user for clarification in plain text; calling this tool serves as your question to the user.
After calling this tool, STOP and wait for the user to make a selection via the UI.
Do NOT list the options in your text response; the UI itself serves as the list and confirmation.`,
      inputSchema: {
        type: 'object',
        properties: {
          title: { type: 'string', description: "Title to display above the dropdown" },
          options: { type: 'array', items: { type: 'string' }, description: "List of resources to display in the dropdown" },
        },
        required: ['options'],
      },
      handler: async (args: any) => {
        if (!args.options || args.options.length === 0) {
          throw new Error("options cannot be empty");
        }

        return {
          structuredContent: {
            status: "PENDING_USER_INPUT",
            options: args.options,
            message: "Present these options to the user. Wait until selection is made",
          },
        };
      },
    },
  ];
}

export function getDropdownResources(config: Config): ResourceDefinition[] {
  return [
    {
      uri: resourceURI,
      name: "GKE Resource Dropdown UI",
      mimeType: mimeType,
      handler: async () => {
        const filePath = path.join(__dirname, '../../ui/dist/apps/dropdown/index.html');
        try {
          const content = fs.readFileSync(filePath, 'utf8');
          return {
            contents: [
              {
                uri: resourceURI,
                mimeType: mimeType,
                text: content,
              },
            ],
          };
        } catch (error: any) {
          throw new Error(`Could not read dropdown UI file: ${error.message}`);
        }
      },
    },
  ];
}
