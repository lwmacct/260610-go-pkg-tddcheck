package layerdeps

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
)

// Rules declares import direction rules for layered internal packages.
type Rules struct {
	root   string
	config rulekit.Config
}

// New creates rules for the supplied module root.
func New(root string, options ...rulekit.Option) Rules {
	values := rulekit.NewRuleOptions(root, options...)
	return Rules{root: values.Root, config: values.Config}
}

type LayerDependencyViolation struct {
	File       string
	Line       int
	ImportPath string
	Message    string
}

func (r Rules) Assert(t *testing.T) {
	t.Helper()

	violations, err := r.LayerDependencyViolations()
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) == 0 {
		return
	}

	lines := make([]string, 0, len(violations))
	for _, violation := range violations {
		lines = append(lines, fmt.Sprintf(
			"%s:%d: %s: %s",
			violation.File,
			violation.Line,
			violation.Message,
			violation.ImportPath,
		))
	}
	t.Fatalf("invalid layer dependencies:\n  - %s", strings.Join(lines, "\n  - "))
}

func (r Rules) LayerDependencyViolations() ([]LayerDependencyViolation, error) {
	config := r.config.WithDefaults()
	root, err := rulekit.ResolveRuleRoot(r.root, "Rules")
	if err != nil {
		return nil, err
	}
	modulePath, err := rulekit.ModulePath()
	if err != nil {
		return nil, err
	}
	importPrefixes := moduleImportPrefixes(modulePath, root)

	files, err := rulekit.ModuleFiles(r.root, "Rules", config, func(name string) bool {
		return strings.HasSuffix(name, ".go")
	})
	if err != nil {
		return nil, err
	}

	var violations []LayerDependencyViolation
	for _, file := range files {
		fileViolations, err := layerDependencyViolationsInFile(root, importPrefixes, config, file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}
	return violations, nil
}

func layerDependencyViolationsInFile(root string, importPrefixes []string, config rulekit.Config, filename string) ([]LayerDependencyViolation, error) {
	rel, err := filepath.Rel(root, filename)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	sourceLayer, ok := sourceLayerFromRel(config, parts)
	if !ok {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fileSet, filename, nil, parser.ImportsOnly|parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var violations []LayerDependencyViolation
	for _, importSpec := range parsedFile.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		targetLayer, targetRel, ok := moduleImportLayer(importPrefixes, importPath)
		if !ok {
			continue
		}
		if message, invalid := invalidLayerDependency(config, sourceLayer, targetLayer, targetRel); invalid {
			position := fileSet.Position(importSpec.Pos())
			violations = append(violations, LayerDependencyViolation{
				File:       rulekit.DisplayFilename(filename),
				Line:       position.Line,
				ImportPath: importPath,
				Message:    message,
			})
		}
	}
	return violations, nil
}

func sourceLayerFromRel(config rulekit.Config, parts []string) (string, bool) {
	for _, part := range parts {
		if slices.Contains(config.LayerDirs, part) {
			return part, true
		}
	}
	return "", false
}

func moduleImportPrefixes(modulePath string, root string) []string {
	prefixes := []string{modulePath + "/internal/"}
	projectRoot, err := rulekit.FindProjectRoot()
	if err != nil {
		return prefixes
	}
	rel, err := filepath.Rel(projectRoot, root)
	if err != nil || strings.HasPrefix(rel, "..") {
		return prefixes
	}
	var prefix string
	if rel == "." {
		prefix = modulePath + "/"
	} else {
		prefix = modulePath + "/" + filepath.ToSlash(rel) + "/"
	}
	if !slices.Contains(prefixes, prefix) {
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func moduleImportLayer(importPrefixes []string, importPath string) (string, string, bool) {
	for _, prefix := range importPrefixes {
		if !strings.HasPrefix(importPath, prefix) {
			continue
		}
		rel := strings.TrimPrefix(importPath, prefix)
		layer, _, ok := strings.Cut(rel, "/")
		if !ok {
			return "", "", false
		}
		return layer, rel, true
	}
	return "", "", false
}

func invalidLayerDependency(config rulekit.Config, sourceLayer string, targetLayer string, targetRel string) (string, bool) {
	for _, rule := range config.LayerRules {
		if sourceLayer != rule.SourceLayer || targetLayer != rule.TargetLayer {
			continue
		}
		if rule.TargetRelPrefix != "" && !strings.HasPrefix(targetRel, rule.TargetRelPrefix) {
			continue
		}
		message := rule.Message
		if message == "" {
			message = sourceLayer + " must not import " + targetLayer
		}
		return message, true
	}
	return "", false
}
