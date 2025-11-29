package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
var _ = fmt.Fprint
var _ = os.Stdout

func main() {
	// TODO: Uncomment the code below to pass the first stage
L:
	for true {
		fmt.Fprint(os.Stdout, "$ ")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')

		check(err)

		input = strings.TrimSpace(input)
		output := runCommand(input)

		switch output {
		case "exit":
			break L
		default:
			fmt.Println(output)
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func runCommand(input string) string {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, " ")
	if len(parts) == 0 {
		return ""
	}

	command := parts[0]

	switch command {
	case "exit":
		return "exit"
	case "echo":
		if len(parts) > 1 {
			return strings.Join(parts[1:], " ")
		}
		return ""
	default:
		return fmt.Sprintf("%s: command not found", command)
	}
}
