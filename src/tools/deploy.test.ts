import { describe, it, expect } from 'vitest';
import { getDeployTools } from './deploy.js';
import { Config } from '../config/config.js';

describe('Deploy Tool', () => {
  const config = new Config('1.0.0');
  const tools = getDeployTools(config);
  const deployTool = tools.find(t => t.name === 'gke_deploy');

  if (!deployTool) {
    throw new Error('gke_deploy tool not found');
  }

  it('should return prompt for valid request', async () => {
    const result = await deployTool.handler({ user_request: "deploy my-app to staging" });
    expect(result.content[0].text).toContain("You are an expert GKE (Google Kubernetes Engine) deployment assistant.");
  });

  it('should throw error for empty request', async () => {
    await expect(deployTool.handler({ user_request: "" })).rejects.toThrow("argument 'user_request' cannot be empty");
  });

  it('should throw error for whitespace request', async () => {
    await expect(deployTool.handler({ user_request: "   " })).rejects.toThrow("argument 'user_request' cannot be empty");
  });
});
