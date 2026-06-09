package tddcheck

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitStagedFiles returns project-relative files staged for commit.
func GitStagedFiles(ctx context.Context, root string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("git staged files: %w: %s", err, msg)
		}

		return nil, fmt.Errorf("git staged files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, normalizeRel(line))
		}
	}

	return files, nil
}
