package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/awsqed/docker-compose-formatter/formatter"
)

func main() {
	inputFile := flag.String("input", "", "Input docker-compose file (required)")
	outputFile := flag.String("output", "", "Output file (if not specified, prints to stdout)")
	indent := flag.Int("indent", 2, "Number of spaces for indentation")
	inPlace := flag.Bool("w", false, "Write result to source file instead of stdout")
	check := flag.Bool("check", false, "Check if file is formatted without making changes")

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

	// Format the docker-compose file
	formatted, err := formatter.Format(data, *indent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting file: %v\n", err)
		os.Exit(1)
	}

	// Check mode
	if *check {
		if string(data) != string(formatted) {
			fmt.Fprintf(os.Stderr, "File is not formatted\n")
			os.Exit(1)
		}
		fmt.Println("File is formatted")
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
		fmt.Printf("Formatted file written to: %s\n", output)
	} else {
		fmt.Print(string(formatted))
	}
}
