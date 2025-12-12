package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Executor handles command execution including external programs,
// pipes, and redirections
type Executor struct {
	pathFinder *PathFinder
	builtins   *BuiltinCommands
}

// NewExecutor creates a new Executor instance
func NewExecutor(pf *PathFinder, bc *BuiltinCommands) *Executor {
	return &Executor{
		pathFinder: pf,
		builtins:   bc,
	}
}

// Execute runs a command (builtin or external)
func (e *Executor) Execute(input string) (string, error) {
	input = strings.TrimSpace(input)

	// Handle pipes
	if strings.Contains(input, "|") {
		return "", e.executePipe(input)
	}

	parts := ParseArgs(input)
	if len(parts) == 0 {
		return "", nil
	}

	command := parts[0]
	args := parts[1:]

	// Check if it's a builtin command (and not redirected)
	if e.builtins.IsBuiltin(command) && !strings.Contains(input, ">") {
		var buf bytes.Buffer
		if err := e.builtins.Execute(command, args, os.Stdin, &buf); err != nil {
			return "", err
		}
		return strings.TrimSuffix(buf.String(), "\n"), nil
	}

	// Execute external command
	return e.executeExternal(command, args)
}

// executeExternal runs an external program with optional redirection
func (e *Executor) executeExternal(command string, args []string) (string, error) {
	fullPath := e.pathFinder.FindExecutable(command)
	if fullPath == "" {
		return "", fmt.Errorf("%s: command not found", command)
	}

	// Check for output redirection
	redirectType, outputFile, actualArgs := parseRedirection(args)

	// Use command name (not full path) as argv[0] to match shell behavior
	cmd := exec.Command(command, actualArgs...)
	cmd.Path = fullPath

	if outputFile != "" {
		flags := os.O_WRONLY | os.O_CREATE
		if strings.Contains(redirectType, ">>") {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		file, err := os.OpenFile(outputFile, flags, 0644)
		if err != nil {
			return "", fmt.Errorf("redirect error: %v", err)
		}
		defer file.Close()

		switch redirectType {
		case ">", "1>", ">>", "1>>":
			cmd.Stdout = file
			cmd.Stderr = os.Stderr
		case "2>", "2>>":
			cmd.Stdout = os.Stdout
			cmd.Stderr = file
		}

		// Silently run - errors are not returned for redirected commands
		_ = cmd.Run()
		return "", nil
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("%s", stderr.String())
		}
		return "", err
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}

// parseRedirection extracts redirection operators and file from arguments
func parseRedirection(args []string) (redirectType, outputFile string, actualArgs []string) {
	if len(args) < 2 {
		return "", "", args
	}

	redirectType = args[len(args)-2]
	if strings.Contains(redirectType, ">") {
		outputFile = args[len(args)-1]
		actualArgs = args[:len(args)-2]
		return
	}

	return "", "", args
}

// executePipe handles piped commands
func (e *Executor) executePipe(input string) error {
	var commands [][]string
	pipeCommands := strings.Split(input, "|")

	for _, pc := range pipeCommands {
		cmdParts := ParseArgs(strings.TrimSpace(pc))
		commands = append(commands, cmdParts)
	}

	if len(commands) < 2 {
		return nil
	}

	var cmds []*exec.Cmd
	var pipes []*os.File

	// Create pipes
	for i := 0; i < len(commands)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		pipes = append(pipes, r, w)
	}

	// Set up each command
	for i, cmdParts := range commands {
		if len(cmdParts) == 0 {
			continue
		}

		cmdName := cmdParts[0]
		cmdArgs := cmdParts[1:]

		if e.builtins.IsBuiltin(cmdName) {
			// Handle builtin command in pipe
			var stdin, stdout *os.File

			if i == 0 {
				stdin = os.Stdin
			} else {
				stdin = pipes[(i-1)*2]
			}

			if i == len(commands)-1 {
				stdout = os.Stdout
			} else {
				stdout = pipes[i*2+1]
			}

			go func(name string, args []string, in, out *os.File, isLast bool) {
				defer func() {
					if !isLast {
						out.Close()
					}
				}()
				e.builtins.Execute(name, args, in, out)
			}(cmdName, cmdArgs, stdin, stdout, i == len(commands)-1)
		} else {
			// Handle external command
			fullPath := e.pathFinder.FindExecutable(cmdName)
			if fullPath == "" {
				continue
			}

			cmd := exec.Command(fullPath, cmdArgs...)

			if i == 0 {
				cmd.Stdin = os.Stdin
			} else {
				cmd.Stdin = pipes[(i-1)*2]
			}

			if i == len(commands)-1 {
				cmd.Stdout = os.Stdout
			} else {
				cmd.Stdout = pipes[i*2+1]
			}

			cmd.Stderr = os.Stderr
			cmds = append(cmds, cmd)
		}
	}

	// Start all external commands
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	// Close all pipe write ends in parent
	for i := 1; i < len(pipes); i += 2 {
		pipes[i].Close()
	}

	// Wait for all external commands
	for _, cmd := range cmds {
		cmd.Wait()
	}

	// Close all pipe read ends
	for i := 0; i < len(pipes); i += 2 {
		pipes[i].Close()
	}

	return nil
}
