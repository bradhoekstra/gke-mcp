import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';
import { askConfirmation } from './ui.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const cursorRuleHeader = `---
name: GKE MCP Instructions
description: Provides guidance for using the gke-mcp tool with Cursor.
alwaysApply: true
---

# GKE MCP Tool Instructions

This rule provides context for using the gke-mcp tool within Cursor.

`;

export interface InstallOptions {
  version: string;
  installDir: string;
  exePath: string;
  developerMode: boolean;
}

export function newInstallOptions(version: string, projectOnly: boolean, developerMode: boolean): InstallOptions {
  const installDir = projectOnly ? process.cwd() : os.homedir();
  const exePath = process.argv[1];
  
  return {
    version,
    installDir,
    exePath,
    developerMode,
  };
}



export function installCursor(opts: InstallOptions) {
  const mcpDir = path.join(opts.installDir, '.cursor');
  fs.mkdirSync(mcpDir, { recursive: true });
  
  const mcpPath = path.join(mcpDir, 'mcp.json');
  let config: any = {};
  
  if (fs.existsSync(mcpPath)) {
    config = JSON.parse(fs.readFileSync(mcpPath, 'utf8'));
  }
  
  config.mcpServers = config.mcpServers || {};
  config.mcpServers['gke-mcp'] = {
    command: opts.exePath,
    type: 'stdio',
  };
  
  fs.writeFileSync(mcpPath, JSON.stringify(config, null, 2), { mode: 0o600 });
  
  const rulesDir = path.join(mcpDir, 'rules');
  fs.mkdirSync(rulesDir, { recursive: true });
  
  const geminiMdPath = path.join(__dirname, 'GEMINI.md');
  const geminiMdContent = fs.readFileSync(geminiMdPath, 'utf8');
  
  const ruleContent = cursorRuleHeader + geminiMdContent;
  fs.writeFileSync(path.join(rulesDir, 'gke-mcp.mdc'), ruleContent, { mode: 0o600 });
  
  console.log(`Successfully installed for Cursor in ${mcpPath}`);
}

export function installClaudeDesktop(opts: InstallOptions) {
  let configDir = '';
  const homeDir = os.homedir();
  
  if (process.platform === 'darwin') {
    configDir = path.join(homeDir, 'Library', 'Application Support', 'Claude');
  } else if (process.platform === 'win32') {
    configDir = path.join(process.env.APPDATA || '', 'Claude');
  } else {
    configDir = path.join(homeDir, '.config', 'Claude');
  }
  
  const configPath = path.join(configDir, 'claude_desktop_config.json');
  fs.mkdirSync(configDir, { recursive: true });
  
  let config: any = {};
  if (fs.existsSync(configPath)) {
    config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
  }
  
  config.mcpServers = config.mcpServers || {};
  config.mcpServers['gke-mcp'] = {
    command: opts.exePath,
  };
  
  fs.writeFileSync(configPath, JSON.stringify(config, null, 2), { mode: 0o600 });
  console.log(`Successfully installed for Claude Desktop in ${configPath}`);
}

export async function installClaudeCode(opts: InstallOptions) {
  const installDir = opts.installDir;
  const claudeMDPath = path.join(installDir, 'CLAUDE.md');
  
  const exists = fs.existsSync(claudeMDPath);
  
  if (exists) {
    console.log("Warning: CLAUDE.md already exists. The GKE MCP usage instructions will be appended.");
  } else {
    console.log("Note: CLAUDE.md does not exist. A new one will be created.");
  }
  
  const confirmed = await askConfirmation("Would you like to proceed? (yes/no): ");
  if (!confirmed) {
    console.log("Installation canceled.");
    return;
  }
  
  const geminiMdPath = path.join(__dirname, 'GEMINI.md');
  const geminiMdContent = fs.readFileSync(geminiMdPath, 'utf8');
  
  const usageGuideMDPath = path.join(installDir, 'GKE_MCP_USAGE_GUIDE.md');
  fs.writeFileSync(usageGuideMDPath, geminiMdContent, { mode: 0o600 });
  console.log(`Created ${usageGuideMDPath}`);
  
  const claudeLine = `\n# GKE-MCP Server Instructions\n - @${usageGuideMDPath}`;
  fs.appendFileSync(claudeMDPath, claudeLine, { mode: 0o600 });
  console.log(`Added reference to ${usageGuideMDPath} in CLAUDE.md`);
  
  try {
    execSync(`claude mcp add gke-mcp ${opts.exePath}`, { stdio: 'inherit' });
    console.log("Successfully ran 'claude mcp add'");
  } catch (error: any) {
    console.error(`Failed to run 'claude mcp add': ${error.message}`);
  }
}

export function installGeminiCLI(opts: InstallOptions) {
  let contextFilename = "GEMINI.md";
  
  if (opts.developerMode) {
    contextFilename = path.join(__dirname, 'GEMINI.md');
  }
  
  const extensionDir = path.join(opts.installDir, '.gemini', 'extensions', 'gke-mcp');
  fs.mkdirSync(extensionDir, { recursive: true });
  
  const manifest = {
    name: "gke-mcp",
    version: opts.version,
    description: "Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.",
    contextFileName: contextFilename,
    mcpServers: {
      gke: {
        command: opts.exePath,
      },
    },
  };
  
  const manifestPath = path.join(extensionDir, 'gemini-extension.json');
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2), { mode: 0o600 });
  
  if (!opts.developerMode) {
    const geminiMdPath = path.join(__dirname, 'GEMINI.md');
    const geminiMdContent = fs.readFileSync(geminiMdPath, 'utf8');
    fs.writeFileSync(path.join(extensionDir, 'GEMINI.md'), geminiMdContent, { mode: 0o600 });
  }
  
  console.log(`Successfully installed for Gemini CLI in ${extensionDir}`);
}
