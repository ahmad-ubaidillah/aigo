package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/internal/python"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type PythonHandler struct {
	kernel *python.Kernel
}

func NewPythonHandler() *PythonHandler {
	return &PythonHandler{
		kernel: python.NewKernel(&python.KernelOptions{
			Timeout: 60 * time.Second,
			WorkDir: "/tmp/aigo-python",
		}),
	}
}

func (h *PythonHandler) CanHandle(intent string) bool {
	return intent == types.IntentPython
}

type PythonAction struct {
	Code     string   `json:"code"`
	Packages []string `json:"packages,omitempty"`
	Files    []string `json:"files,omitempty"`
}

func (h *PythonHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
	var action PythonAction
	if err := json.Unmarshal([]byte(task.Description), &action); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parse action: %v", err),
		}, nil
	}

	if len(action.Packages) > 0 {
		if err := h.kernel.InstallPackages(action.Packages); err != nil {
			return &types.ToolResult{
				Success: false,
				Error:   fmt.Sprintf("install packages: %v", err),
			}, nil
		}
	}

	if err := h.kernel.Start(); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("start kernel: %v", err),
		}, nil
	}

	result, err := h.kernel.Execute(action.Code)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("execute: %v", err),
		}, nil
	}

	return &types.ToolResult{
		Success: result.Success,
		Output:  result.Output,
		Error:   result.Error,
	}, nil
}

func ParsePythonCode(input string) (*PythonAction, error) {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "{") {
		var action PythonAction
		if err := json.Unmarshal([]byte(input), &action); err != nil {
			return nil, err
		}
		return &action, nil
	}

	return &PythonAction{
		Code: input,
	}, nil
}

func ExtractPythonPackages(code string) []string {
	var packages []string
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "import ") {
			pkg := strings.TrimPrefix(line, "import ")
			pkg = strings.Split(pkg, " ")[0]
			if pkg != "" {
				packages = append(packages, pkg)
			}
		}
		if strings.HasPrefix(line, "from ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				packages = append(packages, parts[1])
			}
		}
	}
	return packages
}

type JupyterHandler struct {
	kernel *python.Kernel
	cells  []string
}

func NewJupyterHandler() *JupyterHandler {
	return &JupyterHandler{
		kernel: python.NewKernel(&python.KernelOptions{
			Timeout: 120 * time.Second,
			WorkDir: "/tmp/aigo-jupyter",
		}),
		cells: make([]string, 0),
	}
}

func (h *JupyterHandler) AddCell(code string) {
	h.cells = append(h.cells, code)
}

func (h *JupyterHandler) ClearCells() {
	h.cells = make([]string, 0)
}

func (h *JupyterHandler) ExecuteAll() (*types.ToolResult, error) {
	if err := h.kernel.Start(); err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("start kernel: %v", err),
		}, nil
	}

	var output strings.Builder
	for _, cell := range h.cells {
		result, err := h.kernel.Execute(cell)
		if err != nil {
			return &types.ToolResult{
				Success: false,
				Error:   fmt.Sprintf("execute cell: %v", err),
			}, nil
		}
		if !result.Success {
			return &types.ToolResult{
				Success: false,
				Error:   result.Error,
			}, nil
		}
		output.WriteString(result.Output)
		output.WriteString("\n")
	}

	return &types.ToolResult{
		Success: true,
		Output:  output.String(),
	}, nil
}
