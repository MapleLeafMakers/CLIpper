package cmdinput

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/shlex"
	"sort"
	"strings"
)

type Model struct {
	TextInput           textinput.Model
	History             []string
	HistoryPos          int
	Completions         []string
	CompletionPos       int
	CompletionOptions   []string
	CompletionRemainder string
	CompletingFrom      string

	tabPresses  int
	TabComplete *TabComplete
}

func New() Model {
	tc := NewTabComplete()
	m := Model{
		TextInput:   textinput.New(),
		History:     []string{""},
		HistoryPos:  0,
		TabComplete: &tc,
	}
	m.TextInput.Prompt = "> "

	m.TextInput.Focus()

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	if msg, ok := msg.(tea.KeyMsg); ok {

		// Count the tab presses for autocomplete purposes later
		if msg.String() == "tab" {
			m.tabPresses++
			if m.tabPresses > 2 {
				m.tabPresses = 2
			}
		} else {
			m.tabPresses = 0
		}

		switch msg.String() {
		case "up":
			m.HistoryBack()
		case "down":
			m.HistoryForward()
		case "tab":
			m.autoComplete()
		default:
			m.TextInput, cmd = m.TextInput.Update(msg)
			cmds = append(cmds, cmd)
			m.History[len(m.History)-1] = m.TextInput.Value()
		}
		return m, tea.Batch(cmds...)
	}
	m.TextInput, cmd = m.TextInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) HistoryBack() {
	m.HistoryPos--
	if m.HistoryPos < 0 {
		m.HistoryPos = 0
	}
	m.SetValue(m.History[m.HistoryPos])
	m.TextInput.SetCursor(len(m.Value()))
}

func (m *Model) HistoryForward() {
	m.HistoryPos++
	if m.HistoryPos >= len(m.History) {
		m.HistoryPos = len(m.History) - 1
	}
	//log.Println(len(m.History), m.HistoryPos)
	m.TextInput.SetValue(m.History[m.HistoryPos])
	m.TextInput.SetCursor(len(m.Value()))
}

func splitToTokens(input string) []string {
	tokens, err := shlex.Split(input)
	if err != nil {
		panic("Probably need to handle incomplete quoted strings or something crazy")
	}
	if strings.HasSuffix(input, " ") || input == "" {
		// if the input ended with a space we're starting the next (empty) token
		tokens = append(tokens, "")
	}
	return tokens
}

func (m *Model) autoComplete() {
	if m.tabPresses == 1 {
		rawInput := m.TextInput.Value()[:m.TextInput.Position()]
		remainder := m.TextInput.Value()[m.TextInput.Position():]
		tokens := splitToTokens(rawInput)
		options := m.TabComplete.Complete(tokens[len(tokens)-1], tokens[:len(tokens)-1])
		if len(options) == 0 {
			return
		}
		lcp := longestCommonPrefix(options)
		newContent := rawInput[:len(rawInput)-(len(tokens[len(tokens)-1]))] + lcp
		m.CompletingFrom = rawInput[:len(rawInput)-(len(tokens[len(tokens)-1]))]
		m.CompletionOptions = options
		m.CompletionRemainder = remainder
		m.TextInput.SetValue(newContent)
		m.TextInput.SetCursor(len(m.TextInput.Value()))
		m.TextInput.SetValue(newContent + remainder)
	} else {
		m.CompletionPos++
		if m.CompletionPos >= len(m.CompletionOptions) {
			m.CompletionPos = 0
		}
		newContent := m.CompletingFrom + m.CompletionOptions[m.CompletionPos]
		m.TextInput.SetValue(newContent)
		m.TextInput.SetCursor(len(m.TextInput.Value()))
		m.TextInput.SetValue(newContent + m.CompletionRemainder)
		m.tabPresses = 1
	}
}

func longestCommonPrefix(strs []string) string {
	var longestPrefix = ""
	var endPrefix = false

	if len(strs) > 0 {
		sort.Strings(strs)
		first := strs[0]
		last := strs[len(strs)-1]

		for i := 0; i < len(first); i++ {
			if !endPrefix && string(last[i]) == string(first[i]) {
				longestPrefix += string(last[i])
			} else {
				endPrefix = true
			}
		}
	}
	return longestPrefix
}

func (m *Model) View() string {

	return m.TextInput.View() // + strconv.Itoa(m.tabPresses)
}

func (m Model) Value() string {
	return m.TextInput.Value()
}

func (m *Model) SetValue(value string) {
	m.TextInput.SetValue(value)
}

func (m *Model) SetTextStyle(style lipgloss.Style) {
	m.TextInput.TextStyle = style
}

func (m *Model) NewEntry() {
	m.History = append(m.History, "")
	m.HistoryPos = len(m.History) - 1
	m.TextInput.SetValue("")
}
