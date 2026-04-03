package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type DoctorResult struct {
	Name   string
	Status string
	Detail string
}

func RunDoctor() []DoctorResult {
	var results []DoctorResult

	results = append(results, DoctorResult{
		Name:   "Go Runtime",
		Status: "ok",
		Detail: fmt.Sprintf("go%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	})

	cfgPaths := ConfigPaths()
	configFound := false
	for _, p := range cfgPaths {
		if _, err := os.Stat(p); err == nil {
			configFound = true
			results = append(results, DoctorResult{
				Name:   "Config",
				Status: "ok",
				Detail: "found at " + p,
			})
			break
		}
	}
	if !configFound {
		results = append(results, DoctorResult{
			Name:   "Config",
			Status: "warn",
			Detail: "no config file found, using defaults",
		})
	}

	if ocPath, err := exec.LookPath("opencode"); err == nil {
		results = append(results, DoctorResult{
			Name:   "OpenCode",
			Status: "ok",
			Detail: "found at " + ocPath,
		})
	} else {
		results = append(results, DoctorResult{
			Name:   "OpenCode",
			Status: "warn",
			Detail: "not found in PATH (optional)",
		})
	}

	home, _ := os.UserHomeDir()
	dbPath := ""
	if home != "" {
		dbPath = fmt.Sprintf("%s/.aigo/aigo.db", home)
	}
	if _, err := os.Stat(dbPath); err == nil {
		results = append(results, DoctorResult{
			Name:   "Database",
			Status: "ok",
			Detail: "found at " + dbPath,
		})
	} else {
		results = append(results, DoctorResult{
			Name:   "Database",
			Status: "info",
			Detail: "will be created on first run",
		})
	}

	skillsPath := ""
	if home != "" {
		skillsPath = fmt.Sprintf("%s/.aigo/skills", home)
	}
	if info, err := os.Stat(skillsPath); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(skillsPath)
		results = append(results, DoctorResult{
			Name:   "Skills",
			Status: "ok",
			Detail: fmt.Sprintf("%d skills in %s", len(entries), skillsPath),
		})
	} else {
		results = append(results, DoctorResult{
			Name:   "Skills",
			Status: "info",
			Detail: "no skills directory",
		})
	}

	envVars := []string{"AIGO_MODEL_DEFAULT", "AIGO_OPENCODE_BINARY", "AIGO_GATEWAY_ENABLED"}
	var setVars []string
	for _, v := range envVars {
		if os.Getenv(v) != "" {
			setVars = append(setVars, v)
		}
	}
	if len(setVars) > 0 {
		results = append(results, DoctorResult{
			Name:   "Environment",
			Status: "ok",
			Detail: strings.Join(setVars, ", "),
		})
	} else {
		results = append(results, DoctorResult{
			Name:   "Environment",
			Status: "info",
			Detail: "no AIGO_* env vars set",
		})
	}

	return results
}
