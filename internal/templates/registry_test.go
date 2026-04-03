package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	if r == nil {
		t.Error("expected registry")
	}
	if len(r.templates) != 5 {
		t.Errorf("expected 5 builtin templates, got %d", len(r.templates))
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := &Template{ID: "test", Name: "Test", Category: "test"}
	err := r.Register(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.templates) != 6 {
		t.Errorf("expected 6 templates, got %d", len(r.templates))
	}
}

func TestRegistry_RegisterEmptyID(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.Register(&Template{})
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	r.Register(&Template{ID: "dup", Name: "Dup"})
	err := r.Register(&Template{ID: "dup", Name: "Dup2"})
	if err == nil {
		t.Error("expected error for duplicate")
	}
}

func TestRegistry_Get(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl, err := r.Get("web_scraper")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Name != "Web Scraper" {
		t.Errorf("expected Web Scraper, got %s", tmpl.Name)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	list := r.List()
	if len(list) != 5 {
		t.Errorf("expected 5 templates, got %d", len(list))
	}
}

func TestRegistry_ListByCategory(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	list := r.ListByCategory("automation")
	if len(list) < 1 {
		t.Errorf("expected at least 1 automation template, got %d", len(list))
	}
}

func TestRegistry_Search(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	results := r.Search("scraper")
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRegistry_SearchEmpty(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	results := r.Search("nonexistent_xyz_12345")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRegistry_Update(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := &Template{ID: "web_scraper", Name: "Updated Scraper"}
	err := r.Update("web_scraper", tmpl)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := r.Get("web_scraper")
	if got.Name != "Updated Scraper" {
		t.Errorf("expected Updated Scraper, got %s", got.Name)
	}
}

func TestRegistry_UpdateNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.Update("nonexistent", &Template{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	r.Register(&Template{ID: "to_remove", Name: "Remove"})
	err := r.Unregister("to_remove")
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Get("to_remove")
	if err == nil {
		t.Error("expected not found")
	}
}

func TestRegistry_UnregisterNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.Unregister("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_IncrementUsage(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	r.IncrementUsage("web_scraper")
	tmpl, _ := r.Get("web_scraper")
	if tmpl.UsageCount != 1 {
		t.Errorf("expected 1, got %d", tmpl.UsageCount)
	}
}

func TestRegistry_IncrementUsageNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.IncrementUsage("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_SaveToFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	r := NewRegistry(dir)
	err := r.SaveToFile("web_scraper")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web_scraper.json")); err != nil {
		t.Error("expected file to exist")
	}
}

func TestRegistry_SaveToFileNoDir(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.SaveToFile("web_scraper")
	if err == nil {
		t.Error("expected error for empty dir")
	}
}

func TestRegistry_SaveToFileNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	r := NewRegistry(dir)
	err := r.SaveToFile("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_LoadFromFile(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.json")
	os.WriteFile(path, []byte(`{"id":"custom","name":"Custom","category":"test"}`), 0644)
	err := r.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Get("custom")
	if err != nil {
		t.Error("expected custom template")
	}
}

func TestRegistry_LoadFromFileInvalid(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	os.WriteFile(path, []byte(`{invalid json}`), 0644)
	err := r.LoadFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRegistry_LoadFromFileNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.LoadFromFile("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_ExportAll(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	dir := t.TempDir()
	err := r.ExportAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 5 {
		t.Errorf("expected 5 files, got %d", len(entries))
	}
}

func TestRegistry_ExportAllInvalidDir(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.ExportAll("/proc/invalid/dir/that/cannot/exist")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_ImportDir(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "imported.json"), []byte(`{"id":"imported","name":"Imported","category":"test"}`), 0644)
	err := r.ImportDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Get("imported")
	if err != nil {
		t.Error("expected imported template")
	}
}

func TestRegistry_ImportDirNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	err := r.ImportDir("/nonexistent/dir")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRegistry_Instantiate(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	wf, err := r.Instantiate("web_scraper", map[string]interface{}{"url": "http://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if wf == nil {
		t.Error("expected workflow")
	}
}

func TestRegistry_InstantiateNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	_, err := r.Instantiate("nonexistent", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestTemplateExecutor(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	e := NewTemplateExecutor(r)
	if e == nil {
		t.Error("expected executor")
	}
}

func TestTemplateExecutor_InstantiateAndRun(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	e := NewTemplateExecutor(r)
	err := e.InstantiateAndRun("web_scraper", map[string]interface{}{"url": "http://example.com"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTemplateExecutor_InstantiateAndRunNotFound(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	e := NewTemplateExecutor(r)
	err := e.InstantiateAndRun("nonexistent", nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLowercaseFold(t *testing.T) {
	t.Parallel()
	if lowercaseFold("HELLO") != "hello" {
		t.Error("expected hello")
	}
	if lowercaseFold("hello") != "hello" {
		t.Error("expected hello")
	}
	if lowercaseFold("") != "" {
		t.Error("expected empty")
	}
}

func TestContainsFold(t *testing.T) {
	t.Parallel()
	if !containsFold("Hello World", "world") {
		t.Error("expected true")
	}
	if containsFold("Hello", "xyz") {
		t.Error("expected false")
	}
}

func TestContainsSubstring(t *testing.T) {
	t.Parallel()
	if !containsSubstring("hello world", "world") {
		t.Error("expected true")
	}
	if containsSubstring("hello", "xyz") {
		t.Error("expected false")
	}
}

func TestContainsAnyFold(t *testing.T) {
	t.Parallel()
	if !containsAnyFold([]string{"apple", "banana"}, "BAN") {
		t.Error("expected true")
	}
	if containsAnyFold([]string{"apple", "banana"}, "xyz") {
		t.Error("expected false")
	}
}

func TestTemplate_WebScraper(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := r.NewWebScraperTemplate()
	if tmpl.ID != "web_scraper" {
		t.Errorf("expected web_scraper, got %s", tmpl.ID)
	}
	if len(tmpl.Nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(tmpl.Nodes))
	}
	if len(tmpl.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(tmpl.Edges))
	}
}

func TestTemplate_APITest(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := r.NewAPITestTemplate()
	if tmpl.ID != "api_test" {
		t.Errorf("expected api_test, got %s", tmpl.ID)
	}
}

func TestTemplate_DataPipeline(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := r.NewDataPipelineTemplate()
	if tmpl.ID != "data_pipeline" {
		t.Errorf("expected data_pipeline, got %s", tmpl.ID)
	}
}

func TestTemplate_ContentGenerator(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := r.NewContentGeneratorTemplate()
	if tmpl.ID != "content_generator" {
		t.Errorf("expected content_generator, got %s", tmpl.ID)
	}
}

func TestTemplate_AutomationWorkflow(t *testing.T) {
	t.Parallel()
	r := NewRegistry("")
	tmpl := r.NewAutomationWorkflowTemplate()
	if tmpl.ID != "automation_workflow" {
		t.Errorf("expected automation_workflow, got %s", tmpl.ID)
	}
}
