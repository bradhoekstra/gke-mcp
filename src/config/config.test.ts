import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Config } from './config.js';
import { execSync } from 'child_process';

vi.mock('child_process', () => ({
  execSync: vi.fn(),
}));

describe('Config', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('should set user agent correctly', () => {
    const version = '1.0.0';
    vi.mocked(execSync).mockReturnValue('test-project\n');
    
    const cfg = new Config(version);
    expect(cfg.userAgent).toBe(`gke-mcp/${version}`);
  });

  it('should get default project ID from gcloud', () => {
    vi.mocked(execSync).mockImplementation((cmd: any) => {
      if (cmd === 'gcloud config get core/project') return 'my-project\n';
      return '';
    });

    const cfg = new Config('1.0.0');
    expect(cfg.defaultProjectID).toBe('my-project');
  });

  it('should get default location from gcloud (region)', () => {
    vi.mocked(execSync).mockImplementation((cmd: any) => {
      if (cmd === 'gcloud config get core/project') return 'my-project\n';
      if (cmd === 'gcloud config get compute/region') return 'us-central1\n';
      throw new Error('not found');
    });

    const cfg = new Config('1.0.0');
    expect(cfg.defaultLocation).toBe('us-central1');
  });

  it('should get default location from gcloud (zone fallback)', () => {
    vi.mocked(execSync).mockImplementation((cmd: any) => {
      if (cmd === 'gcloud config get core/project') return 'my-project\n';
      if (cmd === 'gcloud config get compute/region') throw new Error('not found');
      if (cmd === 'gcloud config get compute/zone') return 'us-central1-a\n';
      throw new Error('not found');
    });

    const cfg = new Config('1.0.0');
    expect(cfg.defaultLocation).toBe('us-central1-a');
  });

  it('should handle empty config', () => {
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error('gcloud not found');
    });

    const cfg = new Config('1.0.0');
    expect(cfg.defaultProjectID).toBe('');
    expect(cfg.defaultLocation).toBe('');
  });
});
