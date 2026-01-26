package commits

// CategorySuggestion represents a suggested changelog category.
type CategorySuggestion struct {
	Category   string  // Added, Fixed, Changed, etc.
	Tier       string  // core, standard, extended, optional
	Confidence float64 // 0.0 - 1.0
	Reasoning  string  // Why this category was chosen
}

// categoryMap maps conventional commit types to changelog categories.
var categoryMap = map[string]CategorySuggestion{
	"feat":     {Category: "Added", Tier: "core", Confidence: 0.95, Reasoning: "feat maps directly to Added"},
	"fix":      {Category: "Fixed", Tier: "core", Confidence: 0.95, Reasoning: "fix maps directly to Fixed"},
	"docs":     {Category: "Changed", Tier: "standard", Confidence: 0.8, Reasoning: "documentation is a change"},
	"style":    {Category: "Changed", Tier: "extended", Confidence: 0.7, Reasoning: "style changes are cosmetic"},
	"refactor": {Category: "Changed", Tier: "standard", Confidence: 0.85, Reasoning: "refactoring changes behavior preservation"},
	"perf":     {Category: "Changed", Tier: "standard", Confidence: 0.85, Reasoning: "performance improvement is a change"},
	"test":     {Category: "Changed", Tier: "extended", Confidence: 0.6, Reasoning: "test changes support code quality"},
	"build":    {Category: "Changed", Tier: "extended", Confidence: 0.6, Reasoning: "build system change"},
	"ci":       {Category: "Changed", Tier: "optional", Confidence: 0.5, Reasoning: "CI is infrastructure"},
	"chore":    {Category: "Changed", Tier: "optional", Confidence: 0.5, Reasoning: "chore is maintenance"},
	"revert":   {Category: "Removed", Tier: "core", Confidence: 0.8, Reasoning: "revert removes previous change"},
}

// SuggestCategory maps a conventional commit to a changelog category.
func SuggestCategory(cc *ConventionalCommit) *CategorySuggestion {
	if cc == nil {
		return nil
	}

	if cc.Breaking {
		return &CategorySuggestion{
			Category:   "Changed",
			Tier:       "core",
			Confidence: 1.0,
			Reasoning:  "breaking change is always a significant change",
		}
	}

	if suggestion, ok := categoryMap[cc.Type]; ok {
		s := suggestion // copy
		return &s
	}

	return &CategorySuggestion{
		Category:   "Changed",
		Tier:       "optional",
		Confidence: 0.3,
		Reasoning:  "unknown commit type, defaulting to Changed",
	}
}
