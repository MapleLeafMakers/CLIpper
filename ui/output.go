package ui

import (
	"clipper/ui/cmdinput"
	"github.com/MapleLeafMakers/tview"
	"github.com/bykof/gostradamus"
	"github.com/gdamore/tcell/v2"
)

const (
	MsgTypeCommand  = iota
	MsgTypeResponse = iota
	MsgTypeInternal = iota
	MsgTypeError    = iota
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
	tui     *TUI
}

func (l *LogContent) Write(logEntry LogEntry) {
	l.entries = append(l.entries, logEntry)
}

func splitToLogEntries(message string, escape bool, base LogEntry) []LogEntry {
	if escape {
		message = cmdinput.Escape(message)
	}
	lines := cmdinput.WordWrap(message, 999) //arbitrary number to force it to only split on actual linebreaks for now
	entries := make([]LogEntry, 0, len(lines))
	for _, line := range lines {
		e := base
		e.Message = line
		entries = append(entries, e)
	}
	return entries
}

func (l *LogContent) WriteResponse(message string) {

	l.entries = append(l.entries, splitToLogEntries(message, true, LogEntry{MsgTypeResponse, gostradamus.Now(), ""})...)
	l.table.ScrollToEnd()

}

func (l *LogContent) WriteCommand(message string) {
	l.entries = append(l.entries, splitToLogEntries(message, true, LogEntry{MsgTypeCommand, gostradamus.Now(), ""})...)
	l.table.ScrollToEnd()
}

func (l *LogContent) WriteInternal(message string) {
	l.entries = append(l.entries, splitToLogEntries(message, false, LogEntry{MsgTypeInternal, gostradamus.Now(), ""})...)
	l.table.ScrollToEnd()
}

func (l *LogContent) WriteError(message string) {
	l.entries = append(l.entries, splitToLogEntries(message, true, LogEntry{MsgTypeError, gostradamus.Now(), ""})...)
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
		var textColor tcell.Color
		switch l.entries[row].Type {
		case MsgTypeCommand:
			textColor = AppConfig.Theme.ConsoleCommandColor.Color()
		case MsgTypeResponse, MsgTypeInternal:
			textColor = AppConfig.Theme.ConsoleResponseColor.Color()
		case MsgTypeError:
			textColor = AppConfig.Theme.ConsoleErrorColor.Color()
		}
		return tview.NewTableCell(" " + l.entries[row].Message).SetTextColor(textColor)
	}
	return nil
}

func (l LogContent) GetRowCount() int {
	return len(l.entries)
}

func (l LogContent) GetColumnCount() int {
	return 2
}
