#!/usr/bin/env node
import { Command } from 'commander';
import { Config } from './config/config.js';
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { installTools } from './tools/index.js';
import { installPrompts } from './prompts/index.js';
import { newInstallOptions, installGeminiCLI, installCursor, installClaudeDesktop, installClaudeCode } from './install/install.js';

const program = new Command();

program
  .name('gke-mcp')
  .description('An MCP Server for Google Kubernetes Engine')
  .version('1.0.0');

program
  .option('--server-mode <mode>', 'transport to use for the server: stdio or http', 'stdio')
  .option('--server-host <host>', 'server host to use when server-mode is http', '127.0.0.1')
  .option('--server-port <port>', 'server port to use when server-mode is http', '8080')
  .option('--allowed-origins <origins>', 'comma-separated list of allowed Origin headers', 'http://localhost')
  .action(async (options) => {
    console.error(`Starting GKE MCP Server in mode '${options.serverMode}'`);
    
    const config = new Config('1.0.0'); // TODO: pass real version
    
    const server = new Server({
      name: 'GKE MCP Server',
      version: '1.0.0',
    }, {
      capabilities: {
        tools: {},
        resources: {},
        prompts: {},
      }
    });

    installTools(server, config);
    installPrompts(server, config);
    
    if (options.serverMode === 'stdio') {
      const transport = new StdioServerTransport();
      await server.connect(transport);
      console.error('Server connected to stdio');
    } else if (options.serverMode === 'http') {
      console.error('HTTP mode not fully implemented yet in TS port');
      // TODO: implement HTTP server with CORS
    }
  });

const installCmd = program.command('install').description('Install the GKE MCP Server into your AI tool settings.');

installCmd.command('gemini-cli')
  .description('Install the GKE MCP Server into your Gemini CLI settings.')
  .option('-d, --developer', 'Install the MCP Server in developer mode for Gemini CLI')
  .option('-p, --project-only', 'Install the MCP Server only for the current project')
  .action((options) => {
    const opts = newInstallOptions('1.0.0', options.projectOnly, options.developer);
    installGeminiCLI(opts);
  });

installCmd.command('cursor')
  .description('Install the GKE MCP Server into your Cursor settings.')
  .option('-p, --project-only', 'Install the MCP Server only for the current project')
  .action((options) => {
    const opts = newInstallOptions('1.0.0', options.projectOnly, false);
    installCursor(opts);
  });

installCmd.command('claude-desktop')
  .description('Install the GKE MCP Server into your Claude Desktop settings.')
  .action((options) => {
    const opts = newInstallOptions('1.0.0', false, false);
    installClaudeDesktop(opts);
  });

installCmd.command('claude-code')
  .description('Install the GKE MCP Server into your Claude Code CLI settings.')
  .option('-p, --project-only', 'Install the MCP Server only for the current project')
  .action(async (options) => {
    const opts = newInstallOptions('1.0.0', options.projectOnly, false);
    await installClaudeCode(opts);
  });

program.parse(process.argv);
