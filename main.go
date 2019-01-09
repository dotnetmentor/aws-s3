package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	// Initialize cli
	c := &cli.CLI{
		Name:    "aws-s3",
		Version: fmt.Sprintf("%s-%s", version, commit),
		Commands: map[string]cli.CommandFactory{
			"prune": pruneCommandFactory,
		},
		Args: args,
	}

	// Run cli
	exitCode, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}

	return exitCode
}
