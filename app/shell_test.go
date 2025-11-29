package main

import "testing"

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
