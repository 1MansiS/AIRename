package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ai_rename/internal/rename"
)

type jsonSuggestion struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type jsonOutput struct {
	Suggestions []jsonSuggestion `json:"suggestions"`
}

func main() {
	provider := flag.String("llm", "ollama", "LLM provider: ollama or claude")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: ai_rename_bin [-llm ollama|claude] <file.go> <row:col>")
		os.Exit(1)
	}

	filePath := args[0]
	parts := strings.SplitN(args[1], ":", 2)
	if len(parts) != 2 {
		fmt.Fprintln(os.Stderr, "selector must be <row>:<col>")
		os.Exit(1)
	}

	row, err := strconv.Atoi(parts[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "row must be an integer")
		os.Exit(1)
	}
	col, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "col must be an integer")
		os.Exit(1)
	}

	result, err := rename.Run(filePath, rename.Selector{
		Kind: "position",
		Row:  row,
		Col:  col,
	}, *provider)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var suggs []jsonSuggestion
	for _, s := range result.Suggestions {
		suggs = append(suggs, jsonSuggestion{Name: s.Name, Reason: s.Reason})
	}

	if err := json.NewEncoder(os.Stdout).Encode(jsonOutput{Suggestions: suggs}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
