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
	content.container.SetBorder(true).SetTitle("Toolhead").SetTitleColor(tcell.ColorLightYellow)
	return content
}

func buildAxis(tui *TUI, axis string) tview.FormItem {
	//var color tcell.Color
	var idx int
	prec := 2
	switch axis {
	case "X":
		idx = 0
		//color = tcell.ColorRed
	case "Y":
		idx = 1
		//color = tcell.ColorGreen
	case "Z":
		prec = 3
		idx = 2
		//color = tcell.ColorBlue
	}
	gcode_position := tui.State["gcode_move"]["gcode_position"].([]interface{})
	val := gcode_position[idx].(float64)
	field := tview.NewInputField().
		SetLabel(axis + ": ").
		SetText(strconv.FormatFloat(val, 'f', prec, 64))
	return field
}
