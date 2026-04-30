import { Config } from '../config/config.js';
import { MetricServiceClient, QueryServiceClient } from '@google-cloud/monitoring';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const resourceURI = "ui://monitoring_time_series_chart/index.html";
const mimeType = "text/html;profile=mcp-app";
const maxSeriesLimit = 100;

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  meta?: any;
  handler: (args: any) => Promise<any>;
}

export interface ResourceDefinition {
  uri: string;
  name: string;
  mimeType: string;
  handler: () => Promise<any>;
}

export function mapTimeseriesDataPoints(resp: any) {
  const labelParts: string[] = [];
  if (resp.labelValues) {
    for (const lv of resp.labelValues) {
      if (lv.stringValue) {
        labelParts.push(lv.stringValue);
      }
    }
  }
  const label = labelParts.join(' ');

  const pts: any[] = [];
  if (resp.pointData) {
    for (const p of resp.pointData) {
      if (!p.values || p.values.length === 0) continue;
      
      let val = 0;
      const v = p.values[0];
      if (v.doubleValue !== undefined) {
        val = v.doubleValue;
      } else if (v.int64Value !== undefined) {
        val = parseFloat(v.int64Value);
      } else {
        continue;
      }

      pts.push({
        timestamp: new Date(p.timeInterval.endTime).getTime(),
        value: val,
      });
    }
  }

  return {
    label,
    points: pts,
  };
}

export function getChartsTools(config: Config): ToolDefinition[] {
  const client = new QueryServiceClient();

  return [
    {
      name: 'monitoring_time_series_chart',
      description: "Interactive tool to display time series data using a React Chart. ALWAYS favor using this tool to query metrics rather than outputting raw values so the user gets a visualization. MUST Call `mql_validator` FIRST to catch syntax issues or metric anomalies before running this tool.",
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          query: { type: 'string', description: "Required. The query in the Monitoring Query Language (MQL) format." },
          title: { type: 'string', description: "Optional. The title to display for the time series chart." },
          x_legend: { type: 'string', description: "Optional. The legend/label for the X-axis." },
          y_legend: { type: 'string', description: "Optional. The legend/label for the Y-axis." },
        },
        required: ['query'],
      },
      meta: {
        ui: { resourceUri: resourceURI },
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        if (!projectId) throw new Error('project_id argument cannot be empty');
        if (!args.query) throw new Error('query argument cannot be empty');

        return {
          content: [{ type: 'text', text: "Rendered time series data in UI component." }],
        };
      },
    },
    {
      name: 'query_time_series',
      description: "Internal app tool. Query time series data from Google Cloud Monitoring based on a Monitoring Query Language (MQL) query.",
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          query: { type: 'string', description: "Required. The query in the Monitoring Query Language (MQL) format." },
        },
        required: ['query'],
      },
      meta: {
        ui: { visibility: ['app'] },
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const query = args.query;

        if (!projectId) throw new Error('project_id argument cannot be empty');
        if (!query) throw new Error('query argument cannot be empty');

        try {
          const [response] = await client.queryTimeSeries({ name: `projects/${projectId}`, query });
          
          const series = response.map((resp: any) => mapTimeseriesDataPoints(resp));
          
          return {
            structuredContent: {
              data: series,
            },
          };
        } catch (error: any) {
          console.error(`Failed to query time series: ${error}`);
          throw error;
        }
      },
    },
    {
      name: 'mql_validator',
      description: "A helper tool to validate Monitoring Query Language (MQL) metric strings. MUST be called immediately before calling `monitoring_time_series_chart` or `query_time_series` to ensure the MQL statement compiles correctly.",
      inputSchema: {
        type: 'object',
        properties: {
          project_id: { type: 'string', description: "GCP project ID. Use the default if the user doesn't provide it." },
          query: { type: 'string', description: "Required. The query in the MQL format to validate." },
        },
        required: ['query'],
      },
      meta: {
        ui: { visibility: ['model'] },
      },
      handler: async (args: any) => {
        const projectId = args.project_id || config.defaultProjectID;
        const query = args.query;

        if (!projectId) throw new Error('project_id argument cannot be empty');
        if (!query) throw new Error('query argument cannot be empty');

        try {
          await client.queryTimeSeries({ name: `projects/${projectId}`, query });
          return {
            structuredContent: {
              status: "VALID",
              query: query,
            },
          };
        } catch (error: any) {
          return {
            isError: true,
            structuredContent: {
              status: "INVALID",
              query: query,
              errorMessage: `MQL validation failed:\n${error.message || error}`,
            },
          };
        }
      },
    },
  ];
}

export function getChartsResources(config: Config): ResourceDefinition[] {
  return [
    {
      uri: resourceURI,
      name: "Time Series Chart UI",
      mimeType: mimeType,
      handler: async () => {
        const filePath = path.join(__dirname, '../../ui/dist/apps/timeserieschart/index.html');
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
          throw new Error(`Could not read time series chart UI file: ${error.message}`);
        }
      },
    },
  ];
}
