// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gkeagent provides the orchestration agent that wraps skills as sub-agents.
package gkeagent

import (
	"context"
	"fmt"
	"iter"
	"os"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// Agent handles orchestration by wrapping skills as sub-agents.
type Agent struct {
	client         *genai.Client
	cfg            *config.Config
	skillsDir      string
	adkAgent       agent.Agent
	adkRunner      *runner.Runner
	sessionService session.Service
}

// NewAgent creates a new Agent.
func NewAgent(ctx context.Context, cfg *config.Config, skillsDir string) (*Agent, error) {
	projectID := cfg.DefaultProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("default project ID not set in config")
	}
	location := cfg.DefaultLocation()
	if location == "" {
		location = "us-central1"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  projectID,
		Location: location,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	sessSvc := session.InMemoryService()
	a := &Agent{
		client:         client,
		cfg:            cfg,
		skillsDir:      skillsDir,
		sessionService: sessSvc,
	}

	adkAgent, err := agent.New(agent.Config{
		Name:        "gke_agent",
		Description: "Orchestration agent for GKE workflows.",
		Run:         a.adkRun,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK agent: %w", err)
	}
	a.adkAgent = adkAgent

	adkRunner, err := runner.New(runner.Config{
		AppName:        "gke-agent",
		Agent:          adkAgent,
		SessionService: sessSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK runner: %w", err)
	}
	a.adkRunner = adkRunner

	return a, nil
}

// Run handles the main agent request.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	msg := genai.NewContentFromText(prompt, "")

	sessionID := uuid.New().String()

	_, err := a.sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "gke-agent",
		UserID:    "default-user",
		SessionID: sessionID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	events := a.adkRunner.Run(ctx, "default-user", sessionID, msg, agent.RunConfig{})

	var result string
	for event, err := range events {
		if err != nil {
			return "", err
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				result += part.Text
			}
		}
	}

	return result, nil
}

// adkRun implements the ADK agent's Run method.
func (a *Agent) adkRun(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		skills, err := a.listSkills()
		if err != nil {
			yield(nil, fmt.Errorf("failed to list skills: %w", err))
			return
		}

		systemInstruction := "You are the GKE Agent, a master orchestration agent for GKE. You have access to the following skills that you can offer as sub-agents:\n"
		for _, skill := range skills {
			systemInstruction += fmt.Sprintf("- %s\n", skill)
		}
		systemInstruction += "\nTo invoke a skill, you should simulate a tool call to 'invoke_skill' with the skill name and your prompt."

		userPrompt := ""
		if ctx.UserContent() != nil && len(ctx.UserContent().Parts) > 0 {
			userPrompt = ctx.UserContent().Parts[0].Text
		}

		config := &genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(systemInstruction, ""),
		}

		resp, err := a.client.Models.GenerateContent(ctx, "gemini-2.5-pro", genai.Text(userPrompt), config)
		if err != nil {
			yield(nil, fmt.Errorf("failed to generate content: %w", err))
			return
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			yield(nil, fmt.Errorf("empty response from model"))
			return
		}

		var result string
		for _, part := range resp.Candidates[0].Content.Parts {
			result += part.Text
		}

		event := session.NewEvent(ctx.InvocationID())
		event.Content = genai.NewContentFromText(result, "")

		yield(event, nil)
	}
}

// InvokeSkill simulates invoking a specific skill as a sub-agent.

func (a *Agent) listSkills() ([]string, error) {
	entries, err := os.ReadDir(a.skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}
	return skills, nil
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	agent, err := NewAgent(ctx, c, "skills")
	if err != nil {
		return err
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "gke_agent",
		Description: "Orchestration agent for GKE workflows. Wraps all repository skills.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args *struct {
		Prompt string `json:"prompt" jsonschema:"The request for the agent."`
	}) (*mcp.CallToolResult, any, error) {
		result, err := agent.Run(ctx, args.Prompt)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: result,
				},
			},
		}, nil, nil
	})

	return nil
}
