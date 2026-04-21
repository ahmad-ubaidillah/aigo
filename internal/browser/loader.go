package browser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads a workflow YAML file from disk.
func Load(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow: %w", err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses workflow YAML from a byte slice.
func LoadFromBytes(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}
	if wf.Actions == nil || len(wf.Actions) == 0 {
		return nil, fmt.Errorf("workflow has no actions")
	}
	return &wf, nil
}

// InterpolateEnv replaces ${VAR} placeholders in the workflow's string fields
// with values from the provided env map. The workflow's own Env fields serve
// as defaults and are overridden by the supplied env map.
func InterpolateEnv(wf *Workflow, env map[string]string) {
	merged := make(map[string]string)
	for k, v := range wf.Env {
		merged[k] = v
	}
	for k, v := range env {
		merged[k] = v
	}
	wf.Env = merged
}

// interpolateString replaces ${VAR} in s using the given env map.
func interpolateString(s string, env map[string]string) string {
	if env == nil {
		return s
	}
	re := regexp.MustCompile(`\$\{(\w+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if val, ok := env[key]; ok {
			return val
		}
		return match
	})
}
