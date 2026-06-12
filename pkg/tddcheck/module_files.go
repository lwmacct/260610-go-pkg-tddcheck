package tddcheck

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var moduleLayerDirs = []string{"domain", "usecase", "adapter", "runtime", "infra"}

func moduleFiles(root string, ruleName string, match func(string) bool) ([]string, error) {
	roots, err := moduleScanRoots(root, ruleName)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipModuleScanDir(entry.Name()) {
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

func modulePackageDirs(root string, ruleName string) ([]string, error) {
	roots, err := moduleScanRoots(root, ruleName)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipModuleScanDir(entry.Name()) {
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

func moduleScanRoots(root string, ruleName string) ([]string, error) {
	resolved, err := resolveRuleRoot(root, ruleName)
	if err != nil {
		return nil, err
	}

	var layered []string
	for _, layer := range moduleLayerDirs {
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

func resolveRuleRoot(root string, ruleName string) (string, error) {
	if root == "" {
		return "", errors.New(ruleName + ".Root is empty")
	}
	if filepath.IsAbs(root) {
		return root, nil
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(projectRoot, root), nil
}

func shouldSkipModuleScanDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules"
}
