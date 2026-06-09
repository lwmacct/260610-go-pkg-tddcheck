package tddcheck

// Option configures a Check run.
type Option func(*config)

type config struct {
	root        string
	stagedOnly  bool
	rules       []Rule
	ignoreGlobs []string
}

func defaultConfig() config {
	return config{
		root:  ".",
		rules: defaultRules(),
		ignoreGlobs: []string{
			"vendor/**",
			"testdata/**",
			"**/testdata/**",
		},
	}
}

func defaultRules() []Rule {
	return []Rule{
		RuleChangedCodeNeedsTest(),
		RuleExportedDeclsNeedTests(),
		RuleNoSkippedOrEmptyTests(),
	}
}

// WithRoot sets the project root. The default is the current directory.
func WithRoot(root string) Option {
	return func(c *config) {
		if root != "" {
			c.root = root
		}
	}
}

// WithStagedOnly makes checks focus on files staged in git.
func WithStagedOnly(enabled bool) Option {
	return func(c *config) {
		c.stagedOnly = enabled
	}
}

// WithRules replaces the enabled rule set.
func WithRules(rules ...Rule) Option {
	return func(c *config) {
		c.rules = append([]Rule(nil), rules...)
	}
}

// WithDefaultRules replaces the enabled rules with the package default rules.
func WithDefaultRules() Option {
	return func(c *config) {
		c.rules = defaultRules()
	}
}

// WithIgnoreGlobs replaces the ignore glob list. Glob patterns use slash
// separated project-relative paths and support a simple ** segment.
func WithIgnoreGlobs(patterns ...string) Option {
	return func(c *config) {
		c.ignoreGlobs = append([]string(nil), patterns...)
	}
}
