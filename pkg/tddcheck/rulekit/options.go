package rulekit

// RuleOptions are the common construction options for an individual rule package.
type RuleOptions struct {
	Root   string
	Config Config
}

// Option customizes an individual rule package constructor.
type Option func(*RuleOptions)

// WithConfig applies project-specific policy to a rule.
func WithConfig(config Config) Option {
	return func(options *RuleOptions) {
		options.Config = config
	}
}

// NewRuleOptions applies common rule construction options.
func NewRuleOptions(root string, options ...Option) RuleOptions {
	values := RuleOptions{Root: root}
	for _, option := range options {
		option(&values)
	}
	return values
}
