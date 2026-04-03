package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/internal/tools"
)

func BuildToolSelectionPrompt(task string, availableTools []tools.Tool, contextSummary string) string {
	var b strings.Builder

	b.WriteString("Choose the best tool for the following task.\n\n")
	b.WriteString(fmt.Sprintf("## Task\n%s\n\n", task))

	b.WriteString("## Available Tools\n")
	for _, t := range availableTools {
		b.WriteString(FormatToolSchema(t))
		b.WriteString("\n")
	}

	if contextSummary != "" {
		b.WriteString("## Context\n")
		b.WriteString(contextSummary)
		b.WriteString("\n")
	}

	b.WriteString("## Instructions\n")
	b.WriteString("Respond with a single JSON object containing:\n")
	b.WriteString(`- "tool_name": the exact name of the tool to use` + "\n")
	b.WriteString(`- "parameters": an object with the parameters for the tool` + "\n")
	b.WriteString("Do not include any text outside the JSON object.\n")

	return b.String()
}

func FormatToolSchema(t tools.Tool) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("### %s\n", t.Name()))
	b.WriteString(fmt.Sprintf("Description: %s\n", t.Description()))

	schema := t.Schema()
	if len(schema) > 0 {
		schemaJSON, err := json.MarshalIndent(schema, "  ", "  ")
		if err == nil {
			b.WriteString(fmt.Sprintf("Parameters:\n  %s\n", string(schemaJSON)))
		}
	}

	return b.String()
}

type toolSelectionResponse struct {
	ToolName   string         `json:"tool_name"`
	Parameters map[string]any `json:"parameters"`
}

func ParseLLMToolResponse(content string) (string, map[string]any, error) {
	content = strings.TrimSpace(content)

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return "", nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := content[start : end+1]

	var resp toolSelectionResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return "", nil, fmt.Errorf("parse LLM response: %w", err)
	}

	if resp.ToolName == "" {
		return "", nil, fmt.Errorf("empty tool_name in LLM response")
	}

	if resp.Parameters == nil {
		resp.Parameters = make(map[string]any)
	}

	return resp.ToolName, resp.Parameters, nil
}
