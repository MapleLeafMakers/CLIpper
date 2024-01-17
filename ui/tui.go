package ui

import (
	"clipper/jsonrpcclient"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Padding(0, 1)
	}()
)

type Model struct {
	Content  string
	Ready    bool
	Viewport viewport.Model
	Input    textinput.Model
	Client   *jsonrpcclient.Client
}

func NewTUI(client *jsonrpcclient.Client) *tea.Program {
	model := Model{
		Client: client,
	}
	program := tea.NewProgram(model)
	return program
}

func (m Model) readIncoming() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.Client.Incoming
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) AppendLog(line string) {
	m.Content = m.Content + "\n" + line
	m.Viewport.SetContent(m.Content)
	m.Viewport.GotoBottom()
}

func (m *Model) handleIncoming(req jsonrpcclient.IncomingJsonRPCRequest) {
	switch req.Method {
	case "notify_gcode_response":
		for _, line := range req.Params {
			m.AppendLog(line.(string))
		}
	}
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
			m.Content = m.Content + "\n" + m.Input.Value()
			m.Viewport.SetContent(m.Content)
			m.Viewport.GotoBottom()
			go m.Client.Call("printer.gcode.script", map[string]interface{}{"script": m.Input.Value()})
			m.Input.SetValue("")
		} else {
			m.Input, cmd = m.Input.Update(msg)
			cmds = append(cmds, cmd)
		}
	case jsonrpcclient.IncomingJsonRPCRequest:
		m.handleIncoming(msg)
		cmds = append(cmds, m.readIncoming())

	case tea.WindowSizeMsg:

		verticalMarginHeight := 4
		m.Content = m.Content + "\n" + fmt.Sprintf("%dx%d", msg.Width, msg.Height)
		m.Viewport.SetContent(m.Content)
		m.Viewport.GotoBottom()
		if !m.Ready {
			m.Viewport = viewport.New(msg.Width-4, msg.Height-verticalMarginHeight)
			m.Viewport.YPosition = 0

			m.Viewport.SetContent(m.Content)

			m.Input = textinput.New()
			m.Input.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("end"))
			m.Input.Width = msg.Width - 4
			m.Input.ShowSuggestions = true
			m.Input.SetSuggestions([]string{"bed_mesh_calibrate", "bed_mesh_clear", "query_probe", "set_kinematic_position"})
			m.Input.Prompt = "> "
			m.Input.Focus()
			m.Ready = true
			cmds = append(cmds, m.readIncoming())

		} else {
			m.Viewport.Width = msg.Width - 4
			m.Viewport.Height = msg.Height - verticalMarginHeight
			m.Input.Width = msg.Width - 4
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

	b := lipgloss.RoundedBorder()
	b.BottomRight = "┤"
	b.BottomLeft = "├"
	vp := lipgloss.NewStyle().BorderStyle(b).Width(m.Viewport.Width).Padding(0, 1).Render(m.Viewport.View())
	inp := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Width(m.Viewport.Width).Padding(0, 1).BorderLeft(true).BorderRight(true).BorderBottom(true).Render(m.Input.View())
	return fmt.Sprintf("%s\n%s", vp, inp)
}
