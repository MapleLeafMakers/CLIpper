package ui

import (
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"log"
	"strconv"
)

type ToolheadPanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	table     *tview.Table
	container *tview.Flex
	tabIndex  int
	cells     [][]*tview.TableCell
}

func NewToolheadPanel(tui *TUI) ToolheadPanelContent {
	content := ToolheadPanelContent{
		tui:       tui,
		table:     tview.NewTable(),
		container: tview.NewFlex().SetDirection(tview.FlexRow),
		cells:     make([][]*tview.TableCell, 0, 3),
	}

	content.cells = append(content.cells, buildAxis("X", tcell.GetColor("red")))
	content.cells = append(content.cells, buildAxis("Y", tcell.GetColor("green")))
	content.cells = append(content.cells, buildAxis("Z", tcell.GetColor("blue")))
	content.table.SetContent(content)
	content.container.AddItem(content.table, 0, 1, false)
	content.container.SetBorder(true).SetTitle("M[o[]tion")
	spdInput := tview.NewInputField().SetLabel(" Speed: ")
	content.container.AddItem(spdInput, 1, 0, false)
	return content
}

func (t ToolheadPanelContent) GetRowCount() int    { return 3 }
func (t ToolheadPanelContent) GetColumnCount() int { return 3 }
func (t ToolheadPanelContent) GetCell(row, column int) *tview.TableCell {
	cell := t.cells[row][column]
	switch column {
	case 1:
		gcode_position, ok := t.tui.State["gcode_move"]["gcode_position"].([]interface{})
		if ok && len(gcode_position) >= 3 {
			prec := 2
			if row == 2 {
				prec = 3
			}
			cell.SetText(strconv.FormatFloat(gcode_position[row].(float64), 'f', prec, 64))
		}
	case 2:
		if row == 2 {
			homing_origin, ok := t.tui.State["gcode_move"]["homing_origin"].([]interface{})
			if ok && len(homing_origin) >= 3 {
				cell.SetText("[" + strconv.FormatFloat(homing_origin[row].(float64), 'f', 3, 64) + "]")
			}
		}
	}
	return cell
}

func (t ToolheadPanelContent) moveToolhead(axis string, dist float64) {
	log.Println("Moving Toolhead: ", axis, ":", dist)
}

func buildAxis(axis string, color tcell.Color) []*tview.TableCell {
	cells := make([]*tview.TableCell, 0, 3)
	cells = append(cells, tview.NewTableCell(" "+axis+":").SetTextColor(color))
	cells = append(cells, tview.NewTableCell("?"))
	cells = append(cells, tview.NewTableCell("").SetTextColor(AppConfig.Theme.TertiaryTextColor.Color()).SetAlign(tview.AlignRight).SetExpansion(1))
	return cells
}

func buildSpeedSlider(tui *TUI) {

}
