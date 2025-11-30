package main

import (
	"reflect"
	"testing"
)

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
					tt.input, tt.expected, got)
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
			command: "echo world\\ \\ \\ \\ \\ \\ script",
			want:    "world      script",
		},
		{
			name:    "escape chracter outside quotes 2",
			command: "echo test\\nexample",
			want:    "testnexample",
		},
		{
			name:    "Backslash within double quotes 1",
			command: "echo \"A \\ escapes itself\"",
			want:    "A \\ escapes itself",
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
