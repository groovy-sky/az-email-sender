package main

import (
	"os"

	"github.com/groovy-sky/azemailsender/internal/cli/app"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	application := app.NewApp("azemailsender-cli", version, commit, date)
	exitCode := application.Run(os.Args)
	os.Exit(exitCode)
}