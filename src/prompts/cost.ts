import { Config } from '../config/config.js';

export interface PromptDefinition {
  name: string;
  description: string;
  arguments: { name: string; description: string; required: boolean }[];
  handler: (args: any) => Promise<any>;
}

const gkeCostPromptTemplate = `
You are a GKE cost and optimization expert. Answer the user's question about GKE costs, optimization, or billing using the comprehensive cost context available in the GKE MCP server.
User Question: {{user_question}}
Based on the GKE cost context available, provide a detailed and helpful response that includes:
1. **Direct Answer**: Address the specific cost question or optimization request
2. **BigQuery Integration**: Explain how to use BigQuery for cost analysis if relevant
3. **Cost Allocation**: Mention GKE Cost Allocation requirements when applicable
4. **Actionable Steps**: Provide concrete next steps or commands when possible
5. **Resource References**: Point to relevant GCP documentation or console links
Key points to remember:
- GKE costs come from GCP Billing Detailed BigQuery Export
- BigQuery CLI (bq) is preferred over BigQuery Studio when available
- GKE Cost Allocation must be enabled for namespace and workload-level cost data
- Required parameters include BigQuery table path, time frame, project ID, cluster details
- Use the cost analysis queries from the GKE MCP documentation as templates
Always be helpful, specific, and actionable in your response.
`;

export function getCostPrompts(config: Config): PromptDefinition[] {
  return [
    {
      name: 'gke:cost',
      description: "Answer natural language questions about GKE-related costs by leveraging the bundled cost context instructions within the gke-mcp server.",
      arguments: [
        {
          name: 'user_question',
          description: "The user's natural language question about GKE costs",
          required: true,
        },
      ],
      handler: async (args: any) => {
        const userQuestion = args.user_question;
        if (!userQuestion || userQuestion.trim() === '') {
          throw new Error("argument 'user_question' cannot be empty");
        }

        const filledTemplate = gkeCostPromptTemplate.replace('{{user_question}}', userQuestion);

        return {
          description: "GKE Cost Analysis Prompt",
          messages: [
            {
              role: 'user',
              content: { type: 'text', text: filledTemplate },
            },
          ],
        };
      },
    },
  ];
}
