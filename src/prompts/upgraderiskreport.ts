import { Config } from '../config/config.js';

export interface PromptDefinition {
  name: string;
  description: string;
  arguments: { name: string; description: string; required: boolean }[];
  handler: (args: any) => Promise<any>;
}

const gkeUpgradeRiskReportPromptTemplate = `
# GKE Upgrade Risk Report Generation

**1. Input Parameters:**
  - Cluster Name: {{clusterName}}
  - Cluster Location: {{clusterLocation}}
  - Target Version: {{targetVersion}}

**2. Your Role:**
You are a GKE expert. Your task is to generate a comprehensive upgrade risk report for the specified GKE cluster, analyzing the potential risks of upgrading from its current version to the 'Target Version'.

**3. Primary Goal:**
Produce a report outlining potential risks, and actionable recommendations to ensure a safe and smooth GKE upgrade. The report should be based on the changes introduced between the cluster's current control plane version and the 'Target Version'.

**4. Handling Missing Target Version:**
If 'Target Version' is not provided:
  a. State that the target version is required.
  b. Use \`gcloud container get-server-config\` to fetch available GKE versions.
  c. Filter this list to show only versions NEWER than the cluster's current control plane version and compatible with the cluster's release channel.
  d. Present these versions to the user to help them choose a 'Target Version'.

**5. Information Gathering & Tools:**
Assume you have the ability to run the following commands to gather necessary information:
  - **Cluster Details:** Use \`gcloud\` to get cluster details like control plane version, release channel, node pool versions, etc.
  - **In-Cluster Resources:** Use \`kubectl\` (after \`gcloud container clusters get-credentials\`) for inspecting workloads, APIs in use, etc.
  - **Kubernetes Changelogs:** Use the \`get_k8s_changelog\` tool to fetch kubernetes changelogs.
  - **GKE Release Notes:** Use the \`get_gke_release_notes\` tool to fetch GKE release notes.

**6. Changelog Analysis:**
  - **Minor Versions:** Include changelogs for ALL minor versions from the current control plane minor version up to AND INCLUDING the target minor version. (e.g., 1.29.x to 1.31.y requires looking at changes in 1.29, 1.30, 1.31).
  - **Patch Versions:** Analyze changes for EVERY patch version BETWEEN the current version (exclusive) and the target version (inclusive). (e.g., 1.29.1 to 1.29.5 means analyzing 1.29.2, 1.29.3, 1.29.4, 1.29.5).
  - **GKE Versions:** Analyze changes for GKE version BETWEEN the current version (exclusive) and the target version (inclusive). (e.g., 1.29.1-gke.123000 to 1.29.5-gke.234000 means analyzing 1.29.1-gke.123500, 1.29.1-gke.124000 etc, and 1.29.5-gke.234000).

**7. Risk Identification - Focus on:**
  - **API Deprecations/Removals:** Especially those affecting in-use cluster resources.
  - **Breaking Changes:** Significant behavioral changes in existing, stable features.
  - **Default Configuration Changes:** Modifications to defaults that could alter workload behavior.
  - **New Feature Interactions:** Potentially disruptive interactions between new features and existing setups.
  - Changes REQUIRING manual action before upgrade to prevent outages.

**8. Report Format:**
Present the risks as a single list, ordered by severity. Each risk item MUST follow this markdown structure:

\`\`\`markdown
# Short Risk Title

## Description

(Detailed description of the change and the potential risk it introduces for THIS specific upgrade)

## Verification Recommendations

(Clear, actionable steps or commands to check if the cluster is affected by this risk. Include example \`kubectl\` or \`gcloud\` commands where appropriate. Reference specific documentation links if possible.)

## Mitigation Recommendations

(Clear, actionable steps, configuration changes, or code adjustments to mitigate the risk BEFORE the upgrade. Provide examples and link to docs.)
\`\`\`

**9. Principles:**
  - Be specific for each risk; avoid grouping unrelated issues.
  - Ensure Verification and Mitigation steps are practical and provide sufficient detail for a GKE administrator to act upon.
  - Base the analysis SOLELY on the changes between the cluster's current version and the target version.
  - Do not read or write any local files generating the report.
  - In the final report, keep only risks which have mitigation actions, ignore those which have no mitigation actions.
`;

export function getUpgradeRiskReportPrompts(config: Config): PromptDefinition[] {
  return [
    {
      name: 'gke:upgrade-risk-report',
      description: "Generate GKE cluster upgrade risk report.",
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
        {
          name: 'target_version',
          description: "A version user want to upgrade their cluster to.",
          required: false,
        },
      ],
      handler: async (args: any) => {
        const clusterName = args.cluster_name;
        const clusterLocation = args.cluster_location;
        const targetVersion = args.target_version || '';

        if (!clusterName || clusterName.trim() === '') {
          throw new Error("argument 'cluster_name' cannot be empty");
        }
        if (!clusterLocation || clusterLocation.trim() === '') {
          throw new Error("argument 'cluster_location' cannot be empty");
        }

        const filledTemplate = gkeUpgradeRiskReportPromptTemplate
          .replace('{{clusterName}}', clusterName)
          .replace('{{clusterLocation}}', clusterLocation)
          .replace('{{targetVersion}}', targetVersion);

        return {
          description: "GKE Cluster Upgrade Risk Report Prompt",
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
