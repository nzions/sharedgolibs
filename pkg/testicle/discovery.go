package testicle

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// TestInfo represents information about a discovered test
type TestInfo struct {
	Name     string
	Package  string
	File     string
	Line     int
	Function *ast.FuncDecl
}

// Discovery handles test discovery using AST parsing
type Discovery struct {
	dir    string
	logger *Logger
	fset   *token.FileSet
}

// NewDiscovery creates a new test discovery instance
func NewDiscovery(dir string, logger *Logger) *Discovery {
	return &Discovery{
		dir:    dir,
		logger: logger,
		fset:   token.NewFileSet(),
	}
}

// DiscoverTests discovers all test functions in the specified directory
func (d *Discovery) DiscoverTests(ctx context.Context) ([]*TestInfo, error) {
	d.logger.Debug("🔍 Starting test discovery in %s", d.dir)

	var tests []*TestInfo

	err := filepath.Walk(d.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Only process Go test files
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "\\vendor\\") {
			return nil
		}

		d.logger.Debug("📁 Parsing test file: %s", path)

		fileTests, err := d.parseTestFile(path)
		if err != nil {
			d.logger.Warn("Failed to parse %s: %v", path, err)
			return nil // Continue with other files
		}

		tests = append(tests, fileTests...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	d.logger.Debug("🔍 Discovery complete: found %d test(s)", len(tests))
	return tests, nil
}

// parseTestFile parses a single test file and extracts test functions
func (d *Discovery) parseTestFile(filename string) ([]*TestInfo, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Parse the file
	file, err := parser.ParseFile(d.fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var tests []*TestInfo

	// Walk through all declarations
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if it's a test function
		if !d.isTestFunction(funcDecl) {
			continue
		}

		position := d.fset.Position(funcDecl.Pos())

		test := &TestInfo{
			Name:     funcDecl.Name.Name,
			Package:  file.Name.Name,
			File:     filename,
			Line:     position.Line,
			Function: funcDecl,
		}

		tests = append(tests, test)
	}

	return tests, nil
}

// isTestFunction checks if a function declaration is a test function
func (d *Discovery) isTestFunction(fn *ast.FuncDecl) bool {
	name := fn.Name.Name

	// Must start with "Test" or "Benchmark" or "Example"
	if !strings.HasPrefix(name, "Test") &&
		!strings.HasPrefix(name, "Benchmark") &&
		!strings.HasPrefix(name, "Example") {
		return false
	}

	// Must have the correct signature
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}

	param := fn.Type.Params.List[0]
	if len(param.Names) != 1 {
		return false
	}

	// Check parameter type
	starExpr, ok := param.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Should be *testing.T, *testing.B, or similar
	pkgIdent, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return pkgIdent.Name == "testing"
}
