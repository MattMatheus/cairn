package main

import (
	"context"
	"os"

	"cairn/internal/cli"
)

func main() {
	code := cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	os.Exit(code)
}
