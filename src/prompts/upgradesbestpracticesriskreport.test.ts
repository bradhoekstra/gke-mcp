import { describe, it, expect } from 'vitest';
import { getUpgradesBestPracticesRiskReportPrompts } from './upgradesbestpracticesriskreport.js';
import { Config } from '../config/config.js';

describe('Upgrades Best Practices Risk Report Prompt', () => {
  const config = new Config('1.0.0');
  const prompts = getUpgradesBestPracticesRiskReportPrompts(config);
  const prompt = prompts.find(p => p.name === 'gke:upgrades-best-practices-risk-report');

  if (!prompt) {
    throw new Error('gke:upgrades-best-practices-risk-report prompt not found');
  }

  it('should return prompt for valid request', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
    });
    expect(result.messages[0].content.text).toContain("my-cluster");
  });

  it('should throw error for empty cluster_name', async () => {
    await expect(prompt.handler({
      cluster_name: "",
      cluster_location: "us-central1",
    })).rejects.toThrow("argument 'cluster_name' cannot be empty");
  });

  it('should throw error for empty cluster_location', async () => {
    await expect(prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "",
    })).rejects.toThrow("argument 'cluster_location' cannot be empty");
  });

  it('should throw error for whitespace cluster_name', async () => {
    await expect(prompt.handler({
      cluster_name: "   ",
      cluster_location: "us-central1",
    })).rejects.toThrow("argument 'cluster_name' cannot be empty");
  });

  it('should have correct description', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
    });
    expect(result.description).toBe("GKE Cluster Upgrade Best Practices Risk Report Prompt");
  });

  it('should have correct message role', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
    });
    expect(result.messages[0].role).toBe("user");
  });

  it('should contain expected sections', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
    });
    const text = result.messages[0].content.text;
    expect(text).toContain("Maintenance Windows");
    expect(text).toContain("Pod Disruption Budgets");
    expect(text).toContain("Node Pool Upgrades");
  });
});
