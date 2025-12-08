package main

import (
	"bufio"
	"fmt"
	"os"
)

type History struct {
	File   string
	Items  []string
	MaxLen int
}

func (history *History) Write(cmd string) {
	history.Items = append(history.Items, cmd)
}

func (history *History) GetLine(i int) (string, error) {
	if len(history.Items) < i {
		return "", fmt.Errorf("invalid input")
	}
	return history.Items[i], nil
}

func (history *History) Get() {
	total := len(history.Items)
	start := total - history.MaxLen

	if start < 0 {
		start = 0
	}

	for i := start; i < total; i++ {
		line, err := history.GetLine(i)
		if err != nil {
			continue
		}
		fmt.Printf("%d  %s\n", i, line)
	}
}

func (history *History) ReadFromFile() error {
	file, err := os.Open(history.File)
	if err != nil {
		return fmt.Errorf("error opening history file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		history.Items = append(history.Items, scanner.Text())
	}

	return scanner.Err()
}
