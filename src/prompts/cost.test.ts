import { describe, it, expect } from 'vitest';
import { getCostPrompts } from './cost.js';
import { Config } from '../config/config.js';

describe('Cost Prompt', () => {
  const config = new Config('1.0.0');
  const prompts = getCostPrompts(config);
  const costPrompt = prompts.find(p => p.name === 'gke:cost');

  if (!costPrompt) {
    throw new Error('gke:cost prompt not found');
  }

  it('should return prompt for valid question', async () => {
    const result = await costPrompt.handler({ user_question: "How do I optimize my GKE costs?" });
    expect(result.messages[0].content.text).toContain("How do I optimize my GKE costs?");
  });

  it('should throw error for empty question', async () => {
    await expect(costPrompt.handler({ user_question: "" })).rejects.toThrow("argument 'user_question' cannot be empty");
  });

  it('should throw error for whitespace question', async () => {
    await expect(costPrompt.handler({ user_question: "   " })).rejects.toThrow("argument 'user_question' cannot be empty");
  });

  it('should have correct description', async () => {
    const result = await costPrompt.handler({ user_question: "test" });
    expect(result.description).toBe("GKE Cost Analysis Prompt");
  });

  it('should have correct message role', async () => {
    const result = await costPrompt.handler({ user_question: "test" });
    expect(result.messages[0].role).toBe("user");
  });

  it('should contain expected sections', async () => {
    const result = await costPrompt.handler({ user_question: "test" });
    const text = result.messages[0].content.text;
    expect(text).toContain("BigQuery Integration");
    expect(text).toContain("Cost Allocation");
    expect(text).toContain("Actionable Steps");
  });
});
