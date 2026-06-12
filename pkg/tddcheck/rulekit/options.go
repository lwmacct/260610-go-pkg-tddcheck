package rulekit

// RuleOptions are the common construction options for an individual rule package.
type RuleOptions struct {
	Root   string
	Config Config
}

// Option customizes an individual rule package constructor.
type Option func(*RuleOptions)

// WithConfig 将项目专属策略应用到规则。
func WithConfig(config Config) Option {
	return func(options *RuleOptions) {
		options.Config = config
	}
}

// NewRuleOptions 应用通用规则构造选项。
func NewRuleOptions(root string, options ...Option) RuleOptions {
	values := RuleOptions{Root: root}
	for _, option := range options {
		option(&values)
	}
	return values
}
