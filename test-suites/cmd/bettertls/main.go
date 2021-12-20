package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	subcommands := map[string]func([]string) error{
		"server":             runServer,
		"get-test":           getTest,
		"run-tests":          runTests,
		"generate-manifests": generateManifests,
		"show-results":       showResults,
	}

	var subcommand func([]string) error
	if len(os.Args) >= 2 {
		subcommand = subcommands[os.Args[1]]
	}
	if subcommand == nil {
		fmt.Printf("Usage: %s <server|get-test> ...\n", os.Args[0])
		c := make([]string, 0, len(subcommands))
		for k := range subcommands {
			c = append(c, k)
		}
		fmt.Printf("Supported sub-commands: %s\n", strings.Join(c, ", "))
		return
	}

	err := subcommand((os.Args[2:]))
	if err != nil && err != flag.ErrHelp {
		panic(err)
	}
}
