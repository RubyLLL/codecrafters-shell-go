package main

import "strings"

// ParseArgs parses a command string into individual arguments,
// handling single quotes, double quotes, and escape sequences.
func ParseArgs(input string) []string {
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
