package cmdinput

import (
	"errors"
	"github.com/gdamore/tcell/v2"
	"github.com/google/shlex"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type CompletionState struct {
	RawText   string
	CursorPos int
	tokens    []string
}

type CommandContext map[string]interface{}

type TokenCompleter interface {
	Match(token string, ctx CommandContext) (bool, string, *TokenCompleter)
	Complete(token string, ctx CommandContext) ([]Suggestion, bool)
}

type TabCompleter struct {
	completer       CommandTokenCompleter
	completionState CompletionState
}

func NewTabCompleter() TabCompleter {
	return TabCompleter{
		completer: CommandTokenCompleter{
			Registry:   map[string]Command{},
			caseMap:    map[string]string{},
			ContextKey: "cmd",
		},
	}
}

func (t *TabCompleter) RegisterCommand(cmd string, command Command) {
	t.completer.caseMap[strings.ToLower(cmd)] = cmd
	t.completer.Registry[cmd] = command
}

func (t *TabCompleter) AutoComplete(currentText string, cursorPos int, ctx CommandContext) (entries []Suggestion, menuOffset int) {
	inText := currentText[:cursorPos]
	tokens, err := shlex.Split(inText)
	if err != nil {
		return entries, menuOffset
	}
	if len(tokens) == 0 {
		return entries, menuOffset
	}
	if strings.HasSuffix(inText, " ") {
		tokens = append(tokens, "")
	}

	t.completionState = CompletionState{currentText, cursorPos, tokens}
	lastTokenIdx := len(tokens) - 1

	var currentCompleter TokenCompleter = t.completer

	for i, token := range tokens {
		if i == lastTokenIdx {
			// this is the one we're completing
			results, _ := currentCompleter.Complete(token, ctx)
			return results, strings.LastIndex(inText, tokens[lastTokenIdx])
		} else {
			// these tokens are ones we're just matching through to get to the right one
			match, _, next := currentCompleter.Match(token, ctx)
			if !match {
				// bail out, no completions match
				break
			}
			if *next == nil {
				// we're done
				break
			}
			currentCompleter = *next
		}
	}
	return entries, menuOffset
}

func (t *TabCompleter) OnAutoCompleted(text string, index, source int) (closeMenu bool, fullText string, cursorPos int) {
	switch source {
	case AutocompletedNavigate:
		return false, t.completionState.RawText, t.completionState.CursorPos
	default:
		currentText := t.completionState.RawText
		inText := currentText[:t.completionState.CursorPos]
		afterText := currentText[t.completionState.CursorPos:]
		preText := inText[:strings.LastIndex(inText, t.completionState.tokens[len(t.completionState.tokens)-1])]
		return true, preText + text + afterText, len(preText) + len(text)
	}
}

func (t *TabCompleter) Parse(currentText string, ctx CommandContext) error {
	tokens, _ := shlex.Split(currentText)
	if len(tokens) == 0 {
		return errors.New("NoInput")
	}

	var currentCompleter TokenCompleter = t.completer

	for _, token := range tokens {
		match, _, next := currentCompleter.Match(token, ctx)
		if !match {
			// bail out, no completions match
			return errors.New("No Match for " + token)
		}
		if next == nil {
			// we're done
			break
		}
		currentCompleter = *next
	}
	return nil
}

//
// Token Completers
//

type Command interface {
	Call(ctx CommandContext) error
	GetCompleter(ctx CommandContext) TokenCompleter
	GetHelp() string
}

type CommandTokenCompleter struct {
	ContextKey string
	Registry   map[string]Command
	caseMap    map[string]string
}

func (c CommandTokenCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	key, ok := c.caseMap[strings.ToLower(token)]
	if !ok {
		if !strings.HasPrefix(token, "/") {
			// we should treat this as gcode.
			var completer TokenCompleter = AnythingCompleter{"value"}
			return true, token, &completer
		} else {
			return false, "", nil
		}

	}
	cmd := c.Registry[key]
	ctx[c.ContextKey] = cmd
	completer := cmd.GetCompleter(ctx)
	return true, token, &completer
}

func (c CommandTokenCompleter) Complete(token string, ctx CommandContext) (results []Suggestion, match bool) {
	lowerToken := strings.ToLower(token)
	sortedKeys := make([]string, 0, len(c.caseMap))
	for k, _ := range c.caseMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, lowerCmdName := range sortedKeys {
		if strings.HasPrefix(lowerCmdName, lowerToken) {
			if lowerCmdName == lowerToken {
				match = true
			}
			results = append(results, Suggestion{c.caseMap[lowerCmdName], c.Registry[c.caseMap[lowerCmdName]].GetHelp()})
		}
	}
	return results, match
}

type StaticTokenCompleter struct {
	ContextKey string
	Registry   map[string]TokenCompleter
}

func (c StaticTokenCompleter) buildCaseMap() map[string]string {
	m := make(map[string]string, len(c.Registry))
	for k, _ := range c.Registry {
		m[strings.ToLower(k)] = k
	}
	return m
}

func (c StaticTokenCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	caseMap := c.buildCaseMap()

	normalizedName, ok := caseMap[strings.ToLower(token)]
	if !ok {
		return false, "", nil
	}
	ctx[c.ContextKey] = normalizedName
	completer := c.Registry[normalizedName]
	return true, normalizedName, &completer
}

func (c StaticTokenCompleter) Complete(token string, ctx CommandContext) (results []Suggestion, match bool) {
	lowerToken := strings.ToLower(token)
	caseMap := c.buildCaseMap()
	sortedKeys := make([]string, 0, len(c.Registry))
	for k, _ := range caseMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, lowerCmdName := range sortedKeys {
		if strings.HasPrefix(lowerCmdName, lowerToken) {
			if lowerCmdName == lowerToken {
				match = true
			}
			results = append(results, Suggestion{caseMap[lowerCmdName], ""})
		}
	}
	return results, match
}

type BoolTokenCompleter struct {
	ContextKey string
	Next       TokenCompleter
}

func (c BoolTokenCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	t := strings.ToLower(token)
	if t == "true" {
		ctx[c.ContextKey] = true
	} else if t == "false" {
		ctx[c.ContextKey] = false
	}
	return true, t, &c.Next
}

func (c BoolTokenCompleter) Complete(token string, ctx CommandContext) (results []Suggestion, match bool) {
	lowerToken := strings.ToLower(token)
	for _, lower := range []string{"false", "true"} {
		if strings.HasPrefix(lower, lowerToken) {
			if lower == lowerToken {
				match = true
			}
			results = append(results, Suggestion{lower, ""})
		}
	}
	return results, match
}

func NewBoolTokenCompleter(contextKey string, nextCompleter TokenCompleter) BoolTokenCompleter {
	return BoolTokenCompleter{contextKey, nextCompleter}
}

// ColorTokenCompleter
type ColorTokenCompleter struct {
	ContextKey string
	Next       TokenCompleter
}

var hexColorRegexp = regexp.MustCompile("^#[0-9a-f]{6}$")

func NewColorTokenCompleter(contextKey string, nextCompleter TokenCompleter) ColorTokenCompleter {
	return ColorTokenCompleter{contextKey, nextCompleter}
}

func (c ColorTokenCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	color := tcell.GetColor(token)
	if color == tcell.ColorDefault && strings.ToLower(token) != "default" {
		return false, "", nil
	}
	ctx[c.ContextKey] = token
	return true, token, &c.Next
}

func (c ColorTokenCompleter) Complete(token string, ctx CommandContext) (results []Suggestion, match bool) {
	lowerToken := strings.ToLower(token)
	sortedKeys := make([]string, 0, len(tcell.ColorNames)+1)
	sortedKeys = append(sortedKeys, "default")
	for k, _ := range tcell.ColorNames {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, colorName := range sortedKeys {
		if strings.HasPrefix(colorName, lowerToken) {
			if colorName == lowerToken {
				match = true
			}
			results = append(results, Suggestion{colorName, ""})
		}
	}
	return results, match
}

// FileTokenCompleter

type FileTokenCompleter struct {
	ContextKey string
	Next       TokenCompleter
}

func NewFileTokenCompleter(contextKey string, nextCompleter TokenCompleter) FileTokenCompleter {
	return FileTokenCompleter{
		contextKey, nextCompleter,
	}
}

func (f FileTokenCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	ctx[f.ContextKey] = token
	return true, token, &f.Next
}

func (f FileTokenCompleter) Complete(token string, ctx CommandContext) (result []Suggestion, match bool) {
	var pattern string
	if filepath.IsLocal(token) {
		pattern = filepath.Clean(token)
	} else {
		pattern = token
	}
	matches, _ := filepath.Glob(pattern + "*")
	results := make([]Suggestion, len(matches))
	for i := 0; i < len(matches); i++ {
		results[i] = Suggestion{matches[i], ""}
	}
	return results, false
}

type AnythingCompleter struct {
	ContextKey string
}

func (a AnythingCompleter) Match(token string, ctx CommandContext) (bool, string, *TokenCompleter) {
	value, ok := ctx[a.ContextKey]
	if ok {
		token = (value.(string)) + " " + token
	}
	ctx[a.ContextKey] = token
	var next TokenCompleter
	next = a
	return true, token, &next
}

func (a AnythingCompleter) Complete(token string, ctx CommandContext) (result []Suggestion, match bool) {
	return []Suggestion{}, false
}
