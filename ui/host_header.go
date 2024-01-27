package ui

import (
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
)

type HostHeaderContent struct {
	tview.TableContentReadOnly
	tui  *TUI
	cell *tview.TableCell
}

func (h HostHeaderContent) GetCell(row, column int) *tview.TableCell {
	var bgColor tcell.Color
	var text string

	if h.tui.RpcClient.IsConnected {
		text = h.tui.hostname
		bgColor = tcell.ColorGreen
	} else {
		text = "- Not Connected -"
		bgColor = tcell.ColorRed
	}

	return h.cell.SetBackgroundColor(bgColor).SetText(text)
}

func (h HostHeaderContent) GetRowCount() int { return 1 }

func (h HostHeaderContent) GetColumnCount() int { return 1 }

func NewHostHeader(tui *TUI) *tview.Table {
	table := tview.NewTable()
	content := HostHeaderContent{tui: tui, cell: tview.NewTableCell("").SetExpansion(1).SetAlign(tview.AlignCenter)}
	table.SetContent(content)
	return table
}
