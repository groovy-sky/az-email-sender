package simplecli

import (
	"fmt"
	"os"
	"strings"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Usage       string
	LongDesc    string
	Examples    string
	Run         func(*Context) error
	Flags       []*Flag
	Subcommands []*Command
}

// Flag represents a command-line flag
type Flag struct {
	Name        string
	Short       string
	Description string
	Value       interface{}
	Required    bool
	EnvVar      string
}

// Context holds the execution context for a command
type Context struct {
	Args      []string
	Flags     map[string]interface{}
	Command   *Command
	GlobalCtx *GlobalContext
}

// GlobalContext holds global CLI state
type GlobalContext struct {
	AppName     string
	Description string
	Version     string
	Commit      string
	Date        string
	GlobalFlags []*Flag
	Commands    []*Command
}

// NewGlobalContext creates a new global CLI context
func NewGlobalContext(appName, description string) *GlobalContext {
	return &GlobalContext{
		AppName:     appName,
		Description: description,
		Commands:    make([]*Command, 0),
		GlobalFlags: make([]*Flag, 0),
	}
}

// AddGlobalFlag adds a global flag
func (g *GlobalContext) AddGlobalFlag(flag *Flag) {
	g.GlobalFlags = append(g.GlobalFlags, flag)
}

// AddCommand adds a command
func (g *GlobalContext) AddCommand(cmd *Command) {
	g.Commands = append(g.Commands, cmd)
}

// Run executes the CLI
func (g *GlobalContext) Run() error {
	args := os.Args[1:]
	
	if len(args) == 0 {
		g.printHelp()
		return nil
	}

	// Check for global help
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		g.printHelp()
		return nil
	}

	// Parse global flags and find command
	globalFlags, remainingArgs, err := g.parseGlobalFlags(args)
	if err != nil {
		return err
	}

	if len(remainingArgs) == 0 {
		g.printHelp()
		return nil
	}

	// Find command
	cmdName := remainingArgs[0]
	cmd := g.findCommand(cmdName)
	if cmd == nil {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Parse command flags and arguments
	ctx, err := g.parseCommand(cmd, globalFlags, remainingArgs[1:])
	if err != nil {
		return err
	}

	// Run command (use the command from context in case it's a subcommand)
	return ctx.Command.Run(ctx)
}

// parseGlobalFlags parses global flags from arguments
func (g *GlobalContext) parseGlobalFlags(args []string) (map[string]interface{}, []string, error) {
	flags := make(map[string]interface{})
	var remaining []string
	
	// Set defaults for global flags
	for _, flag := range g.GlobalFlags {
		flags[flag.Name] = flag.Value
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		
		if !strings.HasPrefix(arg, "-") {
			remaining = append(remaining, args[i:]...)
			break
		}

		// Find matching global flag
		var matchedFlag *Flag
		var value string
		hasValue := false

		// Handle --flag=value format
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			arg = parts[0]
			value = parts[1]
			hasValue = true
		}

		// Find flag by name or short name
		for _, flag := range g.GlobalFlags {
			if arg == "--"+flag.Name || (flag.Short != "" && arg == "-"+flag.Short) {
				matchedFlag = flag
				break
			}
		}

		if matchedFlag == nil {
			// Check if this looks like a flag we should handle later
			if strings.HasPrefix(arg, "-") {
				// This is likely a command-specific flag, keep it for later
				remaining = append(remaining, args[i:]...)
				break
			} else {
				// Not a flag, add to remaining
				remaining = append(remaining, args[i:]...)
				break
			}
		}

		// Get flag value
		if !hasValue {
			switch matchedFlag.Value.(type) {
			case bool:
				flags[matchedFlag.Name] = true
				i++
				continue
			default:
				if i+1 >= len(args) {
					return nil, nil, fmt.Errorf("flag --%s requires a value", matchedFlag.Name)
				}
				value = args[i+1]
				i += 2
			}
		} else {
			i++
		}

		// Set flag value
		err := g.setFlagValue(flags, matchedFlag, value)
		if err != nil {
			return nil, nil, err
		}
	}

	return flags, remaining, nil
}

// parseCommand parses command-specific flags and arguments
func (g *GlobalContext) parseCommand(cmd *Command, globalFlags map[string]interface{}, args []string) (*Context, error) {
	flags := make(map[string]interface{})
	
	// Copy global flags
	for k, v := range globalFlags {
		flags[k] = v
	}
	
	// Check for subcommands first, before processing any flags
	if len(cmd.Subcommands) > 0 && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subCmdName := args[0]
		var subCmd *Command
		for _, sc := range cmd.Subcommands {
			if sc.Name == subCmdName {
				subCmd = sc
				break
			}
		}
		
		if subCmd != nil {
			// Parse subcommand flags
			subCtx, err := g.parseCommand(subCmd, globalFlags, args[1:])
			if err != nil {
				return nil, err
			}
			
			// Copy global flags to subcommand context
			for k, v := range flags {
				if _, exists := subCtx.Flags[k]; !exists {
					subCtx.Flags[k] = v
				}
			}
			
			return subCtx, nil
		}
	}
	
	// Set defaults for command flags
	for _, flag := range cmd.Flags {
		flags[flag.Name] = flag.Value
		
		// Check environment variable
		if flag.EnvVar != "" {
			if envVal := os.Getenv(flag.EnvVar); envVal != "" {
				err := g.setFlagValue(flags, flag, envVal)
				if err != nil {
					return nil, fmt.Errorf("invalid environment variable %s: %w", flag.EnvVar, err)
				}
			}
		}
	}

	var cmdArgs []string
	i := 0
	
	for i < len(args) {
		arg := args[i]
		
		// If this command has subcommands, check if this arg is a subcommand name
		if len(cmd.Subcommands) > 0 && !strings.HasPrefix(arg, "-") {
			for _, subCmd := range cmd.Subcommands {
				if arg == subCmd.Name {
					// Found subcommand, add everything to cmdArgs
					cmdArgs = append(cmdArgs, args[i:]...)
					goto endFlagParsing
				}
			}
		}
		
		if !strings.HasPrefix(arg, "-") {
			cmdArgs = append(cmdArgs, args[i:]...)
			break
		}

		// Check for help
		if arg == "--help" || arg == "-h" {
			g.printCommandHelp(cmd)
			os.Exit(0)
		}

		// Find matching command flag
		var matchedFlag *Flag
		var value string
		hasValue := false

		// Handle --flag=value format
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			arg = parts[0]
			value = parts[1]
			hasValue = true
		}

		// Find flag by name or short name (check command flags first, then global flags)
		for _, flag := range cmd.Flags {
			if arg == "--"+flag.Name || (flag.Short != "" && arg == "-"+flag.Short) {
				matchedFlag = flag
				break
			}
		}
		
		// If not found in command flags, check global flags
		if matchedFlag == nil {
			for _, flag := range g.GlobalFlags {
				if arg == "--"+flag.Name || (flag.Short != "" && arg == "-"+flag.Short) {
					matchedFlag = flag
					break
				}
			}
		}

		if matchedFlag == nil {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}

		// Get flag value
		if !hasValue {
			switch matchedFlag.Value.(type) {
			case bool:
				flags[matchedFlag.Name] = true
				i++
				continue
			default:
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag --%s requires a value", matchedFlag.Name)
				}
				value = args[i+1]
				i += 2
			}
		} else {
			i++
		}

		// Set flag value
		err := g.setFlagValue(flags, matchedFlag, value)
		if err != nil {
			return nil, err
		}
	}

endFlagParsing:

	// If this command requires subcommands but we don't have valid cmdArgs, show error
	if len(cmd.Subcommands) > 0 {
		if len(cmdArgs) == 0 {
			g.printCommandHelp(cmd)
			return nil, fmt.Errorf("subcommand required")
		}
		
		// If we get here, it means the first arg wasn't a recognized subcommand
		return nil, fmt.Errorf("unknown subcommand: %s", cmdArgs[0])
	}

	// Validate required flags
	for _, flag := range cmd.Flags {
		if flag.Required && flags[flag.Name] == flag.Value {
			return nil, fmt.Errorf("required flag --%s not provided", flag.Name)
		}
	}

	return &Context{
		Args:      cmdArgs,
		Flags:     flags,
		Command:   cmd,
		GlobalCtx: g,
	}, nil
}

// setFlagValue sets a flag value with type conversion
func (g *GlobalContext) setFlagValue(flags map[string]interface{}, flag *Flag, value string) error {
	switch flag.Value.(type) {
	case string:
		flags[flag.Name] = value
	case bool:
		if value == "true" || value == "1" || value == "" {
			flags[flag.Name] = true
		} else if value == "false" || value == "0" {
			flags[flag.Name] = false
		} else {
			return fmt.Errorf("invalid boolean value for --%s: %s", flag.Name, value)
		}
	case []string:
		if existing, ok := flags[flag.Name].([]string); ok {
			flags[flag.Name] = append(existing, value)
		} else {
			flags[flag.Name] = []string{value}
		}
	default:
		flags[flag.Name] = value
	}
	return nil
}

// findCommand finds a command by name
func (g *GlobalContext) findCommand(name string) *Command {
	for _, cmd := range g.Commands {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}

// printHelp prints the main help message
func (g *GlobalContext) printHelp() {
	fmt.Printf("%s\n\n", g.Description)
	fmt.Printf("Usage:\n  %s [command]\n\n", g.AppName)
	
	if len(g.Commands) > 0 {
		fmt.Println("Available Commands:")
		for _, cmd := range g.Commands {
			fmt.Printf("  %-12s %s\n", cmd.Name, cmd.Description)
		}
		fmt.Println()
	}
	
	if len(g.GlobalFlags) > 0 {
		fmt.Println("Flags:")
		for _, flag := range g.GlobalFlags {
			flagStr := "--" + flag.Name
			if flag.Short != "" {
				flagStr = "-" + flag.Short + ", " + flagStr
			}
			fmt.Printf("  %-20s %s\n", flagStr, flag.Description)
		}
		fmt.Println()
	}
	
	fmt.Printf("Use \"%s [command] --help\" for more information about a command.\n", g.AppName)
}

// printCommandHelp prints help for a specific command
func (g *GlobalContext) printCommandHelp(cmd *Command) {
	fmt.Printf("%s\n\n", cmd.LongDesc)
	fmt.Printf("Usage:\n  %s %s\n\n", g.AppName, cmd.Usage)
	
	if cmd.Examples != "" {
		fmt.Printf("Examples:\n%s\n\n", cmd.Examples)
	}
	
	if len(cmd.Subcommands) > 0 {
		fmt.Println("Available Commands:")
		for _, subCmd := range cmd.Subcommands {
			fmt.Printf("  %-12s %s\n", subCmd.Name, subCmd.Description)
		}
		fmt.Println()
	}
	
	if len(cmd.Flags) > 0 {
		fmt.Println("Flags:")
		for _, flag := range cmd.Flags {
			flagStr := "--" + flag.Name
			if flag.Short != "" {
				flagStr = "-" + flag.Short + ", " + flagStr
			}
			fmt.Printf("  %-20s %s\n", flagStr, flag.Description)
		}
		fmt.Println()
	}
	
	if len(g.GlobalFlags) > 0 {
		fmt.Println("Global Flags:")
		for _, flag := range g.GlobalFlags {
			flagStr := "--" + flag.Name
			if flag.Short != "" {
				flagStr = "-" + flag.Short + ", " + flagStr
			}
			fmt.Printf("  %-20s %s\n", flagStr, flag.Description)
		}
	}
}

// Helper methods for Context

// GetString gets a string flag value
func (c *Context) GetString(name string) string {
	if val, ok := c.Flags[name].(string); ok {
		return val
	}
	return ""
}

// GetBool gets a boolean flag value
func (c *Context) GetBool(name string) bool {
	if val, ok := c.Flags[name].(bool); ok {
		return val
	}
	return false
}

// GetStringSlice gets a string slice flag value
func (c *Context) GetStringSlice(name string) []string {
	if val, ok := c.Flags[name].([]string); ok {
		return val
	}
	return []string{}
}