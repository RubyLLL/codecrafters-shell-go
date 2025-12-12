package main

import (
	"os"
	"path/filepath"
	"strings"
)

// PathFinder handles PATH resolution and executable lookup
type PathFinder struct {
	paths []string
}

// NewPathFinder creates a new PathFinder with the system PATH
func NewPathFinder() *PathFinder {
	return &PathFinder{
		paths: strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)),
	}
}

// FindExecutable searches for a command in PATH directories
// Returns the full path if found, empty string otherwise
func (pf *PathFinder) FindExecutable(command string) string {
	for _, p := range pf.paths {
		fp := filepath.Join(p, command)
		if info, err := os.Stat(fp); err == nil && info.Mode().IsRegular() && (info.Mode()&0111 != 0) {
			return fp
		}
	}
	return ""
}

// FetchAllExecutables returns all executable files found in PATH directories
func (pf *PathFinder) FetchAllExecutables() []string {
	executables := make(map[string]struct{})

	for _, path := range pf.paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			continue // skip if cannot read
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Check if executable by owner (unix)
			if info.Mode()&0111 != 0 {
				executables[entry.Name()] = struct{}{}
			}
		}
	}

	var result []string
	for exe := range executables {
		result = append(result, exe)
	}

	return result
}

// GetPaths returns the list of PATH directories
func (pf *PathFinder) GetPaths() []string {
	return pf.paths
}
