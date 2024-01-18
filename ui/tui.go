package ui

import (
	"clipper/jsonrpcclient"
	"clipper/ui/cmdinput"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wrap"
	"github.com/muesli/termenv"
	"os"
	"reflect"
	"strings"
	"time"
)

var ()

type LogEntry struct {
	timestamp string
	message   string
}

type UIOptions struct {
	LogIncoming     bool
	TimestampFormat string
	BorderColor     lipgloss.TerminalColor
}

type Model struct {
	Content []LogEntry

	Ready    bool
	Viewport viewport.Model
	Input    cmdinput.Model
	Client   *jsonrpcclient.Client
	output   *termenv.Output
	Options  *UIOptions

	formattedTitle  string
	renderedContent string

	inputStyle    lipgloss.Style
	viewportStyle lipgloss.Style
}

func formatTitle(output *termenv.Output, version string) string {
	return output.String("🚣 CLIpper").Foreground(output.Color("#FFFF00")).String() + " " + version
}

func NewTUI(client *jsonrpcclient.Client, version string) *tea.Program {
	output := termenv.NewOutput(os.Stdout)
	model := Model{
		Client:         client,
		Content:        []LogEntry{},
		output:         output,
		formattedTitle: formatTitle(output, version),
	}
	model.LoadOptions()
	program := tea.NewProgram(model)
	return program
}

func (m *Model) generateStyles() {
	b := lipgloss.RoundedBorder()
	b.BottomRight = "┤"
	b.BottomLeft = "├"
	m.viewportStyle = lipgloss.NewStyle().BorderStyle(b).BorderForeground(m.Options.BorderColor).Padding(0, 1)
	m.inputStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).BorderLeft(true).BorderRight(true).BorderBottom(true).BorderForeground(m.Options.BorderColor)
}

func (m *Model) LoadOptions() {
	m.Options = &UIOptions{
		LogIncoming:     false,
		TimestampFormat: time.Kitchen,
		BorderColor:     lipgloss.Color("#00FF00"),
	}
	m.generateStyles()
}

func (m *Model) getLogTimestamp() string {
	return m.output.String(time.Now().Format(m.Options.TimestampFormat)).Foreground(m.output.Color("#666666")).String()
}

func (m *Model) readIncoming() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.Client.Incoming
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) renderEntry(logEntry LogEntry) string {
	timestampWidth := ansi.PrintableRuneWidth(logEntry.timestamp)
	pad := strings.Repeat(" ", timestampWidth)
	w := m.Viewport.Width - (timestampWidth + 5)
	wrapped := wrap.String(logEntry.message, w)
	out := ""
	div := m.output.String(" │ ").Foreground(m.output.Color("#666666")).String()
	for i, line := range strings.Split(wrapped, "\n") {
		if i == 0 {
			out += logEntry.timestamp + div + line + "\n"
		} else {
			out += pad + div + line + "\n"
		}
	}

	return out
}

func (m *Model) renderContent() string {
	rendered := ""
	//l := len(m.Content)
	timestampWidth := ansi.PrintableRuneWidth(m.getLogTimestamp())
	div := strings.Repeat("─", timestampWidth+1) + "┼" + strings.Repeat("─", m.Viewport.Width-(timestampWidth+4))
	div = m.output.String(div).Foreground(m.output.Color("#666666")).String()
	for i, logEntry := range m.Content {
		if i > 0 {
			rendered = rendered + div + "\n"
		}
		rendered = rendered + m.renderEntry(logEntry)

	}
	m.renderedContent = strings.Trim(rendered, "\n")
	return m.renderedContent
}

func (m *Model) AppendLog(logEntry LogEntry) {
	first := len(m.Content) == 0
	m.Content = append(m.Content, logEntry)
	if !first {
		timestampWidth := ansi.PrintableRuneWidth(logEntry.timestamp)
		m.renderedContent = m.renderedContent + "\n"
		div := strings.Repeat("─", timestampWidth+1) + "┼" + strings.Repeat("─", m.Viewport.Width-(timestampWidth+4))
		m.renderedContent = m.renderedContent + m.output.String(div).Foreground(m.output.Color("#666666")).String() + "\n"
	}
	m.renderedContent = m.renderedContent + strings.Trim(m.renderEntry(logEntry), "\n")
	m.Viewport.SetContent(m.renderedContent)
	m.Viewport.GotoBottom()
}

func (m *Model) handleIncoming(req jsonrpcclient.IncomingJsonRPCRequest) {
	switch req.Method {
	case "notify_gcode_response":
		for _, line := range req.Params {
			m.AppendLog(LogEntry{timestamp: m.getLogTimestamp(), message: line.(string)})
		}
	default:
		if m.Options.LogIncoming {
			encoded, _ := json.Marshal(req)
			m.AppendLog(LogEntry{message: string(encoded), timestamp: m.getLogTimestamp()})
		}
	}
}

func (m *Model) cmdSet(rawArgs string) {
	args := strings.Fields(rawArgs)
	r := reflect.ValueOf(m.Options)
	f := reflect.Indirect(r).FieldByName(args[0])

	switch f.Kind() {
	case reflect.Bool:
		var b bool
		err := json.Unmarshal([]byte(args[1]), &b)
		if err != nil {
			m.AppendLog(LogEntry{timestamp: m.getLogTimestamp(), message: "Invalid boolean value: " + args[1]})
		} else {
			f.SetBool(b)
		}
	case reflect.Interface:
		colorType := reflect.TypeOf((*lipgloss.TerminalColor)(nil)).Elem()
		if f.Type().Implements(colorType) {
			f.Set(reflect.ValueOf(lipgloss.Color(args[1])))
			m.generateStyles()
		}
	default:
		m.AppendLog(LogEntry{timestamp: m.getLogTimestamp(), message: "Kind " + f.Kind().String()})
	}
}

func (m *Model) cmdClear() {
	m.Content = []LogEntry{}
	m.renderedContent = ""
	m.Viewport.SetContent("")
}

func (m *Model) processCommand(input string) {
	split := strings.SplitN(input, " ", 2)
	cmd := strings.ToLower(split[0])
	rawArgs := ""
	if len(split) == 2 {
		rawArgs = split[1]
	}
	switch cmd {
	case "set":
		m.cmdSet(rawArgs)
	case "clear":
		m.cmdClear()
	}

}

func (m *Model) RegisterCompleters() {
	// /clear
	m.Input.TabComplete.RegisterCompletion(cmdinput.NewStringCompleter("/clear"))

	// /set
	r := reflect.TypeOf(*m.Options)
	optionKeys := make([]string, r.NumField())
	for i := 0; i < len(optionKeys); i++ {
		optionKeys[i] = r.Field(i).Name
	}
	m.Input.TabComplete.RegisterCompletion(
		cmdinput.NewStringCompleter("/set"),
		cmdinput.NewListCompleter(optionKeys...),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		} else if key == "pgup" || key == "pgdown" {
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else if key == "enter" {
			val := m.Input.Value()
			if val != "" {
				if strings.HasPrefix(val, "/") {
					m.processCommand(val[1:])
				} else {
					m.AppendLog(LogEntry{message: m.Input.Value(), timestamp: m.getLogTimestamp()})
					go m.Client.Call("printer.gcode.script", map[string]interface{}{"script": val})
				}
				m.Input.NewEntry()
			}
		} else {
			m.Input, cmd = m.Input.Update(msg)
			cmds = append(cmds, cmd)
		}
	case jsonrpcclient.IncomingJsonRPCRequest:
		m.handleIncoming(msg)
		cmds = append(cmds, m.readIncoming())

	case tea.WindowSizeMsg:
		verticalMarginHeight := 4
		if !m.Ready {
			m.Viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.Viewport.YPosition = 0
			m.Viewport.SetContent(m.renderContent())
			m.Input = cmdinput.New()
			m.RegisterCompleters()
			m.Input.TextInput.Width = msg.Width - 40
			m.Ready = true
			cmds = append(cmds, m.readIncoming())

		} else {
			m.Viewport.Width = msg.Width - 4
			m.Viewport.Height = msg.Height - verticalMarginHeight
			m.Input.TextInput.Width = msg.Width - 4
			m.Viewport.SetContent(m.renderContent())
		}
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)

	default:
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	vp := m.viewportStyle.Width(m.Viewport.Width).Render(m.Viewport.View())
	// replace the top left corner with a title bar
	borderStyle := lipgloss.NewStyle().Foreground(m.Options.BorderColor)
	title := borderStyle.Render("╭─┤ ") + (m.formattedTitle) + borderStyle.Render(" ├")
	titleW := ansi.PrintableRuneWidth(title)
	_, rest := AnsiSplitAt(vp, titleW)
	vp = title + borderStyle.Render(rest)
	inp := m.inputStyle.Width(m.Viewport.Width).Render(m.Input.View())
	return fmt.Sprintf("%s\n%s", vp, inp)
}

func AnsiSplitAt(input string, chars int) (string, string) {
	var n int
	var isAnsi bool

	for i, c := range input {
		if c == ansi.Marker {
			isAnsi = true
		} else if isAnsi {
			if ansi.IsTerminator(c) {
				isAnsi = false
			}
		} else {
			n += runewidth.RuneWidth(c)
			if n == chars+1 {
				return input[:i], input[i:]
			}
		}
	}
	return input, ""
}
