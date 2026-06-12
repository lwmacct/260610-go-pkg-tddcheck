package tddcheck

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ModuleLayerRules declares import direction rules for layered internal packages.
type ModuleLayerRules struct {
	// Root is the layered module root directory. Relative paths are resolved from go.mod.
	Root   string
	Config Config
}

type LayerDependencyViolation struct {
	File       string
	Line       int
	ImportPath string
	Message    string
}

func (r ModuleLayerRules) AssertLayerDependencies(t *testing.T) {
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

func (r ModuleLayerRules) LayerDependencyViolations() ([]LayerDependencyViolation, error) {
	config := r.Config.withDefaults()
	root, err := resolveRuleRoot(r.Root, "ModuleLayerRules")
	if err != nil {
		return nil, err
	}
	modulePath, err := modulePath()
	if err != nil {
		return nil, err
	}
	internalPrefix := modulePath + "/internal/"

	files, err := moduleFilesWithConfig(r.Root, "ModuleLayerRules", config, func(name string) bool {
		return strings.HasSuffix(name, ".go")
	})
	if err != nil {
		return nil, err
	}

	var violations []LayerDependencyViolation
	for _, file := range files {
		fileViolations, err := layerDependencyViolationsInFile(root, internalPrefix, config, file)
		if err != nil {
			return nil, err
		}
		violations = append(violations, fileViolations...)
	}
	return violations, nil
}

func layerDependencyViolationsInFile(root string, internalPrefix string, config Config, filename string) ([]LayerDependencyViolation, error) {
	rel, err := filepath.Rel(root, filename)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return nil, nil
	}
	sourceLayer := parts[0]
	if !slices.Contains(config.LayerDirs, sourceLayer) {
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
		targetLayer, targetRel, ok := internalImportLayer(internalPrefix, importPath)
		if !ok {
			continue
		}
		if message, invalid := invalidLayerDependency(config, sourceLayer, targetLayer, targetRel); invalid {
			position := fileSet.Position(importSpec.Pos())
			violations = append(violations, LayerDependencyViolation{
				File:       displayFilename(filename),
				Line:       position.Line,
				ImportPath: importPath,
				Message:    message,
			})
		}
	}
	return violations, nil
}

func internalImportLayer(internalPrefix string, importPath string) (string, string, bool) {
	if !strings.HasPrefix(importPath, internalPrefix) {
		return "", "", false
	}
	rel := strings.TrimPrefix(importPath, internalPrefix)
	layer, _, ok := strings.Cut(rel, "/")
	if !ok {
		return "", "", false
	}
	return layer, rel, true
}

func invalidLayerDependency(config Config, sourceLayer string, targetLayer string, targetRel string) (string, bool) {
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

func modulePath() (string, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	// #nosec G304 -- projectRoot is discovered by walking to the repository go.mod.
	data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				return value, nil
			}
		}
	}
	return "", errors.New("module path not found in go.mod")
}
