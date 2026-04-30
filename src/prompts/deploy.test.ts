import { describe, it, expect } from 'vitest';
import { getDeployPrompts } from './deploy.js';
import { Config } from '../config/config.js';

describe('Deploy Prompt', () => {
  const config = new Config('1.0.0');
  const prompts = getDeployPrompts(config);
  const deployPrompt = prompts.find(p => p.name === 'gke:deploy');

  if (!deployPrompt) {
    throw new Error('gke:deploy prompt not found');
  }

  it('should return prompt for valid request', async () => {
    const result = await deployPrompt.handler({ user_request: "Deploy my-app.yaml to staging" });
    expect(result.messages[0].content.text).toContain("You are an expert GKE (Google Kubernetes Engine) deployment assistant.");
  });

  it('should throw error for empty request', async () => {
    await expect(deployPrompt.handler({ user_request: "" })).rejects.toThrow("argument 'user_request' cannot be empty");
  });

  it('should throw error for whitespace request', async () => {
    await expect(deployPrompt.handler({ user_request: "   " })).rejects.toThrow("argument 'user_request' cannot be empty");
  });
});
