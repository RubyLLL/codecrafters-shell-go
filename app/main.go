package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

const (
	EXIT string = "exit"
	ECHO string = "echo"
	TYPE string = "type"
	PWD  string = "pwd"
	CD   string = "cd"
)

var supportedCommand = []string{EXIT, ECHO, TYPE, PWD, CD}

var paths = strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))

func fetchAllExecutables() ([]string, error) {
	executables := make(map[string]struct{})
	for _, path := range paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			continue // skip if cannot read
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			info, err := entry.Info()
			if err != nil {
				continue
			}
			mode := info.Mode()
			// Check if executable by owner (unix)
			if mode&0111 != 0 {
				executables[name] = struct{}{}
			}
		}
	}

	var result []string
	for exe := range executables {
		result = append(result, exe)
	}

	return result, nil
}

type BellWrapper struct {
	Inner    readline.AutoCompleter
	tabPress bool
	rl       *readline.Instance
}

func (w *BellWrapper) Do(line []rune, pos int) ([][]rune, int) {
	if w.Inner == nil {
		fmt.Fprint(os.Stdout, "\x07")
		return nil, 0
	}

	matches, offset := w.Inner.Do(line, pos)

	// remove duplicates
	seen := make(map[string]struct{})
	uniqueMatches := make([][]rune, 0, len(matches))
	for _, match := range matches {
		s := string(match)
		if _, found := seen[s]; !found {
			seen[s] = struct{}{}
			uniqueMatches = append(uniqueMatches, match)
		}
	}
	matches = uniqueMatches

	sort.Slice(matches, func(i, j int) bool {
		return string(matches[i]) < string(matches[j])
	})

	if len(matches) == 0 {
		fmt.Fprint(os.Stdout, "\x07")
		return nil, 0
	} else if len(matches) == 1 {
		return matches, offset
	} else {
		lcp, length := longestCommonPrefix(matches)

		// offset: the start of the current partial word
		// pos: the position of the cursor
		// together they form the current partial word
		currentPartialWord := line[offset:pos]

		// if the current partial word is already the LCP and tab was pressed again,
		// or if there's no LCP to apply
		if (length > 0 && string(currentPartialWord) == lcp && w.tabPress) || (length == 0 && w.tabPress) {
			// list all options
			w.tabPress = false // reset for future partial completions
			strs := make([]string, 0, len(matches))
			currentPrefix := string(line[:offset])
			for _, s := range matches {
				strs = append(strs, currentPrefix+strings.TrimSpace(string(s)))
			}
			fmt.Fprintf(os.Stdout, "\n%s\n", strings.Join(strs, "  "))
			w.rl.Refresh()
			return nil, 0
		} else if length > 0 { // first tab press with multiple matches and LCP
			// LCP for partial completion
			w.tabPress = true
			// readline needs to know what to complete to. It takes the first element of [][]rune
			// and replaces line[offset:pos] with it.
			// So, we need to return the LCP as a single match.
			return [][]rune{[]rune(lcp)}, offset
		} else { // Multiple matches, no LCP, first tab
			w.tabPress = true
			fmt.Fprint(os.Stdout, "\x07")
			return nil, 0
		}
	}
}

func (w *BellWrapper) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	if key != '\t' {
		w.tabPress = false
	}
	return nil, 0, false
}

func longestCommonPrefix(items [][]rune) (string, int) {
	if len(items) == 0 {
		return "", 0
	}

	// If only one element, the entire thing is the prefix.
	if len(items) == 1 {
		return string(items[0]), len(items[0])
	}

	// Find the minimum length among all items.
	minLen := len(items[0])
	for _, it := range items[1:] {
		if len(it) < minLen {
			minLen = len(it)
		}
	}

	if minLen == 0 {
		return string([]rune{}), 0
	}

	// Compare character by character.
	prefixLen := 0
	for i := 0; i < minLen; i++ {
		ch := items[0][i]
		for _, it := range items[1:] {
			if it[i] != ch {
				// mismatch found
				return string(items[0][:prefixLen]), prefixLen
			}
		}
		prefixLen++
	}

	// All characters up to minLen matched
	return string(items[0][:prefixLen]), prefixLen
}

func main() {
	executableFiles, err := fetchAllExecutables()
	check(err, "Failed to fetch executable files")

	allCommands := append(supportedCommand, executableFiles...)
	items := make([]readline.PrefixCompleterInterface, 0, len(allCommands))
	for _, cmd := range allCommands {
		items = append(items, readline.PcItem(cmd))
	}
	base := readline.NewPrefixCompleter(items...)

	// Wrap it with our bell behavior
	completer := &BellWrapper{
		Inner:    base,
		tabPress: false,
	}

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

	for {
		line, err := rl.Readline()
		if err != nil { // EOF or Ctrl+D
			break
		}

		input := strings.TrimSpace(line)
		var output string
		if strings.Contains(input, "|") {
			pipe(input)
		} else {
			output = runCommand(input)
		}

		switch output {
		case EXIT:
			return
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
		check(err, "Failed to change directory")
		return ""

	default:
		if fullpath, err := executable(parts[0], paths); err == nil && fullpath != "" {
			output := executeScript(command, parts[1:]...)
			return fmt.Sprintf("%s", output)
		}

		return fmt.Sprintf("%s: command not found", command)
	}
}

func check(err error, msg string) {
	if err != nil {
		fmt.Sprintf("%s: %s", msg, err.Error())
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

func pipe(input string) {

	parts := strings.Split(input, "|")
	commands := make([]*exec.Cmd, 0, len(parts))

	for _, cmd := range parts {
		fields := strings.Fields(strings.TrimSpace(cmd))
		command := exec.Command(fields[0], fields[1:]...)
		commands = append(commands, command)
	}

	pipes := make([]*io.PipeWriter, 0, len(commands)-1)
	for i := 0; i < len(commands)-1; i++ {
		pr, pw := io.Pipe()

		commands[i].Stdout = pw
		commands[i+1].Stdin = pr
		pipes = append(pipes, pw)
	}

	commands[len(commands)-1].Stdout = os.Stdout

	for _, command := range commands {
		if err := command.Start(); err != nil {
			fmt.Printf("error running command: %#v", err)
		}
	}

	for i := 0; i < len(commands)-1; i++ {
		pw := pipes[i]
		command := commands[i]

		go func(pw *io.PipeWriter, command *exec.Cmd) {
			command.Wait() // wait for upstream
			pw.Close()     // signal downstram
		}(pw, command)
	}

	// wait the last command to finish
	if err := commands[len(commands)-1].Wait(); err != nil {
		fmt.Printf("error closing the last command: %#v", err)
	}

}

func executeScript(command string, args ...string) string {
	var redirectType string
	var outputFile string

	if len(args) >= 2 {
		redirectType = args[len(args)-2]
		if strings.Contains(redirectType, ">") {
			outputFile = args[len(args)-1]
			args = args[:len(args)-2]
		}
	}

	cmd := exec.Command(command, args...)

	if outputFile != "" {
		flags := os.O_WRONLY | os.O_CREATE
		if strings.Contains(redirectType, ">>") {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		file, err := os.OpenFile(outputFile, flags, 0644)
		check(err, "redirect error")
		defer file.Close()

		switch redirectType {
		case ">", "1>", ">>", "1>>":
			cmd.Stdout = file
			cmd.Stderr = os.Stderr
		case "2>", "2>>":
			cmd.Stdout = os.Stdout
			cmd.Stderr = file
		}

		_ = cmd.Run()
		return ""
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return stderr.String()
		}
		return err.Error()
	}
	return strings.TrimSuffix(string(out), "\n")
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
				case '\\', '"':
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
