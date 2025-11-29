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
	fmt.Fprint(os.Stdout, "$ ")
	readUserCommand()
}

func readUserCommand() {
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	check(err)

	command = strings.TrimSpace(command)
	fmt.Printf("%s: command not found", command)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
