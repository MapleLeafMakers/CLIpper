package cmdinput

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/shlex"
	"sort"
	"strings"
)

type Model struct {
	TextInput   textinput.Model
	History     []string
	HistoryPos  int
	Completions []string
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

func (m *Model) autoComplete() {
	//prefix := m.TextInput.Value()[:m.TextInput.Position()]
	//options := []string{}
	//for _, c := range m.Completions {
	//	if strings.HasPrefix(c, prefix) {
	//		options = append(options, c)
	//	}
	//}
	//
	//if m.tabPresses == 1 {
	//	lcp := longestCommonPrefix(options)
	//	if len(lcp) > 0 {
	//		m.TextInput.SetValue(lcp)
	//		m.TextInput.SetCursor(len(lcp))
	//	}
	//}
	rawInput := m.TextInput.Value()[:m.TextInput.Position()]
	tokens, err := shlex.Split(rawInput)
	if err != nil {
		panic("Probably need to handle incomplete quoted strings or something crazy")
	}
	if strings.HasSuffix(rawInput, " ") || rawInput == "" {
		// if the input ended with a space we're starting the next (empty) token
		tokens = append(tokens, "")
	}

	options := m.TabComplete.Complete(tokens[len(tokens)-1], tokens[:len(tokens)-1])
	if len(options) == 0 {
		return
	}

	if m.tabPresses == 1 {
		//m.TextInput.SetValue(fmt.Sprintf("%+v", options))
		//m.TextInput.SetCursor(len(m.TextInput.Value()))

		lcp := longestCommonPrefix(options)
		newContent := rawInput[:len(rawInput)-(len(tokens[len(tokens)-1]))] + lcp
		if newContent != rawInput {
			m.TextInput.SetValue(newContent)
			m.TextInput.SetCursor(len(m.TextInput.Value()))
			m.tabPresses = 0
		}
	} else {
		// 2nd tab press, initiates some kind of menu...

	}
}

func longestCommonPrefix(strs []string) string {
	var longestPrefix string = ""
	var endPrefix = false

	if len(strs) > 0 {
		sort.Strings(strs)
		first := string(strs[0])
		last := string(strs[len(strs)-1])

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

func (m *Model) NewEntry() {
	m.History = append(m.History, "")
	m.HistoryPos = len(m.History) - 1
	m.TextInput.SetValue("")
}
