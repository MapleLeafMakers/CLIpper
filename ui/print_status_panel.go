package ui

import (
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"time"
)

var stateIcons map[string]string = map[string]string{
	"complete": "✓",
	"paused":   "⏸",
	"printing": "▶",
	"error":    "⚠",
	"standby":  "…",
}

type PrintStatusPanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	table     *tview.Table
	container *tview.Flex
	tabIndex  int
}

func NewPrintStatusPanel(tui *TUI) *PrintStatusPanelContent {
	content := PrintStatusPanelContent{
		tui:       tui,
		table:     tview.NewTable(),
		container: tview.NewFlex().SetDirection(tview.FlexRow),
	}

	content.table.SetContent(content)
	content.container.AddItem(content.table, 0, 1, false)
	content.container.SetBorder(true).SetTitle("[P[]rint")
	return &content
}

func (t PrintStatusPanelContent) GetRowCount() int    { return 2 }
func (t PrintStatusPanelContent) GetColumnCount() int { return 1 }
func (t PrintStatusPanelContent) GetCell(row, column int) *tview.TableCell {
	switch row {
	case 0:
		fname := t.getFilename()
		state := t.getPrintState()
		icon := stateIcons[state]
		msg := ""
		if state != "standby" {
			msg = icon + " " + fname
		}
		return tview.NewTableCell(msg).SetAlign(tview.AlignCenter)
	case 1:
		msg, _ := t.tui.State["display_status"]["message"].(string)

		w := tview.TaggedStringWidth(msg)
		if w-30 > 0 {
			msgOffset := time.Now().Unix() % int64(w-30)
			if msgOffset > 0 {
				msg = "…" + msg[msgOffset:]
			} else {
				msg = msg[msgOffset:]
			}

		}

		if msg != "" {
			msg = "ⓘ " + msg
		}

		return tview.NewTableCell(msg).SetStyle(tcell.StyleDefault.Italic(true))
	}
	return nil
}

func (t PrintStatusPanelContent) getPrintState() string {
	print_stats, ok := t.tui.State["print_stats"]
	if ok {
		state, ok := print_stats["state"].(string)
		if ok {
			return state
		}
	}
	return ""
}

func (t PrintStatusPanelContent) getFilename() string {
	print_stats, ok := t.tui.State["print_stats"]
	if ok {
		fname, ok := print_stats["filename"].(string)
		if ok {
			return fname
		}
	}
	return ""
}
