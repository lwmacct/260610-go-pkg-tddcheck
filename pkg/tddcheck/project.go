package tddcheck

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

// Project is an AST index for a Go module or package tree.
type Project struct {
	Root         string
	Files        []*GoFile
	ChangedFiles map[string]bool
	Packages     map[string]*Package
}

// GoFile contains parsed data for one Go file.
type GoFile struct {
	Path        string
	AbsPath     string
	Package     string
	IsTest      bool
	IsGenerated bool
	FileSet     *token.FileSet
	AST         *ast.File
	Decls       []Decl
	Tests       []TestFunc
}

// Decl describes one top-level declaration relevant to TDD checks.
type Decl struct {
	Name     string
	Recv     string
	Kind     DeclKind
	Exported bool
	File     string
	Line     int
}

// DeclKind describes declaration categories.
type DeclKind string

const (
	DeclFunc   DeclKind = "func"
	DeclMethod DeclKind = "method"
	DeclType   DeclKind = "type"
)

// TestFunc describes one Go test, benchmark, fuzz, or example function.
type TestFunc struct {
	Name    string
	Kind    TestKind
	File    string
	Line    int
	Empty   bool
	Skipped bool
}

// TestKind describes Go test-like function kinds.
type TestKind string

const (
	TestKindTest      TestKind = "test"
	TestKindBenchmark TestKind = "benchmark"
	TestKindFuzz      TestKind = "fuzz"
	TestKindExample   TestKind = "example"
)

// Package groups parsed files by import directory and package name.
type Package struct {
	Dir       string
	Name      string
	Files     []*GoFile
	TestNames map[string]TestFunc
	Decls     []Decl
}

func (p *Project) packageForFile(file *GoFile) *Package {
	dir := filepath.ToSlash(filepath.Dir(file.Path))
	if dir == "." {
		dir = ""
	}
	key := dir + "|" + file.Package
	pkg := p.Packages[key]
	if pkg == nil {
		pkg = &Package{
			Dir:       dir,
			Name:      file.Package,
			TestNames: make(map[string]TestFunc),
		}
		p.Packages[key] = pkg
	}

	return pkg
}

func normalizeRel(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func isGoTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}
