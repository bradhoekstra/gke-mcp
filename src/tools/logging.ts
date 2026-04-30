import { Config } from '../config/config.js';
import { Logging } from '@google-cloud/logging';
import * as Handlebars from 'handlebars';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const supportedLogTypes: Record<string, boolean> = {
  "k8s_audit_logs": true,
  "k8s_application_logs": true,
  "k8s_event_logs": true,
};

export function getLoggingTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'query_logs',
      description: "Query Google Cloud Platform logs using Logging Query Language (LQL). Before using this tool, it's **strongly** recommended to call the 'get_log_schema' tool to get information about supported log types and their schemas. Logs are returned in ascending order, based on the timestamp (i.e. oldest first).",
      inputSchema: {
        type: 'object',
        properties: {
          query: { type: 'string', description: "LQL query string to filter and retrieve log entries. Don't specify time ranges in this filter. Use 'time_range' instead." },
          project_id: { type: 'string', description: "GCP project ID to query logs from. Required." },
          time_range: {
            type: 'object',
            properties: {
              start_time: { type: 'string', description: "Start time for log query (RFC3339 format)" },
              end_time: { type: 'string', description: "End time for log query (RFC3339 format)" },
            },
          },
          since: { type: 'string', description: "Only return logs newer than a relative duration like 5s, 2m, or 3h." },
          limit: { type: 'integer', description: "Maximum number of log entries to return. Cannot be greater than 100. Defaults to 10." },
          format: { type: 'string', description: "Template string to format each log entry. If empty, the full JSON representation is returned." },
        },
        required: ['project_id'],
      },
      handler: async (args: any) => {
        const projectId = args.project_id;
        const limit = args.limit || 10;
        if (limit > 100) {
          throw new Error('limit parameter cannot be greater than 100');
        }

        let filter = args.query || '';

        if (args.since) {
          const match = args.since.match(/^(\d+)([smh])$/);
          if (!match) {
            throw new Error(`invalid since parameter: ${args.since}`);
          }
          const val = parseInt(match[1], 10);
          const unit = match[2];
          let ms = 0;
          if (unit === 's') ms = val * 1000;
          if (unit === 'm') ms = val * 60 * 1000;
          if (unit === 'h') ms = val * 60 * 60 * 1000;
          
          const startTime = new Date(Date.now() - ms).toISOString();
          filter += (filter ? ' AND ' : '') + `timestamp >= "${startTime}"`;
        }

        if (args.time_range) {
          if (args.time_range.start_time) {
            filter += (filter ? ' AND ' : '') + `timestamp >= "${args.time_range.start_time}"`;
          }
          if (args.time_range.end_time) {
            filter += (filter ? ' AND ' : '') + `timestamp <= "${args.time_range.end_time}"`;
          }
        }

        const logging = new Logging({ projectId });

        try {
          const [entries] = await logging.getEntries({
            filter: filter,
            maxResults: limit + 1,
            orderBy: 'timestamp asc',
          });

          const truncated = entries.length > limit;
          const resultEntries = truncated ? entries.slice(0, limit) : entries;

          let allLogLines = '';
          if (resultEntries.length === 0) {
            allLogLines = 'No log entries found.';
          } else {
            let formatter = (entry: any) => JSON.stringify(entry.toJSON(), null, 2);
            
            if (args.format) {
              const hbTemplate = args.format.replace(/\{\{\s*\.([a-zA-Z0-9_]+)\s*\}\}/g, '{{$1}}');
              const template = Handlebars.compile(hbTemplate);
              formatter = (entry: any) => template(entry.metadata);
            }

            allLogLines = resultEntries.map(e => {
              try {
                return formatter(e);
              } catch (err: any) {
                return `Error formatting entry: ${err.message || err}`;
              }
            }).join('\n');
          }

          let result = `Project ID: ${projectId}\nLQL Query:\n\`\`\`\n${filter}\n\`\`\`\nResult:\n\n${allLogLines}`;
          if (truncated) {
            result += `\n\nWarning: Results truncated. The query returned more than the limit of ${limit} log entries.`;
          }

          return {
            content: [{ type: 'text', text: result }],
          };
        } catch (error: any) {
          console.error(`Failed to query logs: ${error}`);
          throw error;
        }
      },
    },
    {
      name: 'get_log_schema',
      description: 'Get the schema for a specific log type.',
      inputSchema: {
        type: 'object',
        properties: {
          log_type: { type: 'string', description: "The type of log to get schema for. Supported values are: ['k8s_audit_logs', 'k8s_application_logs', 'k8s_event_logs']." },
        },
        required: ['log_type'],
      },
      handler: async (args: any) => {
        const logType = args.log_type;
        if (!supportedLogTypes[logType]) {
          throw new Error(`unsupported log_type: ${logType}`);
        }

        const fileName = `${logType}.md`;
        const filePath = path.join(__dirname, 'logging', 'schemas', fileName);
        
        try {
          const content = fs.readFileSync(filePath, 'utf8');
          return {
            content: [{ type: 'text', text: content }],
          };
        } catch (error: any) {
          throw new Error(`could not find schema for log_type ${logType}: ${error.message}`);
        }
      },
    },
  ];
}
