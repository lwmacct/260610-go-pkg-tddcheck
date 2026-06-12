package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		root        string
		stagedOnly  bool
		showVersion bool
	)

	flags := flag.NewFlagSet("tddcheck", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&root, "root", ".", "project root")
	flags.BoolVar(&stagedOnly, "staged", false, "check staged git changes")
	flags.BoolVar(&showVersion, "version", false, "print version")

	if err := flags.Parse(os.Args[1:]); err != nil {
		return 2
	}
	if showVersion {
		fmt.Println(version)
		return 0
	}

	result := tddcheck.Check(
		tddcheck.WithRoot(root),
		tddcheck.WithChanged(stagedOnly),
	)
	if result.Err != nil {
		fmt.Fprintln(os.Stderr, "tddcheck:", result.Err)
		return 2
	}

	fmt.Println(result.Text())
	if !result.Passed {
		return 1
	}

	return 0
}
