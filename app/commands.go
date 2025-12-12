package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Command represents a shell command that can be executed
type Command interface {
	Name() string
	Execute(args []string, stdin io.Reader, stdout io.Writer) error
}

// BuiltinCommands holds all builtin commands
type BuiltinCommands struct {
	commands   map[string]Command
	pathFinder *PathFinder
	history    *History
}

// NewBuiltinCommands creates a new BuiltinCommands instance
func NewBuiltinCommands(pf *PathFinder, hist *History) *BuiltinCommands {
	bc := &BuiltinCommands{
		commands:   make(map[string]Command),
		pathFinder: pf,
		history:    hist,
	}

	// Register all builtin commands
	bc.register(&EchoCommand{})
	bc.register(&TypeCommand{pathFinder: pf, builtins: bc})
	bc.register(&PwdCommand{})
	bc.register(&CdCommand{})
	bc.register(&ExitCommand{history: hist})
	bc.register(&HistoryCommand{history: hist})

	return bc
}

func (bc *BuiltinCommands) register(cmd Command) {
	bc.commands[cmd.Name()] = cmd
}

// IsBuiltin checks if a command is a builtin
func (bc *BuiltinCommands) IsBuiltin(name string) bool {
	_, exists := bc.commands[name]
	return exists
}

// Execute executes a builtin command
func (bc *BuiltinCommands) Execute(name string, args []string, stdin io.Reader, stdout io.Writer) error {
	cmd, exists := bc.commands[name]
	if !exists {
		return fmt.Errorf("%s: command not found", name)
	}
	return cmd.Execute(args, stdin, stdout)
}

// GetCommandNames returns all builtin command names
func (bc *BuiltinCommands) GetCommandNames() []string {
	names := make([]string, 0, len(bc.commands))
	for name := range bc.commands {
		names = append(names, name)
	}
	return names
}

// EchoCommand implements the echo builtin
type EchoCommand struct{}

func (c *EchoCommand) Name() string { return "echo" }

func (c *EchoCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	output := ""
	if len(args) > 0 {
		output = strings.Join(args, " ")
	}
	fmt.Fprintln(stdout, output)
	return nil
}

// TypeCommand implements the type builtin
type TypeCommand struct {
	pathFinder *PathFinder
	builtins   *BuiltinCommands
}

func (c *TypeCommand) Name() string { return "type" }

func (c *TypeCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return nil
	}

	arg := args[0]
	if c.builtins.IsBuiltin(arg) {
		fmt.Fprintf(stdout, "%s is a shell builtin\n", arg)
		return nil
	}

	if fullPath := c.pathFinder.FindExecutable(arg); fullPath != "" {
		fmt.Fprintf(stdout, "%s is %s\n", arg, fullPath)
		return nil
	}

	fmt.Fprintf(stdout, "%s: not found\n", arg)
	return nil
}

// PwdCommand implements the pwd builtin
type PwdCommand struct{}

func (c *PwdCommand) Name() string { return "pwd" }

func (c *PwdCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %v", err)
	}
	fmt.Fprintln(stdout, cwd)
	return nil
}

// CdCommand implements the cd builtin
type CdCommand struct{}

func (c *CdCommand) Name() string { return "cd" }

func (c *CdCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	targetDir := ""

	if len(args) == 0 || args[0] == "~" {
		// go to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cd: %v", err)
		}
		targetDir = homeDir
	} else {
		targetDir = args[0]
	}

	if err := os.Chdir(targetDir); err != nil {
		return fmt.Errorf("cd: %s: No such file or directory", targetDir)
	}

	return nil
}

// ExitCommand implements the exit builtin
type ExitCommand struct {
	history *History
}

func (c *ExitCommand) Name() string { return "exit" }

func (c *ExitCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	exitCode := 0

	if len(args) > 0 {
		code, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(stdout, "exit: %s: numeric argument required\n", args[0])
			exitCode = 2 // Common shell exit code for invalid argument
		} else {
			exitCode = code
		}
	}

	// write history to history file
	c.history.WriteToFile()

	os.Exit(exitCode)
	return nil
}

// HistoryCommand implements the history builtin
type HistoryCommand struct {
	history *History
}

func (c *HistoryCommand) Name() string { return "history" }

func (c *HistoryCommand) Execute(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 1 {
		// history <num>
		if cnt, err := strconv.Atoi(args[0]); err == nil {
			c.history.MaxLen = cnt
		}
	} else if len(args) == 2 {
		switch args[0] {
		case "-r":
			// history -r <history_file>
			c.history.File = args[1]
			if err := c.history.ReadFromFile(); err != nil {
				return fmt.Errorf("history: %v", err)
			}
			return nil
		case "-w":
			// history -w <history_file>
			c.history.File = args[1]
			if err := c.history.WriteToFile(); err != nil {
				return fmt.Errorf("history: %v", err)
			}
			return nil
		case "-a":
			// history -a <history_file>
			c.history.File = args[1]
			if err := c.history.AppendToFile(); err != nil {
				return fmt.Errorf("history: %v", err)
			}
			return nil
		}
	}

	c.history.Get()
	return nil
}
