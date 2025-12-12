package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// BellWrapper wraps readline's AutoCompleter to provide custom tab completion behavior
type BellWrapper struct {
	Inner    readline.AutoCompleter
	tabPress bool
	rl       *readline.Instance
}

// Do handles tab completion logic
func (w *BellWrapper) Do(line []rune, pos int) ([][]rune, int) {
	if w.Inner == nil {
		fmt.Fprint(os.Stdout, "\x07")
		return nil, 0
	}

	matches, offset := w.Inner.Do(line, pos)

	// remove duplicates
	matches = removeDuplicates(matches)

	// sort matches
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

// OnChange resets tab state when user types something other than tab
func (w *BellWrapper) OnChange(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
	if key != '\t' {
		w.tabPress = false
	}
	return nil, 0, false
}

// removeDuplicates removes duplicate entries from matches
func removeDuplicates(matches [][]rune) [][]rune {
	seen := make(map[string]struct{})
	uniqueMatches := make([][]rune, 0, len(matches))

	for _, match := range matches {
		s := string(match)
		if _, found := seen[s]; !found {
			seen[s] = struct{}{}
			uniqueMatches = append(uniqueMatches, match)
		}
	}

	return uniqueMatches
}

// longestCommonPrefix finds the longest common prefix among all items
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
		return "", 0
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

// SetupCompleter creates and configures tab completion
func SetupCompleter(builtins *BuiltinCommands, pathFinder *PathFinder) (*BellWrapper, error) {
	executableFiles := pathFinder.FetchAllExecutables()
	allCommands := append(builtins.GetCommandNames(), executableFiles...)

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

	return completer, nil
}
