package tddcheck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func findModuleRoot(callerSkip int) (string, error) {
	_, file, _, ok := runtime.Caller(callerSkip)
	if !ok {
		return "", errors.New("locate caller")
	}

	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat go.mod: %w", err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", filepath.Dir(file))
		}
		dir = parent
	}
}
