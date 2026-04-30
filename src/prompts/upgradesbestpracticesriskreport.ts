import { Config } from '../config/config.js';

export interface PromptDefinition {
  name: string;
  description: string;
  arguments: { name: string; description: string; required: boolean }[];
  handler: (args: any) => Promise<any>;
}

const promptTemplate = `
# GKE Upgrades Best Practices Risk Report Generation

**1. Input Parameters:**
  - Cluster Name: {{clusterName}}
  - Cluster Location: {{clusterLocation}}

**2. Your Role:**
You are a GKE expert. Your task is to verify the cluster whether it follows GKE upgrades best practices and give a comprehensive risk report based on the verification.

**3. Primary Goal:**
Produce a report outlining actual risks, and actionable recommendations on how to mitigate the risks to ensure a safe and smooth GKE upgrades. The report should be based on the GKE upgrades best practices and the cluster actual state.

**4. Information Gathering & Tools:**
Assume you have the ability to run the following commands to gather necessary information:
  - **Cluster Details:** Use \`gcloud\` to get cluster details.
  - **In-Cluster Resources:** Use \`kubectl\` (after \`gcloud container clusters get-credentials\`) for inspecting workloads.

**5. GKE Upgrades Best Practices:**

**5.1. Maintenance Windows**

**Context:** If a cluster doesn't have a maintenance window set, GKE can perform automatic upgrades at any time. Upgrades are rolled out across different regions over several days, so the exact timing of an automatic upgrade without a maintenance window can be unpredictable. A significant number of clusters do not have a maintenance window set, which can lead to unexpected disruptions. There is no default maintenance window configured when a GKE cluster is created. User must explicitly create a maintenance window to control when automatic upgrades can occur.

**Analysis:** You must check whether the cluster has maintenance window set and it is not allowing upgrades at any time.

**5.2. Pod Disruption Budgets (PDBs)**

**Context:** PDBs are a Kubernetes feature that you can use to protect your applications from voluntary disruptions, such as node upgrades. GKE respects PDBs for up to 60 minutes during a node drain. If the pods are not terminated within this time, they will be forcefully removed. For some long-running workloads, this 60-minute graceful termination period may not be sufficient. There is no default PDB for user's workloads. User must create a PDB for each of their applications to define how many concurrent disruptions it can tolerate.

**Analysis:** You must conduct a thorough review of all user-managed applications running in the cluster and check whether there is a proper PDB configuration set for each of them.

**5.3. Node Pool Upgrades (Surge Upgrades)**

**Context:** Surge upgrades are the default strategy for GKE node pools and are always used for Autopilot clusters. This strategy helps maintain application's capacity by creating a new, upgraded node before draining and removing an old one. For larger clusters, user can speed up the upgrade process by increasing the number of nodes that are upgraded concurrently. All new GKE node pools are automatically configured to use surge upgrades with the settings maxSurge=1 and maxUnavailable=0. This configuration means that during an upgrade, GKE will add one extra node to a node pool and will not take any of existing nodes offline until the new one is ready, thus ensuring there is no reduction in capacity.

**Analysis:** You must ensure that all node pools of the cluster have properly configured upgrade strategy, for example configuration with surge strategy with MaxSurge=0 and MaxUnavailable=1 is not recommended because it allows reduction in capacity.

**7. Risk Identification:**
Check whether the cluster follows each best practice. If a best practice is not implemented then it's a risk that needs mitigation.

**8. Report Format:**
Present the risks as a single list. Each risk item MUST follow this markdown structure:

\`\`\`markdown
# Short Risk Title

## Description

(Detailed description of the risk)

## Mitigation Recommendations

(Clear, actionable steps, commands to to mitigate the risk. Provide examples and link to docs.)
\`\`\`

**9. Principles:**
  - Be specific for each risk; avoid grouping unrelated issues.
  - Ensure Mitigation steps are practical and provide sufficient detail for a GKE administrator to act upon.
  - Do not read or write any local files generating the report.
`;

export function getUpgradesBestPracticesRiskReportPrompts(config: Config): PromptDefinition[] {
  return [
    {
      name: 'gke:upgrades-best-practices-risk-report',
      description: "Generate GKE cluster upgrades best practices risk report.",
      arguments: [
        {
          name: 'cluster_name',
          description: "A name of a GKE cluster user want to upgrade.",
          required: true,
        },
        {
          name: 'cluster_location',
          description: "A location of a GKE cluster user want to upgrade.",
          required: true,
        },
      ],
      handler: async (args: any) => {
        const clusterName = args.cluster_name;
        const clusterLocation = args.cluster_location;

        if (!clusterName || clusterName.trim() === '') {
          throw new Error("argument 'cluster_name' cannot be empty");
        }
        if (!clusterLocation || clusterLocation.trim() === '') {
          throw new Error("argument 'cluster_location' cannot be empty");
        }

        const filledTemplate = promptTemplate
          .replace('{{clusterName}}', clusterName)
          .replace('{{clusterLocation}}', clusterLocation);

        return {
          description: "GKE Cluster Upgrade Best Practices Risk Report Prompt",
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
