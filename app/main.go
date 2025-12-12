package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var history = &History{File: os.Getenv("HISTFILE"), MaxLen: math.MaxInt64}

func main() {
	history.ReadFromFile()

	// Initialize core components
	pathFinder := NewPathFinder()
	builtins := NewBuiltinCommands(pathFinder, history)
	executor := NewExecutor(pathFinder, builtins)

	// Setup tab completion
	completer, err := SetupCompleter(builtins, pathFinder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup completer: %v\n", err)
		os.Exit(1)
	}

	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: completer,
		Listener:     completer,
	})

	completer.rl = rl

	if err != nil {
		panic(err)
	}
	defer rl.Close()

	// Main REPL loop
	for {
		line, err := rl.Readline()
		if err != nil { // EOF or Ctrl+D
			break
		}

		history.Write(line)

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		output, err := executor.Execute(input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else if len(output) > 0 {
			fmt.Println(output)
		}
	}
}
