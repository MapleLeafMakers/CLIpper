package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type TitleBar struct {
	BorderStyle lipgloss.Style
	BorderType  lipgloss.Border

	HostStyle     lipgloss.Style
	AppTitleStyle lipgloss.Style
	VersionStyle  lipgloss.Style

	Host     string
	AppTitle string
	Version  string

	Width int
}

func NewTitleBar(host string, version string) TitleBar {
	return TitleBar{
		BorderType:    lipgloss.NormalBorder(),
		BorderStyle:   lipgloss.NewStyle(),
		HostStyle:     lipgloss.NewStyle(),
		AppTitleStyle: lipgloss.NewStyle(),
		VersionStyle:  lipgloss.NewStyle(),
		Host:          host,
		AppTitle:      "CLIpper",
		Version:       version,
	}
}

func (t TitleBar) Update(msg tea.Msg) (TitleBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.Width = msg.Width
	}
	return t, tea.Batch()
}

func (t TitleBar) View() string {
	host := t.HostStyle.Render(t.Host)
	appTitle := t.AppTitleStyle.Render(t.AppTitle)
	version := t.VersionStyle.Render(t.Version)
	left := t.BorderStyle.Render(t.BorderType.TopLeft+t.BorderType.Top+t.BorderType.MiddleRight) + " " + host + " " + t.BorderStyle.Render(t.BorderType.MiddleLeft)
	right := t.BorderStyle.Render(t.BorderType.MiddleRight) + " " + appTitle + " " + version + " " + t.BorderStyle.Render(t.BorderType.MiddleLeft+t.BorderType.Top+t.BorderType.TopRight)
	space := max(t.Width-lipgloss.Width(left)-lipgloss.Width(right), 0)
	return left + t.BorderStyle.Render(strings.Repeat(t.BorderType.Top, space)) + right

}
