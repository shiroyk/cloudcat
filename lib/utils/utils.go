// Package utils the helpers
package utils

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pair a key and a value pair
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// ZeroOr if value is zero value returns the defaultValue
func ZeroOr[T comparable](value, defaultValue T) T {
	var zero T
	if zero == value {
		return defaultValue
	}
	return value
}

// EmptyOr if slice is empty returns the defaultValue
func EmptyOr[T any](value, defaultValue []T) []T {
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// ExpandPath expands path "." or "~"
func ExpandPath(path string) (string, error) {
	// expand local directory
	if strings.HasPrefix(path, ".") {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, path[1:]), nil
	}
	// expand ~ as shortcut for home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// ReadYaml read the YAML file and convert it to T
func ReadYaml[T any](path string) (t T, err error) {
	path, err = ExpandPath(path)
	if err != nil {
		return
	}
	bytes, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return
	}

	err = yaml.Unmarshal(bytes, &t)
	if err != nil {
		return
	}

	return
}
