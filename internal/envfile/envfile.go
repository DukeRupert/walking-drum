// Package envfile loads KEY=VALUE pairs from a .env file into the process
// environment. It supports comments (lines starting with #) and blank lines.
// Existing environment variables are never overwritten.
package envfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Load reads the .env file at path and sets each KEY=VALUE pair as an
// environment variable. Existing environment variables are not overwritten,
// so values set by the shell or platform take precedence over the file.
//
// Returns an error if the file cannot be opened or contains malformed lines.
func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	pairs, err := Parse(f)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	for key, value := range pairs {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("setenv %s: %w", path, err)
		}
	}

	return nil
}

// Parse reads a .env-formatted stream and returns the parsed key/value pairs.
// Blank lines and comments are skipped. Malformed lines return an error
// annotated with the line number.
func Parse(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		key, value, ok, err := parseLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		if !ok {
			continue
		}
		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return result, nil
}

// parseLine extracts a key and value from a single line of a .env file.
// Returns ok=false for blank lines and comments (these should be skipped).
// Returns an error for malformed lines.
func parseLine(line string) (key, value string, ok bool, err error) {
	line = strings.TrimSpace(line)

	// skip blanks and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}

	// split on first '=', lines without a '=' are malformed
	idx := strings.IndexByte(line, '=')
	if idx == -1 {
		return "", "", false, fmt.Errorf("missing '=' in line: %q", line)
	}

	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])

	if key == "" {
		return "", "", false, fmt.Errorf("empty key in line: %q", line)
	}

	return key, value, true, nil
}
