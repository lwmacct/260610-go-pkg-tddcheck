package tddcheck

import "slices"

// Config describes project-specific architecture policy knobs.
//
// A zero Config uses DefaultConfig. Slice fields distinguish nil from empty:
// nil means use the default value, while an empty non-nil slice disables that
// portion of the policy.
type Config struct {
	LayerDirs  []string
	SkipDirs   []string
	LayerRules []LayerDependencyRule

	MapperForbiddenImports        []string
	HandlerForbiddenImports       []string
	HandlerForbiddenCalls         []string
	RepositoryForbiddenImports    []string
	RepositoryImplementationNames []string

	ValidationResolve ValidationResolveConfig
	DatabaseTest      DatabaseTestConfig
}

// LayerDependencyRule rejects imports from SourceLayer to TargetLayer. When
// TargetRelPrefix is set, only matching target relative import paths are rejected.
type LayerDependencyRule struct {
	SourceLayer     string
	TargetLayer     string
	TargetRelPrefix string
	Message         string
}

// ValidationResolveConfig configures the accepted validation.go Resolve method
// signature used by protocol frameworks such as Huma.
type ValidationResolveConfig struct {
	ContextPackage    string
	ContextType       string
	PathBufferPackage string
	PathBufferType    string
	Message           string
}

// DatabaseTestConfig configures the optional database test boundary rule.
type DatabaseTestConfig struct {
	AllowedPaths      []string
	OpenNeedle        string
	TempDirNeedle     string
	ConfigPathNeedle  string
	OpenMessage       string
	ConfigPathMessage string
}

// DefaultConfig returns the architecture policy previously used by this package.
func DefaultConfig() Config {
	return Config{
		LayerDirs: []string{"domain", "usecase", "adapter", "runtime", "infra"},
		SkipDirs:  []string{".git", ".hg", ".svn", "vendor", "node_modules", "dist", "build"},
		LayerRules: []LayerDependencyRule{
			{
				SourceLayer: "domain",
				TargetLayer: "adapter",
				Message:     "domain must not import adapter",
			},
			{
				SourceLayer: "usecase",
				TargetLayer: "adapter",
				Message:     "usecase must not import adapter",
			},
			{
				SourceLayer:     "runtime",
				TargetLayer:     "adapter",
				TargetRelPrefix: "adapter/httpauth",
				Message:         "runtime must not import HTTP API adapter",
			},
			{
				SourceLayer: "infra",
				TargetLayer: "domain",
				Message:     "infra must not import business layers",
			},
			{
				SourceLayer: "infra",
				TargetLayer: "usecase",
				Message:     "infra must not import business layers",
			},
			{
				SourceLayer: "infra",
				TargetLayer: "adapter",
				Message:     "infra must not import business layers",
			},
		},
		MapperForbiddenImports: []string{
			"context",
			"database/sql",
			"net/http",
			"github.com/danielgtaylor/huma/v2",
			"github.com/uptrace/bun",
			"gorm.io/gorm",
		},
		HandlerForbiddenImports: []string{
			"github.com/uptrace/bun",
			"gorm.io/gorm",
		},
		HandlerForbiddenCalls: []string{
			"NewSelect",
			"NewInsert",
			"NewUpdate",
			"NewDelete",
			"RunInTx",
			"BeginTx",
		},
		RepositoryForbiddenImports: []string{
			"net/http",
			"github.com/danielgtaylor/huma/v2",
			"github.com/coder/websocket",
		},
		RepositoryImplementationNames: []string{"bunRepository"},
		ValidationResolve: ValidationResolveConfig{
			ContextPackage:    "huma",
			ContextType:       "Context",
			PathBufferPackage: "huma",
			PathBufferType:    "PathBuffer",
			Message:           "validation.go Resolve receiver method must implement huma.Resolver or huma.ResolverWithPath",
		},
		DatabaseTest: DatabaseTestConfig{
			AllowedPaths: []string{
				"internal/infra/database/database_test.go",
				"internal/tddcheck/database_boundary_test.go",
			},
			OpenNeedle:        "database.OpenSQLite",
			TempDirNeedle:     "filepath.Join(t.TempDir()",
			ConfigPathNeedle:  "cfg.Server.DB.SQLite = filepath.Join(t.TempDir()",
			OpenMessage:       "ordinary SQLite tests must use dbtest.Open",
			ConfigPathMessage: "ordinary SQLite config tests must use dbtest.Open or explicit test exemption",
		},
	}
}

func (c Config) withDefaults() Config {
	defaults := DefaultConfig()
	if c.LayerDirs == nil {
		c.LayerDirs = defaults.LayerDirs
	}
	if c.SkipDirs == nil {
		c.SkipDirs = defaults.SkipDirs
	}
	if c.LayerRules == nil {
		c.LayerRules = defaults.LayerRules
	}
	if c.MapperForbiddenImports == nil {
		c.MapperForbiddenImports = defaults.MapperForbiddenImports
	}
	if c.HandlerForbiddenImports == nil {
		c.HandlerForbiddenImports = defaults.HandlerForbiddenImports
	}
	if c.HandlerForbiddenCalls == nil {
		c.HandlerForbiddenCalls = defaults.HandlerForbiddenCalls
	}
	if c.RepositoryForbiddenImports == nil {
		c.RepositoryForbiddenImports = defaults.RepositoryForbiddenImports
	}
	if c.RepositoryImplementationNames == nil {
		c.RepositoryImplementationNames = defaults.RepositoryImplementationNames
	}
	if c.ValidationResolve.ContextPackage == "" {
		c.ValidationResolve.ContextPackage = defaults.ValidationResolve.ContextPackage
	}
	if c.ValidationResolve.ContextType == "" {
		c.ValidationResolve.ContextType = defaults.ValidationResolve.ContextType
	}
	if c.ValidationResolve.PathBufferPackage == "" {
		c.ValidationResolve.PathBufferPackage = defaults.ValidationResolve.PathBufferPackage
	}
	if c.ValidationResolve.PathBufferType == "" {
		c.ValidationResolve.PathBufferType = defaults.ValidationResolve.PathBufferType
	}
	if c.ValidationResolve.Message == "" {
		c.ValidationResolve.Message = defaults.ValidationResolve.Message
	}
	if c.DatabaseTest.AllowedPaths == nil {
		c.DatabaseTest.AllowedPaths = defaults.DatabaseTest.AllowedPaths
	}
	if c.DatabaseTest.OpenNeedle == "" {
		c.DatabaseTest.OpenNeedle = defaults.DatabaseTest.OpenNeedle
	}
	if c.DatabaseTest.TempDirNeedle == "" {
		c.DatabaseTest.TempDirNeedle = defaults.DatabaseTest.TempDirNeedle
	}
	if c.DatabaseTest.ConfigPathNeedle == "" {
		c.DatabaseTest.ConfigPathNeedle = defaults.DatabaseTest.ConfigPathNeedle
	}
	if c.DatabaseTest.OpenMessage == "" {
		c.DatabaseTest.OpenMessage = defaults.DatabaseTest.OpenMessage
	}
	if c.DatabaseTest.ConfigPathMessage == "" {
		c.DatabaseTest.ConfigPathMessage = defaults.DatabaseTest.ConfigPathMessage
	}
	return c
}

func stringIn(value string, values []string) bool {
	return slices.Contains(values, value)
}
