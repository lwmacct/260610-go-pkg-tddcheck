package tddcheck

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ScanOptions configures project scanning.
type ScanOptions struct {
	Root        string
	StagedOnly  bool
	IgnoreGlobs []string
}

// Scan parses Go files under a root and returns a project index.
func Scan(ctx context.Context, opts ScanOptions) (*Project, error) {
	root := opts.Root
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	changed, err := changedFiles(ctx, absRoot, opts.StagedOnly)
	if err != nil {
		return nil, err
	}

	project := &Project{
		Root:         absRoot,
		ChangedFiles: changed,
		Packages:     make(map[string]*Package),
	}

	if err := filepath.WalkDir(absRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		rel = normalizeRel(rel)

		if entry.IsDir() {
			if shouldSkipDir(rel) || matchAnyGlob(opts.IgnoreGlobs, rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		if matchAnyGlob(opts.IgnoreGlobs, rel) {
			return nil
		}

		file, err := parseGoFile(path, rel)
		if err != nil {
			return err
		}
		project.Files = append(project.Files, file)

		pkg := project.packageForFile(file)
		pkg.Files = append(pkg.Files, file)
		pkg.Decls = append(pkg.Decls, file.Decls...)
		for _, test := range file.Tests {
			pkg.TestNames[test.Name] = test
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("scan go files: %w", err)
	}

	return project, nil
}

func changedFiles(ctx context.Context, root string, stagedOnly bool) (map[string]bool, error) {
	if !stagedOnly {
		return nil, nil
	}

	files, err := GitStagedFiles(ctx, root)
	if err != nil {
		return nil, err
	}
	changed := make(map[string]bool, len(files))
	for _, file := range files {
		changed[normalizeRel(file)] = true
	}

	return changed, nil
}

func shouldSkipDir(rel string) bool {
	switch rel {
	case ".git", ".hg", ".svn", "vendor":
		return true
	default:
		return strings.HasPrefix(rel, ".git/") ||
			strings.HasPrefix(rel, "vendor/") ||
			strings.Contains(rel, "/vendor/")
	}
}

func parseGoFile(absPath, relPath string) (*GoFile, error) {
	content, err := os.ReadFile(absPath) //nolint:gosec // scanner reads project source files
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", relPath, err)
	}

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, absPath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", relPath, err)
	}

	file := &GoFile{
		Path:        relPath,
		AbsPath:     absPath,
		Package:     astFile.Name.Name,
		IsTest:      isGoTestFile(relPath),
		IsGenerated: isGenerated(content),
		FileSet:     fileSet,
		AST:         astFile,
	}

	if !file.IsGenerated {
		file.Decls = collectDecls(file)
		file.Tests = collectTests(file)
	}

	return file, nil
}

func isGenerated(content []byte) bool {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for i := 0; scanner.Scan() && i < 20; i++ {
		line := scanner.Text()
		if strings.Contains(line, "Code generated") && strings.Contains(line, "DO NOT EDIT") {
			return true
		}
	}

	return false
}

func collectDecls(file *GoFile) []Decl {
	var decls []Decl
	for _, decl := range file.AST.Decls {
		switch typed := decl.(type) {
		case *ast.FuncDecl:
			kind := DeclFunc
			recv := ""
			if typed.Recv != nil && len(typed.Recv.List) > 0 {
				kind = DeclMethod
				recv = receiverName(typed.Recv.List[0].Type)
			}
			decls = append(decls, Decl{
				Name:     typed.Name.Name,
				Recv:     recv,
				Kind:     kind,
				Exported: typed.Name.IsExported(),
				File:     file.Path,
				Line:     file.FileSet.Position(typed.Pos()).Line,
			})
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				decls = append(decls, Decl{
					Name:     typeSpec.Name.Name,
					Kind:     DeclType,
					Exported: typeSpec.Name.IsExported(),
					File:     file.Path,
					Line:     file.FileSet.Position(typeSpec.Pos()).Line,
				})
			}
		}
	}

	return decls
}

func receiverName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverName(typed.X)
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	default:
		return ""
	}
}

func collectTests(file *GoFile) []TestFunc {
	if !file.IsTest {
		return nil
	}

	var tests []TestFunc
	for _, decl := range file.AST.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv != nil {
			continue
		}
		kind, ok := testKind(funcDecl.Name.Name)
		if !ok {
			continue
		}
		tests = append(tests, TestFunc{
			Name:    funcDecl.Name.Name,
			Kind:    kind,
			File:    file.Path,
			Line:    file.FileSet.Position(funcDecl.Pos()).Line,
			Empty:   isEmptyTest(funcDecl),
			Skipped: hasSkipCall(funcDecl),
		})
	}

	return tests
}

func testKind(name string) (TestKind, bool) {
	switch {
	case strings.HasPrefix(name, "Test"):
		return TestKindTest, true
	case strings.HasPrefix(name, "Benchmark"):
		return TestKindBenchmark, true
	case strings.HasPrefix(name, "Fuzz"):
		return TestKindFuzz, true
	case strings.HasPrefix(name, "Example"):
		return TestKindExample, true
	default:
		return "", false
	}
}

func isEmptyTest(fn *ast.FuncDecl) bool {
	return fn.Body == nil || len(fn.Body.List) == 0
}

func hasSkipCall(fn *ast.FuncDecl) bool {
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch selector.Sel.Name {
		case "Skip", "Skipf", "SkipNow":
			found = true
			return false
		default:
			return true
		}
	})

	return found
}
