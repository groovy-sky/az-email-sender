package app

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// App represents the main CLI application
type App struct {
	name    string
	version string
	commit  string
	date    string
}

// GlobalFlags holds global command-line flags
type GlobalFlags struct {
	Config string
	Debug  bool
	Quiet  bool
	JSON   bool
}

// NewApp creates a new CLI application
func NewApp(name, version, commit, date string) *App {
	return &App{
		name:    name,
		version: version,
		commit:  commit,
		date:    date,
	}
}

// Run executes the CLI application
func (a *App) Run(args []string) int {
	// Handle help as first argument
	if len(args) >= 2 && (args[1] == "help" || args[1] == "--help" || args[1] == "-h") {
		if len(args) > 2 {
			// Help for specific command
			globalFlags := &GlobalFlags{}
			return a.runCommand(args[2], []string{args[2], "--help"}, globalFlags)
		}
		a.printUsage()
		return 0
	}

	if len(args) < 2 {
		a.printUsage()
		return 1
	}

	// Parse global flags and find command
	globalFlags := &GlobalFlags{}
	var command string
	var commandIdx int
	var remainingArgs []string

	// Look for the command in the arguments
	for i := 1; i < len(args); i++ {
		if args[i] == "send" || args[i] == "status" || args[i] == "config" || args[i] == "version" {
			command = args[i]
			commandIdx = i
			break
		}
	}

	if command == "" {
		fmt.Fprintf(os.Stderr, "Error: no command specified\n")
		a.printUsage()
		return 1
	}

	// Parse global flags before the command
	if commandIdx > 1 {
		globalFlagSet := flag.NewFlagSet(a.name, flag.ContinueOnError)
		globalFlagSet.Usage = func() {
			a.printUsage()
		}

		globalFlagSet.StringVar(&globalFlags.Config, "config", "", "Configuration file path")
		globalFlagSet.StringVar(&globalFlags.Config, "c", "", "Configuration file path")
		globalFlagSet.BoolVar(&globalFlags.Debug, "debug", false, "Enable debug logging")
		globalFlagSet.BoolVar(&globalFlags.Debug, "d", false, "Enable debug logging")
		globalFlagSet.BoolVar(&globalFlags.Quiet, "quiet", false, "Suppress output except errors")
		globalFlagSet.BoolVar(&globalFlags.Quiet, "q", false, "Suppress output except errors")
		globalFlagSet.BoolVar(&globalFlags.JSON, "json", false, "Output in JSON format")
		globalFlagSet.BoolVar(&globalFlags.JSON, "j", false, "Output in JSON format")

		globalArgs := args[1:commandIdx]
		if err := globalFlagSet.Parse(globalArgs); err != nil {
			return 1
		}
	}

	// Command and its arguments
	remainingArgs = args[commandIdx:]

	return a.runCommand(command, remainingArgs, globalFlags)
}

// runCommand executes a specific command
func (a *App) runCommand(command string, args []string, globalFlags *GlobalFlags) int {
	switch command {
	case "send":
		return a.runSend(args, globalFlags)
	case "status":
		return a.runStatus(args, globalFlags)
	case "config":
		return a.runConfig(args, globalFlags)
	case "version":
		return a.runVersion(args, globalFlags)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		a.printUsage()
		return 1
	}
}

// printUsage prints the main usage information
func (a *App) printUsage() {
	fmt.Printf(`A command-line interface for sending emails using Azure Communication Services.
Supports multiple authentication methods, flexible recipient management,
and both plain text and HTML email content.

Usage:
  %s [global flags] <command> [command flags]

Available Commands:
  send        Send an email
  status      Check email status
  config      Manage configuration
  version     Show version information

Global Flags:
  -c, --config string   Configuration file path
  -d, --debug           Enable debug logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress output except errors
  -h, --help            Show this help

Use "%s <command> --help" for more information about a command.
`, a.name, a.name)
}

// stringSliceFlag implements flag.Value for string slices
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// newStringSliceFlag creates a new string slice flag
func newStringSliceFlag() *stringSliceFlag {
	return &stringSliceFlag{}
}