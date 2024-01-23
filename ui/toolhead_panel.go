package ui

import (
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"strconv"
)

type ToolheadPanelContent struct {
	tui       *TUI
	container *tview.Form
	tabIndex  int
}

func NewToolheadPanel(tui *TUI) ToolheadPanelContent {
	content := ToolheadPanelContent{
		tui:       tui,
		container: tview.NewForm().SetItemPadding(0),
	}
	content.container.SetBorderPadding(0, 0, 1, 1)
	content.container.AddFormItem(buildAxis(tui, "X"))
	content.container.AddFormItem(buildAxis(tui, "Y"))
	content.container.AddFormItem(buildAxis(tui, "Z"))
	content.UpdatePositions()
	content.container.SetBorder(true).SetTitle("Toolhead").SetTitleColor(tcell.ColorLightYellow)
	return content
}

func (c ToolheadPanelContent) UpdatePositions() {
	for idx := 0; idx < 3; idx++ {
		inp := c.container.GetFormItem(idx).(*tview.InputField)
		prec := 2
		if idx == 2 {
			prec = 4
		}
		gcode_position := c.tui.State["gcode_move"]["gcode_position"].([]interface{})
		val := gcode_position[idx].(float64)
		inp.SetText(strconv.FormatFloat(val, 'f', prec, 64))
	}
}

func buildAxis(tui *TUI, axis string) tview.FormItem {
	//var color tcell.Color
	field := tview.NewInputField().
		SetLabel(axis + ": ")
	return field
}
