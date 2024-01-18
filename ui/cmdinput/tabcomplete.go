package cmdinput

import (
	"strings"
)

type Completer func(string) (bool, []string)

type Completion []Completer

func NewStringCompleter(str string) Completer {
	values := []string{str}
	lowerStr := strings.ToLower(str)
	return func(input string) (bool, []string) {
		if strings.HasPrefix(lowerStr, strings.ToLower(input)) {
			return true, values
		}
		return false, []string{}
	}
}

func NewListCompleter(options ...string) Completer {
	lowerOptions := make([]string, len(options))
	for i, o := range options {
		lowerOptions[i] = strings.ToLower(o)
	}
	return func(input string) (bool, []string) {
		matches := []string{}
		lowerInput := strings.ToLower(input)
		for i, lo := range lowerOptions {
			if strings.HasPrefix(lo, lowerInput) {
				matches = append(matches, options[i])
			}
		}
		return len(matches) > 0, matches
	}
}

type TabComplete struct {
	completions []Completion
}

func NewTabComplete() TabComplete {
	return TabComplete{completions: []Completion{}}
}

func (t *TabComplete) RegisterCompletion(completers ...Completer) {
	t.completions = append(t.completions, completers)
}

func (t *TabComplete) Complete(token string, context []string) []string {
	var options []string
	for _, completion := range t.completions {
		satisfied := true
		// doesn't match because we have more args than it accepts
		if len(context)+1 > len(completion) {
			continue
		}
		// go through each completer and test it on the relevant token
		// bail early if one fails
		for i, completer := range completion {
			if i >= len(context) {
				break
			}
			matched, _ := completer(context[i])
			if !matched {
				satisfied = false
				break
			}
		}

		if satisfied {
			tokenCompleter := completion[len(context)]
			matched, opts := tokenCompleter(token)
			if matched {
				options = append(options, opts...)
			}
		}
	}

	return options
}
