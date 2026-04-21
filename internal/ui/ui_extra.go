package ui

type CustomTheme struct {
	Name   string
	Colors map[string]string
	FontSize int
}

func NewCustomTheme(name string) *CustomTheme {
	return &CustomTheme{
		Name:   name,
		Colors: make(map[string]string),
		FontSize: 14,
	}
}

type DesignTokens struct {
	Primary   string
	Secondary string
	Accent   string
	Spacing  map[string]int
	Font     map[string]string
}

func GetDesignTokens() *DesignTokens {
	return &DesignTokens{
		Primary:   "#007acc",
		Secondary: "#6c757d",
		Accent:    "#28a745",
		Spacing: map[string]int{
			"xs": 4,
			"sm": 8,
			"md": 16,
			"lg": 24,
			"xl": 32,
		},
		Font: map[string]string{
			"family": "monospace",
			"size":   "14px",
		},
	}
}

type RichCLI struct{}

func NewRichCLI() *RichCLI {
	return &RichCLI{}
}

func (r *RichCLI) Format(text, style string) string {
	styles := map[string]string{
		"bold":   "\033[1m" + text + "\033[0m",
		"italic": "\033[3m" + text + "\033[0m",
		"dim":    "\033[2m" + text + "\033[0m",
	}
	if s, ok := styles[style]; ok {
		return s
	}
	return text
}

func (r *RichCLI) Table(headers []string, rows [][]string) string {
	return "table output"
}

type Animation struct {
	Type     string
	Duration int
}

func NewAnimation(animType string) *Animation {
	return &Animation{
		Type:     animType,
		Duration: 300,
	}
}

func (a *Animation) Render() string {
	return "animated: " + a.Type
}