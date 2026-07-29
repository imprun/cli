package main

import (
	"os"

	"github.com/imprun/cli/internal/impruncli"
)

func main() {
	os.Exit(impruncli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
