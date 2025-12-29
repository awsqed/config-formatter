package formatter

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Formatter is the interface that all config formatters must implement
type Formatter interface {
	// Format formats the YAML data with consistent indentation and ordering
	Format(data []byte, indent int) ([]byte, error)

	// Name returns the name of the formatter
	Name() string

	// CanHandle returns true if this formatter can handle the given file
	CanHandle(filename string, data []byte) bool
}

// BaseFormatter provides common YAML formatting functionality
type BaseFormatter struct{}

// FormatYAML is a helper function that provides basic YAML formatting
func (bf *BaseFormatter) FormatYAML(data []byte, indent int, formatNode func(*yaml.Node, bool)) ([]byte, error) {
	var root yaml.Node
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Apply formatting to the node tree
	formatNode(&root, true)

	// Marshal back to YAML with specified indentation
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

	err = encoder.Encode(&root)
	if err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}
	encoder.Close()

	// Post-process to fix empty lines (remove trailing spaces)
	result := cleanEmptyLines(buf.Bytes())

	return result, nil
}

// cleanEmptyLines removes trailing spaces from empty lines and removes leading empty lines
func cleanEmptyLines(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))

	// Remove trailing spaces from empty lines
	for i, line := range lines {
		// If line only contains spaces, make it truly empty
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 && len(line) > 0 {
			lines[i] = []byte{}
		}
	}

	// Remove leading empty lines
	start := 0
	for start < len(lines) && len(bytes.TrimSpace(lines[start])) == 0 {
		start++
	}
	if start > 0 {
		lines = lines[start:]
	}

	return bytes.Join(lines, []byte("\n"))
}
