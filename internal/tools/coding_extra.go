package tools

type CodeReviewer struct{}

func NewCodeReviewer() *CodeReviewer {
	return &CodeReviewer{}
}

type Issue struct {
	Severity string
	Message  string
	Line     int
}

func (cr *CodeReviewer) Analyze(code string) []*Issue {
	issues := make([]*Issue, 0)

	if len(code) == 0 {
		issues = append(issues, &Issue{Severity: "warning", Message: "empty code", Line: 0})
	}

	if len(code) > 1000 {
		issues = append(issues, &Issue{Severity: "info", Message: "large function", Line: 0})
	}

	return issues
}

func (cr *CodeReviewer) CheckStyle(code string) []*Issue {
	return cr.Analyze(code)
}

type TestGenerator struct{}

func NewTestGenerator() *TestGenerator {
	return &TestGenerator{}
}

func (tg *TestGenerator) Generate(funcCode string) string {
	return "package main\n\nimport \"testing\"\n\nfunc TestFunction(t *testing.T) {\n\tt.Log(\"test\")\n}\n"
}

type CodeExplainer struct{}

func NewCodeExplainer() *CodeExplainer {
	return &CodeExplainer{}
}

func (ce *CodeExplainer) Explain(symbol string) string {
	explanations := map[string]string{
		"fmt.Println": "Prints to stdout with newline",
		"fmt.Printf":  "Formatted print",
		"os.Exit":     "Exits the program",
	}
	if exp, ok := explanations[symbol]; ok {
		return exp
	}
	return "Unknown symbol"
}

type Diagnostics struct{}

func NewDiagnostics() *Diagnostics {
	return &Diagnostics{}
}

func (d *Diagnostics) Run(file string) []*Issue {
	return []*Issue{
		{Severity: "info", Message: "File analyzed: " + file, Line: 0},
	}
}

func (d *Diagnostics) GetLSPDiagnostics(file string) []*Issue {
	return d.Run(file)
}