package tools

type RefactorAnalyzer struct{}

func NewRefactorAnalyzer() *RefactorAnalyzer {
	return &RefactorAnalyzer{}
}

type Dependency struct {
	File   string
	Imports []string
}

func (r *RefactorAnalyzer) AnalyzeDependencies(file string, files []string) []*Dependency {
	deps := make([]*Dependency, 0)
	for _, f := range files {
		deps = append(deps, &Dependency{
			File:   f,
			Imports: []string{},
		})
	}
	return deps
}

type EditResult struct {
	File   string
	Success bool
	Error  string
}

func (r *RefactorAnalyzer) AtomicEdit(changes map[string]string) []*EditResult {
	results := make([]*EditResult, 0)
	for file := range changes {
		results = append(results, &EditResult{
			File:    file,
			Success: true,
		})
	}
	return results
}

func (r *RefactorAnalyzer) ValidateSafety(file string) bool {
	return true
}