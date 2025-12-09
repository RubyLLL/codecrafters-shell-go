# Overview
This program is an implementation of shell written in Golang. It supports some built-in commands, such as `eho`/`exit`/`pwd`/`cd`/`history` and so on, and executing files found in `$PATH`. It also supports piping, redirecting output to a file, autocompletion and persistent history.

# Build and run
This program can be run by executing `go run app/main.go app/history.go`, it then prompts the user to enter commands just like a real shell would do. The program can be exited by typing `exit` and hit enter.

# Test
To test the correctness of the implementation, just execute `codecrafters test` and observe the output. codecrafters is a tool that runs test cases against the program and shows how the expected output and the actual output diverge when there are failed tests.

# Your task
In this stage, you'll implement support for pipelines involving more than two commands.

Pipelines can chain multiple commands together, connecting the output of each command to the input of the next one.

Desired behavior:
```shell
$ cat /tmp/foo/file | head -n 5 | wc
       5       5      10
$ ls -la /tmp/foo | tail -n 5 | head -n 3 | grep "file"
-rw-r--r-- 1 user user     5 Apr 29 10:06 file
$
```

## notes
- You should always try to implement the solution with minimal changes to the codebase.
- This requires managing multiple pipes and processes.
- Ensure correct setup of stdin/stdout for each command in the chain (except the first command's stdin and the last command's stdout, which usually connect to the terminal or file redirections).
- Proper process cleanup and waiting are crucial.