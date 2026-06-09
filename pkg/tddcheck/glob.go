package tddcheck

import (
	"path"
	"strings"
)

func matchAnyGlob(patterns []string, rel string) bool {
	rel = normalizeRel(rel)
	for _, pattern := range patterns {
		if matchGlob(pattern, rel) {
			return true
		}
	}

	return false
}

func matchGlob(pattern, rel string) bool {
	pattern = strings.TrimSpace(path.Clean(strings.ReplaceAll(pattern, "\\", "/")))
	rel = path.Clean(strings.ReplaceAll(rel, "\\", "/"))
	if pattern == "." || pattern == "" {
		return false
	}
	if pattern == rel {
		return true
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return rel == prefix || strings.HasPrefix(rel, prefix+"/")
	}
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if ok, _ := path.Match(suffix, rel); ok {
			return true
		}
		return strings.HasSuffix(rel, "/"+suffix)
	}
	if strings.Contains(pattern, "/**/") {
		parts := strings.Split(pattern, "/**/")
		return len(parts) == 2 &&
			(strings.HasPrefix(rel, parts[0]+"/") || rel == parts[0]) &&
			(strings.HasSuffix(rel, "/"+parts[1]) || rel == parts[1])
	}
	ok, _ := path.Match(pattern, rel)

	return ok
}
