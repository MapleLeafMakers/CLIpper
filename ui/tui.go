package ui

import (
	"clipper/jsonrpcclient"
	"clipper/ui/cmdinput"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bykof/gostradamus"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wrap"
	"github.com/muesli/termenv"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var ()

type LogEntry struct {
	timestamp string
	message   string
}

type UIOptions struct {
	LogIncoming     bool
	TimestampFormat string

	BorderType       lipgloss.Border
	BorderForeground lipgloss.TerminalColor
	BorderBackground lipgloss.TerminalColor

	InputForeground lipgloss.TerminalColor
	InputBackground lipgloss.TerminalColor
}

type Model struct {
	Content []LogEntry

	Ready    bool
	Viewport viewport.Model
	Input    cmdinput.Model
	TitleBar TitleBar

	Client  *jsonrpcclient.Client
	output  *termenv.Output
	Options *UIOptions

	host            string
	version         string
	renderedContent string

	inputBorderStyle    lipgloss.Style
	viewportBorderStyle lipgloss.Style
	inlineBorderStyle   lipgloss.Style
	titleStyle          lipgloss.Style
	inputTextStyle      lipgloss.Style // broken

	gcodeHelp map[string]string
}

func formatTitle(output *termenv.Output, version string) string {
	if matched, err := regexp.MatchString("^v\\d+\\.\\d+\\.\\d+-[0-9a-f]{7}$", version); err == nil && matched {
		version = strings.Split(version, "-")[0]
	}
	return output.String("🚣 CLIpper " + version).Foreground(output.Color("#FFFF00")).String()

}

func NewTUI(client *jsonrpcclient.Client, version string) *tea.Program {
	output := termenv.NewOutput(os.Stdout)
	model := Model{
		Client:  client,
		Content: []LogEntry{},
		output:  output,
		version: version,
	}

	model.LoadOptions()

	// Todo: do this async
	model.LoadPrinterInfo()

	program := tea.NewProgram(model)
	return program
}

func (m *Model) LoadPrinterInfo() {
	response := m.Client.Call("printer.info", map[string]interface{}{})
	v, ok := response.Result.(map[string]interface{})
	if ok {
		m.host = v["hostname"].(string)
	} else {
		panic("")
	}

}

func (m *Model) LoadHelp() {
	response := m.Client.Call("printer.gcode.help", map[string]interface{}{})
	v, ok := response.Result.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("OHNOES %+v", response.Result))
	}
	for cmd, _ := range v {
		m.Input.TabComplete.RegisterCompletion(cmdinput.NewStringCompleter(cmd))
	}
}

func (m *Model) generateStyles() {

	vpBorder := m.Options.BorderType
	vpBorder.BottomRight = vpBorder.MiddleRight
	vpBorder.BottomLeft = vpBorder.MiddleLeft

	m.viewportBorderStyle = lipgloss.NewStyle().
		BorderStyle(vpBorder).
		BorderLeft(true).
		BorderRight(true).
		BorderForeground(m.Options.BorderForeground).
		BorderBackground(m.Options.BorderBackground).
		Padding(0, 1)

	inpBorder := m.Options.BorderType
	inpBorder.TopRight = vpBorder.MiddleRight
	inpBorder.TopLeft = vpBorder.MiddleLeft

	m.inputBorderStyle = lipgloss.NewStyle().
		BorderStyle(inpBorder).
		Padding(0, 1).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true).
		BorderForeground(m.Options.BorderForeground).
		BorderBackground(m.Options.BorderBackground)

	m.inlineBorderStyle = lipgloss.NewStyle().
		Foreground(m.viewportBorderStyle.GetBorderTopForeground()).
		Background(m.viewportBorderStyle.GetBorderTopBackground())

	m.inputTextStyle = lipgloss.NewStyle().
		// These are broken in the bubbles.textinput
		// They add an extra line under the input
		//Foreground(m.Options.InputForeground).
		//Background(m.Options.InputBackground).
		Inline(true)

}

func (m *Model) LoadOptions() {
	m.Options = &UIOptions{
		LogIncoming:      false,
		TimestampFormat:  "hh:mma",
		BorderType:       lipgloss.NormalBorder(),
		BorderForeground: lipgloss.Color("#00FF00"),
		InputForeground:  lipgloss.Color("#FFFFFF"),
		InputBackground:  lipgloss.Color("#0000FF"),
	}
	m.generateStyles()
}

func (m *Model) getLogTimestamp() string {
	return m.output.String(gostradamus.Now().Format(m.Options.TimestampFormat)).Foreground(m.output.Color("#666666")).String()
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
	for i, logEntry := range m.Content {
		if i > 0 {
			rendered = rendered + "\n"
		}
		rendered = rendered + m.renderEntry(logEntry)

	}
	m.renderedContent = strings.Trim(rendered, "\n")
	return m.renderedContent
}

func (m *Model) AppendLog(logEntry LogEntry) {
	if logEntry.timestamp == "" {
		logEntry.timestamp = m.getLogTimestamp()
	}
	first := len(m.Content) == 0
	m.Content = append(m.Content, logEntry)
	if !first {
		m.renderedContent = m.renderedContent + "\n"
	}
	m.renderedContent = m.renderedContent + strings.Trim(m.renderEntry(logEntry), "\n")
	m.Viewport.SetContent(m.renderedContent)
	m.Viewport.GotoBottom()
}

func (m *Model) handleIncoming(req jsonrpcclient.IncomingJsonRPCRequest) {
	switch req.Method {
	case "notify_gcode_response":
		for _, line := range req.Params {
			m.AppendLog(LogEntry{message: line.(string)})
		}
	default:
		if m.Options.LogIncoming {
			encoded, _ := json.Marshal(req)
			m.AppendLog(LogEntry{message: string(encoded)})
		}
	}
}

func (m *Model) cmdSet(rawArgs string) {
	args := strings.Fields(rawArgs)
	var val interface{}
	var err error
	switch strings.ToLower(args[0]) {
	case "borderforeground":
		val, err = parseColor(args[1])
		if err == nil {
			m.Options.BorderForeground = val.(lipgloss.Color)
		}
	case "borderbackground":
		val, err = parseColor(args[1])
		if err == nil {
			m.Options.BorderBackground = val.(lipgloss.Color)
		}
	case "inputforeground":
		val, err = parseColor(args[1])
		if err == nil {
			m.Options.InputForeground = val.(lipgloss.Color)
		}
	case "inputbackground":
		val, err = parseColor(args[1])
		if err == nil {
			m.Options.InputBackground = val.(lipgloss.Color)
		}
	case "bordertype":
		val, err = parseBorderType(args[1])
		if err == nil {
			m.Options.BorderType = val.(lipgloss.Border)
		}
	case "logincoming":
		val, err = parseBool(args[1])
		if err == nil {
			m.Options.LogIncoming = val.(bool)
		}
	case "timestampformat":
		val, err = parseTimestampFormat(args[1])
		if err == nil {
			m.Options.TimestampFormat = val.(string)
		}
	default:
		err = errors.New("Unknown Option: " + args[0])
	}
	if err != nil {
		m.AppendLog(LogEntry{message: err.Error()})
		return
	}
	m.generateStyles()
}

func (m *Model) cmdClear() {
	m.Content = []LogEntry{}
	m.renderedContent = ""
	m.Viewport.SetContent("")
}

func (m *Model) cmdRpc(rawArgs string) {
	parts := strings.SplitN(rawArgs, " ", 2)
	var params map[string]interface{}
	if len(parts) == 2 {
		json.Unmarshal([]byte(parts[1]), &params)
	} else {
		params = map[string]interface{}{}
	}
	result := m.Client.Call(parts[0], params)
	res, _ := json.MarshalIndent(result.Result, " ", " ")
	if result.Result != nil {
		m.AppendLog(LogEntry{timestamp: m.getLogTimestamp(), message: string(res)})
	}

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
	case "rpc":
		m.cmdRpc(rawArgs)
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

	// /rpc
	m.Input.TabComplete.RegisterCompletion(
		cmdinput.NewStringCompleter("/rpc"),
		cmdinput.NewListCompleter(MoonrakerRPCMethods...),
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
			m.Viewport = viewport.New(msg.Width-2, msg.Height-verticalMarginHeight)
			m.Viewport.YPosition = 0
			m.Viewport.SetContent(m.renderContent())

			m.TitleBar = NewTitleBar(m.host, m.version)
			m.TitleBar.BorderType = m.viewportBorderStyle.GetBorderStyle()
			m.TitleBar.BorderStyle = m.inlineBorderStyle

			m.Input = cmdinput.New()
			m.Input.SetTextStyle(m.inputTextStyle)
			m.Input.TextInput.Width = msg.Width - 4

			m.RegisterCompleters()
			m.LoadHelp()
			m.Ready = true
			cmds = append(cmds, m.readIncoming())

		} else {
			// just resize everything
			m.Viewport.Width = msg.Width - 4
			m.Viewport.Height = msg.Height - verticalMarginHeight
			m.Input.TextInput.Width = msg.Width - 4
			m.TitleBar.Width = msg.Width
			m.Viewport.SetContent(m.renderContent())
		}
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)
		m.TitleBar, cmd = m.TitleBar.Update(msg)
		cmds = append(cmds, cmd)
	default:
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.Input, cmd = m.Input.Update(msg)
		cmds = append(cmds, cmd)
		m.TitleBar, cmd = m.TitleBar.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}
	tb := m.TitleBar.View()
	vp := m.viewportBorderStyle.Width(m.Viewport.Width).Render(m.Viewport.View())
	// replace the top left corner with a title bar
	inp := m.inputBorderStyle.Width(m.Viewport.Width).Render(m.Input.View())
	return fmt.Sprintf("%s\n%s\n%s", tb, vp, inp)
}
