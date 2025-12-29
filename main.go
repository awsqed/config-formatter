package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awsqed/config-formatter/formatter"
	"github.com/awsqed/config-formatter/modules/dockercompose"
	"github.com/awsqed/config-formatter/modules/traefik"
)

var formatters = []formatter.Formatter{
	dockercompose.New(),
	traefik.New(),
}

func main() {
	inputFile := flag.String("input", "", "Input config file (required)")
	outputFile := flag.String("output", "", "Output file (if not specified, prints to stdout)")
	indent := flag.Int("indent", 2, "Number of spaces for indentation")
	inPlace := flag.Bool("w", false, "Write result to source file instead of stdout")
	check := flag.Bool("check", false, "Check if file is formatted without making changes")
	formatterType := flag.String("type", "", "Formatter type to use (docker-compose, traefik). Auto-detected if not specified")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read input file
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Select the appropriate formatter
	var selectedFormatter formatter.Formatter
	filename := filepath.Base(*inputFile)

	if *formatterType != "" {
		// Use specified formatter type
		for _, f := range formatters {
			if f.Name() == *formatterType {
				selectedFormatter = f
				break
			}
		}
		if selectedFormatter == nil {
			fmt.Fprintf(os.Stderr, "Error: unknown formatter type '%s'\n", *formatterType)
			fmt.Fprintln(os.Stderr, "Available formatters:")
			for _, f := range formatters {
				fmt.Fprintf(os.Stderr, "  - %s\n", f.Name())
			}
			os.Exit(1)
		}
	} else {
		// Auto-detect formatter based on file content and name
		for _, f := range formatters {
			if f.CanHandle(filename, data) {
				selectedFormatter = f
				break
			}
		}
		if selectedFormatter == nil {
			fmt.Fprintln(os.Stderr, "Error: could not auto-detect config type")
			fmt.Fprintln(os.Stderr, "Please specify formatter type with -type flag")
			fmt.Fprintln(os.Stderr, "Available formatters:")
			for _, f := range formatters {
				fmt.Fprintf(os.Stderr, "  - %s\n", f.Name())
			}
			os.Exit(1)
		}
	}

	// Format the config file
	formatted, err := selectedFormatter.Format(data, *indent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting file: %v\n", err)
		os.Exit(1)
	}

	// Check mode
	if *check {
		if string(data) != string(formatted) {
			fmt.Fprintf(os.Stderr, "File is not formatted (detected as %s)\n", selectedFormatter.Name())
			os.Exit(1)
		}
		fmt.Printf("File is formatted (detected as %s)\n", selectedFormatter.Name())
		return
	}

	// Determine output destination
	var output string
	if *inPlace {
		output = *inputFile
	} else if *outputFile != "" {
		output = *outputFile
	}

	// Write output
	if output != "" {
		err = os.WriteFile(output, formatted, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Formatted file written to: %s (using %s formatter)\n", output, selectedFormatter.Name())
	} else {
		fmt.Print(string(formatted))
	}
}
