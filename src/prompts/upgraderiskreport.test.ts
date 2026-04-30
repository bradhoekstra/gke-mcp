import { describe, it, expect } from 'vitest';
import { getUpgradeRiskReportPrompts } from './upgraderiskreport.js';
import { Config } from '../config/config.js';

describe('Upgrade Risk Report Prompt', () => {
  const config = new Config('1.0.0');
  const prompts = getUpgradeRiskReportPrompts(config);
  const prompt = prompts.find(p => p.name === 'gke:upgrade-risk-report');

  if (!prompt) {
    throw new Error('gke:upgrade-risk-report prompt not found');
  }

  it('should return prompt for valid request', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
      target_version: "1.28.0",
    });
    expect(result.messages[0].content.text).toContain("my-cluster");
  });

  it('should throw error for empty cluster_name', async () => {
    await expect(prompt.handler({
      cluster_name: "",
      cluster_location: "us-central1",
      target_version: "1.28.0",
    })).rejects.toThrow("argument 'cluster_name' cannot be empty");
  });

  it('should throw error for empty cluster_location', async () => {
    await expect(prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "",
      target_version: "1.28.0",
    })).rejects.toThrow("argument 'cluster_location' cannot be empty");
  });

  it('should throw error for whitespace cluster_name', async () => {
    await expect(prompt.handler({
      cluster_name: "   ",
      cluster_location: "us-central1",
      target_version: "1.28.0",
    })).rejects.toThrow("argument 'cluster_name' cannot be empty");
  });

  it('should have correct description', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
      target_version: "1.28.0",
    });
    expect(result.description).toBe("GKE Cluster Upgrade Risk Report Prompt");
  });

  it('should have correct message role', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
      target_version: "1.28.0",
    });
    expect(result.messages[0].role).toBe("user");
  });

  it('should work without target_version', async () => {
    const result = await prompt.handler({
      cluster_name: "my-cluster",
      cluster_location: "us-central1",
    });
    expect(result.messages[0].content.text).toContain("my-cluster");
  });
});
