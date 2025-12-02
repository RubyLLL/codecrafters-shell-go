package main

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/chzyer/readline"
)

// Mock AutoCompleter for testing BellWrapper
type mockCompleter struct {
	matches [][]rune
	offset  int
}

func (m *mockCompleter) Do(line []rune, pos int) ([][]rune, int) {
	return m.matches, m.offset
}

// Mock Readline Instance for testing BellWrapper
type mockReadlineInstance struct {
	io.Writer
}

func (m *mockReadlineInstance) Refresh() error {
	return nil
}

func (m *mockReadlineInstance) Write(p []byte) (n int, err error) {
	return m.Writer.Write(p)
}

func TestBellWrapperPartialCompletion(t *testing.T) {
	tests := []struct {
		name              string
		initialLine       string
		initialPos        int
		mockCompleter     *mockCompleter
		expectedOutput    string // What should be written to stdout
		expectedLine      string // The line returned by Do
		expectedPos       int    // The pos returned by Do
		tabPressCount     int
		expectBell        bool
		expectListOptions bool
	}{
		{
			name:        "First tab press, multiple matches with LCP",
			initialLine: "ec",
			initialPos:  2,
			mockCompleter: &mockCompleter{
				matches: [][]rune{[]rune("echo"), []rune("exit")},
				offset:  0,
			},
			tabPressCount: 1,
			expectedLine:  "e",
			expectedPos:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdoutBuf bytes.Buffer
			// Temporarily redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			completer := &BellWrapper{
				Inner:    tc.mockCompleter,
				tabPress: false,
			}
			mockRl := &mockReadlineInstance{Writer: &stdoutBuf}
			completer.rl = mockRl

			// Simulate tab presses
			var gotMatches [][]rune
			var gotOffset int
			for i := 0; i < tc.tabPressCount; i++ {
				gotMatches, gotOffset = completer.Do([]rune(tc.initialLine), tc.initialPos)
			}

			w.Close()
			os.Stdout = oldStdout
			io.Copy(&stdoutBuf, r)

			if tc.expectBell && !strings.Contains(stdoutBuf.String(), "\x07") {
				t.Errorf("Expected bell sound, but not found in output: %q", stdoutBuf.String())
			}
			if tc.expectListOptions && !strings.Contains(stdoutBuf.String(), "echo  exit") { // Example check for listing
				t.Errorf("Expected options to be listed, but not found in output: %q", stdoutBuf.String())
			}

			if gotOffset != tc.expectedPos {
				t.Errorf("Do() got offset %d, want %d", gotOffset, tc.expectedPos)
			}
			// Only compare the first element of matches, as that's what readline uses for completion in this mode
			if len(gotMatches) > 0 && string(gotMatches[0]) != tc.expectedLine {
				t.Errorf("Do() got line %q, want %q", string(gotMatches[0]), tc.expectedLine)
			}
		})
	}
}

// echo
func TestRunCommandEcho(t *testing.T) {
	got := runCommand("echo Hello World")
	want := "Hello World"

	if got != want {
		t.Errorf("echo: got %q, want %q", got, want)
	}
}

func TestRunCommandEchoEmpty(t *testing.T) {
	got := runCommand("echo")
	want := ""

	if got != want {
		t.Errorf("echo empty: got %q, want %q", got, want)
	}
}

// unknown
func TestRunCommandUnknown(t *testing.T) {
	got := runCommand("foobar")
	want := "foobar: command not found"

	if got != want {
		t.Errorf("unknown: got %q, want %q", got, want)
	}
}

// exit
func TestRunCommandExit(t *testing.T) {
	got := runCommand("exit")
	want := "exit"

	if got != want {
		t.Errorf("exit: got %q, want %q", got, want)
	}
}

// type
func TestRunCommandTypeInvalid(t *testing.T) {
	got := runCommand("type invalid_command")
	want := "invalid_command: not found"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestRunCommandTypeEcho(t *testing.T) {
	got := runCommand("type echo")
	want := "echo is a shell builtin"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestRunCommandTypeEmpty(t *testing.T) {
	got := runCommand("type ")
	want := ""

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestRunCommandTypeExecutableFile(t *testing.T) {
	got := runCommand("type cat")
	want := "cat is /bin/cat"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestRunCommandTypeNonExist(t *testing.T) {
	got := runCommand("type abc")
	want := "abc: not found"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestExecuteScript(t *testing.T) {
	got := runCommand("ls .")
	want := "main.go\nshell_test.go"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestTypePwd(t *testing.T) {
	got := runCommand("type pwd")
	want := "pwd is a shell builtin"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestPwd(t *testing.T) {
	got := runCommand("pwd")
	want := "/Users/xiaoyuelyu/go/codecrafters-shell-go/app"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestGoToNonExistentAbsoultePath(t *testing.T) {
	got := runCommand("cd /Users/agnes")
	want := "cd: /Users/agnes: No such file or directory"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestGoToAbsolutePath(t *testing.T) {
	runCommand("cd /Users/xiaoyuelyu")
	got := runCommand("pwd")
	want := "/Users/xiaoyuelyu"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestGoToRelativePath(t *testing.T) {
	runCommand("cd ../")
	got := runCommand("pwd")
	want := "/Users/xiaoyuelyu/go/codecrafters-shell-go"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestGoToHomeDir(t *testing.T) {
	runCommand("cd ~")
	got := runCommand("pwd")
	want := "/Users/xiaoyuelyu"

	if got != want {
		t.Errorf("type: got %q, want %q", got, want)
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		// basic cases
		{"echo hello world", []string{"echo", "hello", "world"}},
		{"echo 'hello world'", []string{"echo", "hello world"}},
		{"echo 'hello     shell'", []string{"echo", "hello     shell"}},

		// adjacent quoted + quoted
		{"echo 'example''test'", []string{"echo", "exampletest"}},

		// quoted + unquoted
		{"echo 'script''world'", []string{"echo", "scriptworld"}},
		{"echo hello''world", []string{"echo", "helloworld"}},

		// multiple quoted args
		{"echo '/tmp/file name' '/tmp/file name 1'", []string{
			"echo",
			"/tmp/file name",
			"/tmp/file name 1",
		}},

		// mixing spaces
		{"echo   'a'   b   'c'  d", []string{"echo", "a", "b", "c", "d"}},

		// trailing spaces
		{"echo a b   ", []string{"echo", "a", "b"}},

		// leading spaces
		{"   echo   a   b", []string{"echo", "a", "b"}},

		// single empty quote mid-argument
		{"echo foo''bar", []string{"echo", "foobar"}},
		{"echo ''foo", []string{"echo", "foo"}},
		{"echo foo''", []string{"echo", "foo"}},

		// full tricky sequence
		{"echo 'a''b'c''d''e'f", []string{"echo", "abcdef"}},

		// --- DOUBLE QUOTES ---
		{"echo \"hello world\"", []string{"echo", "hello world"}},
		{"echo \"hello     shell\"", []string{"echo", "hello     shell"}},
		{"echo \"example\"\"test\"", []string{"echo", "exampletest"}},
		{"echo \"a\"\"b\"c\"\"d\"", []string{"echo", "abcd"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseArgs(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("input: %q\nexpected: %#v\ngot:      %#v",
					t.input, tt.expected, got)
			}
		})
	}
}

func TestSingleQuote(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "cat",
			command: "cat '../file name'",
			want:    "hello world!",
		},
		{
			name:    "echo",
			command: "echo 'Hello        World'",
			want:    "Hello        World",
		},
		{
			name:    "adjacent quoted and unquoted content",
			command: "echo hello''world",
			want:    "helloworld",
		},
		{
			name:    "multiple quoted strings",
			command: "echo 'hello shell' 'example''test' script''world",
			want:    "hello shell exampletest scriptworld",
		},
		{
			name:    "escape character outside quotes 1",
			command: "echo world\ \ \ \ \ \ script",
			want:    "world      script",
		},
		{
			name:    "escape chracter outside quotes 2",
			command: "echo test\nexample",
			want:    "testnexample",
		},
		{
			name:    "Backslash within double quotes 1",
			command: "echo \"A \\ escapes itself\"",
			want:    "A \ escapes itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runCommand(tt.command)
			if got != tt.want {
				t.Errorf("single quote failed on %#v = %#v, want %#v", tt.command, got, tt.want)
			}
		})
	}
}