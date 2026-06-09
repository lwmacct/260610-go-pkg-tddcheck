package tddcheck_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lwmacct/260610-tddcheck/pkg/tddcheck"
)

func ExampleCheck() {
	root, err := os.MkdirTemp("", "tddcheck-example-*")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() { _ = os.RemoveAll(root) }()

	source := `package calc

func Parse() {}
`
	test := `package calc

import "testing"

func TestParse(t *testing.T) {
	got := 1
	if got != 1 {
		t.Fatal(got)
	}
}
`
	if err := os.WriteFile(filepath.Join(root, "calc.go"), []byte(source), 0600); err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile(filepath.Join(root, "calc_test.go"), []byte(test), 0600); err != nil {
		fmt.Println(err)
		return
	}

	result := tddcheck.Check(tddcheck.WithRoot(root))
	if result.Err != nil {
		fmt.Println(result.Err)
		return
	}

	fmt.Println(result.Passed)

	// Output:
	// true
}
