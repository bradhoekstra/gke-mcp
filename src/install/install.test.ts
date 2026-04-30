import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { installCursor, installClaudeDesktop, installClaudeCode, installGeminiCLI } from './install.js';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { execSync } from 'child_process';
import * as ui from './ui.js';

vi.mock('child_process', () => ({
  execSync: vi.fn(),
}));

vi.mock('os', async (importOriginal) => {
  const actual = await importOriginal<typeof import('os')>();
  return {
    ...actual,
    homedir: vi.fn(),
  };
});

describe('Install', () => {
  let tmpDir: string;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'install-test-'));
    vi.resetAllMocks();
    vi.mocked(os.homedir).mockReturnValue(tmpDir);
    
    vi.spyOn(ui, 'askConfirmation').mockResolvedValue(true);
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
    vi.restoreAllMocks();
  });

  it('should install for Cursor', () => {
    const opts = {
      version: '1.0.0',
      installDir: tmpDir,
      exePath: '/usr/local/bin/gke-mcp',
      developerMode: false,
    };

    installCursor(opts);

    const mcpPath = path.join(tmpDir, '.cursor', 'mcp.json');
    expect(fs.existsSync(mcpPath)).toBe(true);
    
    const config = JSON.parse(fs.readFileSync(mcpPath, 'utf8'));
    expect(config.mcpServers['gke-mcp'].command).toBe('/usr/local/bin/gke-mcp');
    
    const rulePath = path.join(tmpDir, '.cursor', 'rules', 'gke-mcp.mdc');
    expect(fs.existsSync(rulePath)).toBe(true);
  });

  it('should install for Claude Desktop', () => {
    const opts = {
      version: '1.0.0',
      installDir: tmpDir,
      exePath: '/usr/local/bin/gke-mcp',
      developerMode: false,
    };

    installClaudeDesktop(opts);
    
    let expectedPath = '';
    if (process.platform === 'darwin') {
      expectedPath = path.join(tmpDir, 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json');
    } else if (process.platform === 'win32') {
      return;
    } else {
      expectedPath = path.join(tmpDir, '.config', 'Claude', 'claude_desktop_config.json');
    }
    
    expect(fs.existsSync(expectedPath)).toBe(true);
  });

  it('should install for Claude Code', async () => {
    const opts = {
      version: '1.0.0',
      installDir: tmpDir,
      exePath: '/usr/local/bin/gke-mcp',
      developerMode: false,
    };

    await installClaudeCode(opts);

    const usageGuidePath = path.join(tmpDir, 'GKE_MCP_USAGE_GUIDE.md');
    expect(fs.existsSync(usageGuidePath)).toBe(true);
    
    expect(execSync).toHaveBeenCalledWith(expect.stringContaining('claude mcp add'), expect.anything());
  });

  it('should install for Gemini CLI', () => {
    const opts = {
      version: '1.0.0',
      installDir: tmpDir,
      exePath: '/usr/local/bin/gke-mcp',
      developerMode: false,
    };

    installGeminiCLI(opts);

    const extensionDir = path.join(tmpDir, '.gemini', 'extensions', 'gke-mcp');
    const manifestPath = path.join(extensionDir, 'gemini-extension.json');
    expect(fs.existsSync(manifestPath)).toBe(true);
  });
});
