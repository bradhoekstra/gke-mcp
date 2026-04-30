import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Agent } from './manifestgen.js';
import { Config } from '../config/config.js';

vi.mock('@google/adk/dist/esm/runner/runner.js', () => {
  return {
    Runner: class {
      runAsync = vi.fn().mockImplementation(async function* () {
        yield { content: { parts: [{ text: "apiVersion: apps/v1\nkind: Deployment" }] } };
      });
    },
  };
});

describe('ManifestAgent', () => {
  const config = new Config('1.0.0');
  
  it('should generate manifest successfully', async () => {
    const agent = new Agent(config);
    const result = await agent.run("nginx", "test-session");
    expect(result).toContain("Deployment");
  });

  it('should read instruction.md', () => {
    const agent = new Agent(config);
    expect(agent).toBeTruthy();
  });
});
