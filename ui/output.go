package ui

import (
	"clipper/ui/cmdinput"
	"github.com/MapleLeafMakers/tview"
	"github.com/bykof/gostradamus"
)

const (
	MsgTypeCommand  = iota
	MsgTypeResponse = iota
	MsgTypeInternal = iota
)

type LogEntry struct {
	Type      int
	Timestamp gostradamus.DateTime
	Message   string
}

type LogContent struct {
	tview.TableContentReadOnly
	entries []LogEntry
	table   *tview.Table
}

func (l *LogContent) WriteResponse(message string) {
	lines := cmdinput.WordWrap(cmdinput.Escape(message), 999) //arbitrary number to force it to only split on actual linebreaks for now
	for _, line := range lines {
		l.entries = append(l.entries, LogEntry{MsgTypeResponse, gostradamus.Now(), line})
	}
	l.table.ScrollToEnd()
}

func (l *LogContent) WriteCommand(message string) {
	lines := cmdinput.WordWrap(cmdinput.Escape(message), 999) //arbitrary number to force it to only split on actual linebreaks for now
	for _, line := range lines {
		l.entries = append(l.entries, LogEntry{MsgTypeCommand, gostradamus.Now(), line})
	}
	l.table.ScrollToEnd()
}

func (l LogContent) GetCell(row, column int) *tview.TableCell {

	switch column {
	case 0:
		var txt string
		if l.entries[row].Timestamp.Time().IsZero() {
			txt = ""
		} else {
			txt = l.entries[row].Timestamp.Format(" " + AppConfig.TimestampFormat)
		}
		cell := tview.NewTableCell(txt).
			SetTextColor(AppConfig.Theme.ConsoleTimestampTextColor.Color())
		if txt != "" {
			cell.SetBackgroundColor(AppConfig.Theme.ConsoleTimestampBackgroundColor.Color())
		}
		return cell
	case 1:
		return tview.NewTableCell(" " + l.entries[row].Message).SetTextColor(AppConfig.Theme.ConsoleTextColor.Color())
	}
	return nil
}

func (l LogContent) GetRowCount() int {
	return len(l.entries)
}

func (l LogContent) GetColumnCount() int {
	return 2
}
