package rename

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CodeStylePolicy is your strict style instructions
const CodeStylePolicy = `STRICT OUTPUT REQUIREMENTS:

- Output exactly 3 lines.
- Do NOT include any introductory sentence.
- Do NOT explain your reasoning.
- Do NOT restate the task.
- Each line must follow this exact format:
- Variable names must be concise and idiomatic.
- Prefer conventional short identifiers (n, i, j, a, b, err, ctx, req, resp, fib).
- Do NOT use verbose tutorial-style names.
- If a shorter conventional identifier exists, use it.
- Avoid multi-word identifiers unless absolutely necessary.
- Names should typically be 1-2 words max.
- The justification must be under 5 words.
- No extra commentary.
- No blank lines.

<name> - <very short justification (max 5 words)>
`

// CallLLM dispatches to the appropriate LLM provider and retries until
// exactly 3 output lines are returned. provider is "ollama" or "claude".
func CallLLM(taskPrompt, provider string) ([]string, error) {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var lines []string
		var err error

		switch provider {
		case "claude":
			lines, err = runClaude(taskPrompt)
		default:
			lines, err = runOllama(taskPrompt)
		}
		if err != nil {
			return nil, err
		}

		if len(lines) == 3 {
			return lines, nil
		}

		fmt.Fprintf(os.Stderr, "[llm] retry %d: got %d lines, expecting 3\n", attempt, len(lines))
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to get exactly 3 lines after %d attempts", maxRetries)
}

// runClaude shells out to the `claude` CLI (Claude Code) using the OAuth
// session already established by the user — no API key required.
func runClaude(taskPrompt string) ([]string, error) {
	cmd := exec.Command("claude", "-p", CodeStylePolicy+"\n\n"+taskPrompt)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("[llm] claude error: %s", stderr.String())
		return nil, err
	}

	return parseRenameOutput(stdout.String()), nil
}

// runOllama executes the CLI and parses stdout
func runOllama(prompt string) ([]string, error) {
	cmd := exec.Command(
		"ollama",
		"run",
		"llama3:8b",
		CodeStylePolicy+"\n\n"+prompt,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("[llm] Ollama error: %s", stderr.String())
		return nil, err
	}

	raw := stdout.String()

	return parseRenameOutput(raw), nil
}

// parseRenameOutput trims empty lines and returns lines
func parseRenameOutput(raw string) []string {
	lines := strings.Split(raw, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
