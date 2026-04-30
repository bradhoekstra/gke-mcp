declare module '@google/adk/dist/esm/agents/llm_agent.js' {
  export class LlmAgent {
    constructor(config: any);
  }
}

declare module '@google/adk/dist/esm/tools/function_tool.js' {
  export class FunctionTool {
    constructor(options: any);
  }
}

declare module '@google/adk/dist/esm/runner/runner.js' {
  export class Runner {
    constructor(config: any);
    runAsync(params: any): AsyncGenerator<any, void, undefined>;
  }
}

declare module '@google/adk/dist/esm/sessions/in_memory_session_service.js' {
  export class InMemorySessionService {
    constructor();
    getSession(params: any): Promise<any>;
    createSession(params: any): Promise<any>;
  }
}
