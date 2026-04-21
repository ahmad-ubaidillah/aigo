package evolution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents a codebase manifest for the self-evolution system.
type Manifest struct {
	Module   string             `yaml:"module"`
	Packages []PackageManifest  `yaml:"packages"`
}

// PackageManifest describes a single Go package.
type PackageManifest struct {
	Package      string   `yaml:"package"`
	Path         string   `yaml:"path"`
	Files        []string `yaml:"files"`
	Types        []string `yaml:"types,omitempty"`
	Functions    []string `yaml:"functions,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// GenerateManifest scans the project and produces a YAML manifest.
func GenerateManifest(projectDir string) (string, error) {
	projectDir = filepath.Clean(projectDir)

	// Read go.mod for module name
	moduleName, err := readModuleName(projectDir)
	if err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}

	// Collect packages from internal/ and cmd/
	dirs := []string{"internal", "cmd"}
	pkgMap := make(map[string]*PackageManifest) // keyed by import path

	for _, dir := range dirs {
		root := filepath.Join(projectDir, dir)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip inaccessible entries
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			// Skip test files
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			relDir, _ := filepath.Rel(projectDir, filepath.Dir(path))
			relFile, _ := filepath.Rel(projectDir, path)
			pkgImportPath := moduleName + "/" + relDir

			pm, ok := pkgMap[pkgImportPath]
			if !ok {
				pm = &PackageManifest{
					Package: filepath.Base(filepath.Dir(path)),
					Path:    relDir,
				}
				pkgMap[pkgImportPath] = pm
			}
			pm.Files = append(pm.Files, filepath.Base(relFile))

			// Parse the Go file
			if err := parseGoFile(path, moduleName, pm); err != nil {
				return nil // skip files that fail to parse
			}

			return nil
		})
		if err != nil {
			return "", fmt.Errorf("walking %s: %w", dir, err)
		}
	}

	// Build sorted manifest
	pkgs := make([]PackageManifest, 0, len(pkgMap))
	for _, pm := range pkgMap {
		sort.Strings(pm.Files)
		sort.Strings(pm.Types)
		sort.Strings(pm.Functions)
		sort.Strings(pm.Dependencies)
		pkgs = append(pkgs, *pm)
	}
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].Path < pkgs[j].Path
	})

	manifest := Manifest{
		Module:   moduleName,
		Packages: pkgs,
	}

	out, err := yaml.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("marshaling manifest: %w", err)
	}
	return string(out), nil
}

func readModuleName(projectDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

func parseGoFile(path, moduleName string, pm *PackageManifest) error {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Collect exported types
	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
						pm.Types = append(pm.Types, ts.Name.Name)
					}
				}
			}
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				name := d.Name.Name
				if d.Recv != nil {
					// Method — include receiver type
					recv := formatReceiver(d.Recv.List[0].Type)
					name = recv + "." + name
				}
				pm.Functions = append(pm.Functions, name)
			}
		}
	}

	// Collect internal dependencies
	seen := make(map[string]bool)
	for _, imp := range astFile.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasPrefix(path, moduleName+"/") && path != moduleName {
			rel := strings.TrimPrefix(path, moduleName+"/")
			if !seen[rel] {
				pm.Dependencies = append(pm.Dependencies, rel)
				seen[rel] = true
			}
		}
	}

	return nil
}

func formatReceiver(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return "?"
}
