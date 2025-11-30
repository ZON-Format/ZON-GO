// ZON CLI - Command line tool for encoding and decoding ZON format
//
// Copyright (c) 2025 ZON-FORMAT (Roni Bhakta)
// License: MIT
//
// Usage:
//
//	zon encode <file.json>   # Encode JSON to ZON format
//	zon decode <file.zonf>   # Decode ZON to JSON format
//
// Example:
//
//	zon encode data.json > data.zonf
//	zon decode data.zonf > output.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	zon "github.com/ZON-Format/zon-go"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	inputFile := os.Args[2]

	// Resolve absolute path
	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "encode":
		var data any
		if err := json.Unmarshal(content, &data); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		encoded, err := zon.Encode(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(encoded)

	case "decode":
		decoded, err := zon.Decode(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding: %v\n", err)
			os.Exit(1)
		}

		output, err := json.MarshalIndent(decoded, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(output))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: zon <encode|decode> <file>")
	fmt.Fprintln(os.Stderr, "Example: zon encode data.json > data.zonf")
}
