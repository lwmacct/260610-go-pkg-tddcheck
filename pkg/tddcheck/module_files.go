package tddcheck

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func moduleFilesWithConfig(root string, ruleName string, config Config, match func(string) bool) ([]string, error) {
	roots, err := moduleScanRootsWithConfig(root, ruleName, config)
	if err != nil {
		return nil, err
	}
	config = config.withDefaults()

	var matches []string
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipModuleScanDirWithConfig(entry.Name(), config) {
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

func modulePackageDirsWithConfig(root string, ruleName string, config Config) ([]string, error) {
	roots, err := moduleScanRootsWithConfig(root, ruleName, config)
	if err != nil {
		return nil, err
	}
	config = config.withDefaults()

	seen := make(map[string]struct{})
	for _, scanRoot := range roots {
		err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipModuleScanDirWithConfig(entry.Name(), config) {
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

func moduleScanRootsWithConfig(root string, ruleName string, config Config) ([]string, error) {
	resolved, err := resolveRuleRoot(root, ruleName)
	if err != nil {
		return nil, err
	}
	config = config.withDefaults()

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

func shouldSkipModuleScanDirWithConfig(name string, config Config) bool {
	config = config.withDefaults()
	if strings.HasPrefix(name, ".") {
		return true
	}
	return stringIn(name, config.SkipDirs)
}
