package main

import "testing"

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

func TestRunCommandUnknown(t *testing.T) {
	got := runCommand("foobar")
	want := "foobar: command not found"

	if got != want {
		t.Errorf("unknown: got %q, want %q", got, want)
	}
}

func TestRunCommandExit(t *testing.T) {
	got := runCommand("exit")
	want := "exit"

	if got != want {
		t.Errorf("exit: got %q, want %q", got, want)
	}
}
