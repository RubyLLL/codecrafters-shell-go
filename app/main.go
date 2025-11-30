package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
var _ = fmt.Fprint
var _ = os.Stdout

var supported_cmd = []string{"exit", "echo", "type"}
var paths = strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))

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
	case "type":
		if len(parts) > 1 {
			if supported_command(parts[1]) {
				return fmt.Sprintf("%s is a shell builtin", parts[1])
			} else if fullpath, err := executable(parts[1], paths); err == nil && fullpath != "" {
				return fmt.Sprintf("%s is %s", parts[1], fullpath)
			} else {
				return fmt.Sprintf("%s: not found", parts[1])
			}
		}
		return ""

	default:
		if fullpath, err := executable(parts[0], paths); err == nil && fullpath != "" {
			output := executeScript(command, parts[1:]...)
			return fmt.Sprintf("%s", output)
		}

		return fmt.Sprintf("%s: command not found", command)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func supported_command(command string) bool {
	for _, s := range supported_cmd {
		if s == command {
			return true
		}
	}
	return false
}

func executable(command string, paths []string) (string, error) {
	for _, p := range paths {
		fp := filepath.Join(p, command)
		if info, err := os.Stat(fp); err == nil && info.Mode().IsRegular() && (info.Mode()&0111 != 0) {
			return fp, nil
		}
	}

	return "", nil
}

func executeScript(command string, args ...string) string {
	cmd, err := exec.Command(command, args...).Output()
	check(err)
	output := string(cmd)
	return strings.TrimSuffix(output, "\n")
}
