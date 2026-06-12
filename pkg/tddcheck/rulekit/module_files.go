package rulekit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func ModuleFiles(root string, ruleName string, config Config, match func(string) bool) ([]string, error) {
	roots, err := ModuleScanRoots(root, ruleName, config)
	if err != nil {
		return nil, err
	}
	config = config.WithDefaults()

	var matches []string
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if ShouldSkipModuleScanDir(entry.Name(), config) {
					return filepath.SkipDir
				}
				return nil
			}
			if match(filepath.Base(path)) {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.Sort(matches)
	return matches, nil
}

func ModulePackageDirs(root string, ruleName string, config Config) ([]string, error) {
	roots, err := ModuleScanRoots(root, ruleName, config)
	if err != nil {
		return nil, err
	}
	config = config.WithDefaults()

	seen := make(map[string]struct{})
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if ShouldSkipModuleScanDir(entry.Name(), config) {
					return filepath.SkipDir
				}
				return nil
			}
			if filepath.Ext(path) == ".go" {
				seen[filepath.Dir(path)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	dirs := make([]string, 0, len(seen))
	for dir := range seen {
		dirs = append(dirs, dir)
	}
	slices.Sort(dirs)
	return dirs, nil
}

func ModuleScanRoots(root string, ruleName string, config Config) ([]string, error) {
	resolved, err := ResolveRuleRoot(root, ruleName)
	if err != nil {
		return nil, err
	}
	config = config.WithDefaults()

	var layered []string
	for _, layer := range config.LayerDirs {
		dir := filepath.Join(resolved, layer)
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			layered = append(layered, dir)
		}
	}
	if len(layered) > 0 {
		return layered, nil
	}
	return []string{resolved}, nil
}

func ResolveRuleRoot(root string, ruleName string) (string, error) {
	if root == "" {
		return "", errors.New(ruleName + ".Root is empty")
	}
	if filepath.IsAbs(root) {
		return root, nil
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, root), nil
}

func ShouldSkipModuleScanDir(name string, config Config) bool {
	config = config.WithDefaults()
	if strings.HasPrefix(name, ".") {
		return true
	}
	return StringIn(name, config.SkipDirs)
}

func DisplayFilename(filename string) string {
	projectRoot, err := FindProjectRoot()
	if err == nil {
		if relative, relErr := filepath.Rel(projectRoot, filename); relErr == nil && !strings.HasPrefix(relative, "..") {
			return filepath.ToSlash(relative)
		}
	}

	return filepath.ToSlash(filename)
}

func FindProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		wd = parent
	}
}

func ModulePath() (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	// #nosec G304 -- projectRoot is discovered by walking to the repository go.mod.
	data, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				return value, nil
			}
		}
	}
	return "", errors.New("module path not found in go.mod")
}
