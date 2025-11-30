package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" and "os" imports in stage 1 (feel free to remove this!)
var _ = fmt.Fprint
var _ = os.Stdout

const (
	EXIT string = "exit"
	ECHO string = "echo"
	TYPE string = "type"
	PWD  string = "pwd"
	CD   string = "cd"
)

var supportedCommand = []string{EXIT, ECHO, TYPE, PWD, CD}

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
		case EXIT:
			break L
		default:
			if len(output) > 0 {
				fmt.Println(output)
			}
		}
	}
}

func runCommand(input string) string {
	input = strings.TrimSpace(input)
	parts := parseArgs(input)
	if len(parts) == 0 {
		return ""
	}

	command := parts[0]

	switch command {
	case EXIT:
		return "exit"

	case ECHO:
		if len(parts) > 1 {
			return strings.Join(parts[1:], " ")
		}
		return ""

	case TYPE:
		if len(parts) > 1 {
			if supported(parts[1]) {
				return fmt.Sprintf("%s is a shell builtin", parts[1])
			} else if fullpath, err := executable(parts[1], paths); err == nil && fullpath != "" {
				return fmt.Sprintf("%s is %s", parts[1], fullpath)
			} else {
				return fmt.Sprintf("%s: not found", parts[1])
			}
		}
		return ""

	case PWD:
		output, _ := os.Getwd()
		return fmt.Sprintf("%s", output)

	case CD:
		var target = parts[1]
		if parts[1] == "~" {
			target, _ = os.UserHomeDir()
		} else if !exist(target) {
			return fmt.Sprintf("cd: %s: No such file or directory", parts[1])
		}
		err := os.Chdir(target)
		check(err)
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

func supported(command string) bool {
	return slices.Contains(supportedCommand, command)
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

func exist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func executeScript(command string, args ...string) string {
	cmd, err := exec.Command(command, args...).Output()
	check(err)
	output := string(cmd)
	return strings.TrimSuffix(output, "\n")
}

func parseArgs(input string) []string {
	var args []string
	var curr strings.Builder

	inSingle := false
	inDouble := false

	escaped := false

	for i := 0; i < len(input); i++ {
		c := input[i]

		if escaped {
			if inDouble {
				switch c {
				case '\\':
					curr.WriteByte(c)
				default:
					curr.WriteByte('\\')
					curr.WriteByte(c)
				}
			} else {
				curr.WriteByte(c)
			}

			escaped = false
			continue
		}

		if c == '\\' && !inSingle {
			escaped = true
			continue
		}

		if c == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}

		if c == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if (c == ' ' || c == '\t') && !inSingle && !inDouble {
			if curr.Len() > 0 {
				args = append(args, curr.String())
				curr.Reset()
			}
			continue
		}

		curr.WriteByte(c)
	}

	if curr.Len() > 0 {
		args = append(args, curr.String())
	}

	return args
}
