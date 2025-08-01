package discovery

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// TestFunction represents a discovered test function
type TestFunction struct {
	Name        string            `json:"name"`
	Package     string            `json:"package"`
	File        string            `json:"file"`
	Line        int               `json:"line"`
	Tags        []string          `json:"tags"`
	IsSubtest   bool              `json:"is_subtest"`
	ParentTest  string            `json:"parent_test,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

// TestSuite represents a collection of tests in a package
type TestSuite struct {
	Package   string          `json:"package"`
	Path      string          `json:"path"`
	Tests     []*TestFunction `json:"tests"`
	Benchmarks []*TestFunction `json:"benchmarks"`
}

// Discoverer finds and parses Go test files
type Discoverer struct {
	projectRoot string
	fileSet     *token.FileSet
}

// NewDiscoverer creates a new test discoverer
func NewDiscoverer(projectRoot string) *Discoverer {
	return &Discoverer{
		projectRoot: projectRoot,
		fileSet:     token.NewFileSet(),
	}
}

// DiscoverTests finds all test functions in the project
func (d *Discoverer) DiscoverTests() ([]*TestSuite, error) {
	var suites []*TestSuite
	
	err := filepath.Walk(d.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		
		suite, err := d.parseTestFile(path)
		if err != nil {
			return fmt.Errorf("parsing test file %s: %w", path, err)
		}
		
		if suite != nil && (len(suite.Tests) > 0 || len(suite.Benchmarks) > 0) {
			suites = append(suites, suite)
		}
		
		return nil
	})
	
	return suites, err
}

// parseTestFile parses a single test file and extracts test functions
func (d *Discoverer) parseTestFile(filePath string) (*TestSuite, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	file, err := parser.ParseFile(d.fileSet, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	
	suite := &TestSuite{
		Package: file.Name.Name,
		Path:    filePath,
		Tests:   []*TestFunction{},
		Benchmarks: []*TestFunction{},
	}
	
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if isTestFunction(fn.Name.Name) {
				testFunc := d.extractTestFunction(fn, filePath)
				suite.Tests = append(suite.Tests, testFunc)
			} else if isBenchmarkFunction(fn.Name.Name) {
				benchFunc := d.extractTestFunction(fn, filePath)
				suite.Benchmarks = append(suite.Benchmarks, benchFunc)
			}
		}
		return true
	})
	
	return suite, nil
}

// extractTestFunction creates a TestFunction from an AST function declaration
func (d *Discoverer) extractTestFunction(fn *ast.FuncDecl, filePath string) *TestFunction {
	pos := d.fileSet.Position(fn.Pos())
	
	testFunc := &TestFunction{
		Name:     fn.Name.Name,
		Package:  extractPackageName(filePath),
		File:     filePath,
		Line:     pos.Line,
		Tags:     extractTags(fn.Doc),
		Metadata: make(map[string]string),
	}
	
	// Check if this is a subtest by looking for t.Run calls
	if hasSubtests(fn) {
		testFunc.Metadata["has_subtests"] = "true"
	}
	
	return testFunc
}

// isTestFunction checks if a function name indicates a test function
func isTestFunction(name string) bool {
	return strings.HasPrefix(name, "Test") && len(name) > 4
}

// isBenchmarkFunction checks if a function name indicates a benchmark function
func isBenchmarkFunction(name string) bool {
	return strings.HasPrefix(name, "Benchmark") && len(name) > 9
}

// extractTags parses comment tags from function documentation
func extractTags(doc *ast.CommentGroup) []string {
	var tags []string
	if doc == nil {
		return tags
	}
	
	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)
		if strings.HasPrefix(text, "// @") {
			tag := strings.TrimPrefix(text, "// @")
			tags = append(tags, strings.TrimSpace(tag))
		}
	}
	
	return tags
}

// extractPackageName extracts package name from file path
func extractPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

// hasSubtests checks if a test function contains t.Run calls
func hasSubtests(fn *ast.FuncDecl) bool {
	hasRun := false
	ast.Inspect(fn, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Run" {
					hasRun = true
					return false
				}
			}
		}
		return true
	})
	return hasRun
}
