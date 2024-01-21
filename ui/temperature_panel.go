package ui

import (
	"github.com/rivo/tview"
	"strconv"
)

type TemperaturePanelContent struct {
	tview.TableContentReadOnly
	tui     *TUI
	sensors []string
	table   *tview.Table
}

func (t TemperaturePanelContent) GetCell(row, column int) *tview.TableCell {
	switch column {
	case 0:
		return tview.NewTableCell(t.sensors[row])
	case 1:
		s := t.tui.State[t.sensors[row]]
		sensor, _ := s.(map[string]interface{})
		strconv.FormatFloat(sensor["temperature"].(float64), 'f', 1, 64)
		return tview.NewTableCell(strconv.FormatFloat(sensor["temperature"].(float64), 'f', 2, 64))
	case 2:
		return nil
	}
	return nil
}

func (t TemperaturePanelContent) GetRowCount() int {
	return len(t.sensors)
}

func (t TemperaturePanelContent) GetColumnCount() int {
	return 3
}

func NewTemperaturePanel(tui *TUI) {
	content := TemperaturePanelContent{}
	table := tview.NewTable()
	table.SetContent(content)
}
