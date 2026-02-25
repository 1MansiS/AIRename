package rename

type Suggestion struct {
	Name   string
	Reason string
}

type Debug struct {
	Prompt string
}

type Result struct {
	Suggestions []Suggestion
	Debug       Debug
}
