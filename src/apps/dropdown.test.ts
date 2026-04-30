import { describe, it, expect } from 'vitest';
import { getDropdownTools } from './dropdown.js';
import { Config } from '../config/config.js';

describe('Dropdown App', () => {
  const config = new Config('1.0.0');
  const tools = getDropdownTools(config);
  const dropdownTool = tools.find(t => t.name === 'dropdown');

  if (!dropdownTool) {
    throw new Error('dropdown tool not found');
  }

  it('should return pending response for valid options', async () => {
    const result = await dropdownTool.handler({
      title: "Select a cluster",
      options: ["cluster1", "cluster2"],
    });

    expect(result.structuredContent).toEqual({
      status: "PENDING_USER_INPUT",
      options: ["cluster1", "cluster2"],
      message: "Present these options to the user. Wait until selection is made",
    });
  });

  it('should throw error for empty options', async () => {
    await expect(dropdownTool.handler({ options: [] }))
      .rejects.toThrow("options cannot be empty");
  });

  it('should throw error for missing options', async () => {
    await expect(dropdownTool.handler({}))
      .rejects.toThrow("options cannot be empty");
  });
});
