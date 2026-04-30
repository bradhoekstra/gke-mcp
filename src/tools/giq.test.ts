import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getGiqTools } from './giq.js';
import { Config } from '../config/config.js';
import { exec as execCallback } from 'child_process';

vi.mock('child_process', async (importOriginal) => {
  const actual = await importOriginal<typeof import('child_process')>();
  return {
    ...actual,
    exec: vi.fn(),
    execSync: vi.fn().mockReturnValue('test-project\n'),
  };
});

describe('GIQ Tool', () => {
  // Config constructor calls execSync, which is now mocked above.
  const config = new Config('1.0.0');
  const tools = getGiqTools(config);
  const fetchModelsTool = tools.find(t => t.name === 'giq_fetch_models');
  const generateManifestTool = tools.find(t => t.name === 'giq_generate_manifest');

  if (!fetchModelsTool || !generateManifestTool) {
    throw new Error('GIQ tools not found');
  }

  beforeEach(() => {
    vi.resetAllMocks();
    // Re-mock execSync for Config if needed, but it's already mocked at module level.
  });

  it('should fetch models via gcloud', async () => {
    vi.mocked(execCallback).mockImplementation((cmd: string, callback: any) => {
      callback(null, { stdout: 'model-A\nmodel-B\nmodel-C\n' });
      return {} as any;
    });

    const result = await fetchModelsTool.handler({});
    expect(result.content[0].text).toBe('model-A\nmodel-B\nmodel-C');
  });

  it('should generate manifest via gcloud', async () => {
    vi.mocked(execCallback).mockImplementation((cmd: string, callback: any) => {
      callback(null, { stdout: 'manifest content' });
      return {} as any;
    });

    const result = await generateManifestTool.handler({
      model: 'llama2',
      model_server: 'vllm',
      accelerator: 't4',
    });
    expect(result.content[0].text).toBe('manifest content');
  });
});
