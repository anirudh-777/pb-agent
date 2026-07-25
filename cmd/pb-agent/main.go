package main

import (
	"os"

	"github.com/anirudh-777/pb-agent/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
