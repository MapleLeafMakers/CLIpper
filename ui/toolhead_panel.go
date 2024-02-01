package ui

import (
	"fmt"
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"log"
	"math"
	"strconv"
	"strings"
)

type ToolheadPanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	table     *tview.Table
	container *tview.Flex
	tabIndex  int
	cells     [][]*tview.TableCell
}

func NewToolheadPanel(tui *TUI) *ToolheadPanelContent {
	content := ToolheadPanelContent{
		tui:       tui,
		table:     tview.NewTable(),
		container: tview.NewFlex().SetDirection(tview.FlexRow),
		cells:     make([][]*tview.TableCell, 0, 4),
	}

	content.cells = append(content.cells, content.buildAxis("X", tcell.GetColor("red")))
	content.cells = append(content.cells, content.buildAxis("Y", tcell.GetColor("green")))
	content.cells = append(content.cells, content.buildAxis("Z", tcell.GetColor("blue")))
	content.cells = append(content.cells, content.buildMiscRow())
	content.table.SetContent(content).
		SetSelectable(false, false).
		SetSelectedFunc(content.onSelected).
		SetFocusFunc(func() {
			content.table.Select(0, 1)
			content.table.SetSelectable(true, true)
		}).SetBlurFunc(func() {
		content.table.SetSelectable(false, false)
	})
	content.container.AddItem(content.table, 0, 1, true)
	content.container.SetBorder(true).SetTitle("M[o[]tion")
	return &content
}

func (t ToolheadPanelContent) GetRowCount() int    { return 4 }
func (t ToolheadPanelContent) GetColumnCount() int { return 5 }
func (t ToolheadPanelContent) GetCell(row, column int) *tview.TableCell {
	cell := t.cells[row][column]
	if row < 3 {
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
		case 3:
			if row == 2 {
				homing_origin, ok := t.tui.State["gcode_move"]["homing_origin"].([]interface{})
				if ok && len(homing_origin) >= 3 {
					cell.SetText("[" + strconv.FormatFloat(homing_origin[row].(float64), 'f', 3, 64) + "]")
				}
			}
		}
	} else if row == 3 {
		switch column {
		case 1:
			speed_factor, ok := t.tui.State["gcode_move"]["speed_factor"].(float64)
			if ok {
				cell.SetText(strconv.Itoa(int(math.Trunc(speed_factor*100))) + "%")
			}
		}
	}

	return cell
}

func (t ToolheadPanelContent) moveToolhead(axis string, dist float64) {
	log.Println("Moving Toolhead: ", axis, ":", dist)
}

func (t ToolheadPanelContent) onSelected(row int, column int) {
	if row < 3 {
		axis := "XYZ"[row : row+1]
		switch column {
		case 1:
			t.doMove(axis)
		case 3:
			t.setZOffset()
		case 4:
			t.homeAxis(axis)
		}
	} else if row == 3 && column == 4 {
		// Home All
		t.homeAll()
	} else if row == 3 && column == 1 {
		t.setSpeed()
	}

}

func (t *ToolheadPanelContent) homeAxis(axis string) {
	go t.tui.ExecuteGcode("G28 " + axis)
}

func (t *ToolheadPanelContent) setZOffset() {
	homing_origin, ok := t.tui.State["gcode_move"]["homing_origin"].([]interface{})
	if ok && len(homing_origin) >= 3 {
		zOff := strconv.FormatFloat(homing_origin[2].(float64), 'f', 3, 64)
		go t.tui.promptForInput("Z-Offset> ", zOff, func(entered bool, value string) {
			if entered {
				zOffset, err := strconv.ParseFloat(value, 64)
				if err != nil {
					t.tui.Output.WriteError(err.Error())
				} else {
					t.tui.ExecuteGcode(fmt.Sprintf("SET_GCODE_OFFSET Z=%.3f", zOffset))
				}
			}
		})
	}
}

func (t ToolheadPanelContent) setSpeed() {
	speed_factor, ok := t.tui.State["gcode_move"]["speed_factor"].(float64)
	if ok {
		spd := strconv.Itoa(int(math.Trunc(speed_factor * 100)))
		go t.tui.promptForInput("Speed Factor> ", spd, func(entered bool, value string) {
			if entered {
				spd, err := strconv.Atoi(value)
				if err != nil {
					t.tui.Output.WriteError(err.Error())
				} else {
					t.tui.ExecuteGcode(fmt.Sprintf("M220 S%d", spd))
				}
			}
		})
	}
}

func (t *ToolheadPanelContent) doMove(axis string) {
	axisIdx := strings.Index("XYZ", axis)
	position, ok := t.tui.State["gcode_move"]["gcode_position"].([]interface{})
	if ok && len(position) >= 3 {
		pos := position[axisIdx].(float64)
		feedRate := int(t.tui.State["gcode_move"]["speed"].(float64))
		go t.tui.promptForInput(fmt.Sprintf("New pos. for axis %s> ", axis), strconv.FormatFloat(pos, 'f', -1, 64), func(entered bool, value string) {
			if entered {
				t.tui.ExecuteGcode(fmt.Sprintf("G90\nG1 %s%s F%d", axis, value, feedRate))
			}
		})
	}
}

func (t ToolheadPanelContent) homeAll() {
	go t.tui.ExecuteGcode("G28")
}

func (t *ToolheadPanelContent) buildAxis(axis string, color tcell.Color) []*tview.TableCell {
	axisIdx := strings.Index("XYZ", axis)
	cells := make([]*tview.TableCell, 0, 5)
	cells = append(cells, tview.NewTableCell("   "+axis+":").SetTextColor(color).SetSelectable(false))
	cells = append(cells, tview.NewTableCell("?").SetTextColor(AppConfig.Theme.PrimaryTextColor.Color()).SetClickedFunc(func() bool {
		t.onSelected(axisIdx, 1)
		return false
	}))
	cells = append(cells, tview.NewTableCell("").SetExpansion(1).SetSelectable(false))
	cells = append(cells, tview.NewTableCell("").SetTextColor(AppConfig.Theme.TertiaryTextColor.Color()).SetAlign(tview.AlignRight).SetSelectable(axis == "Z").SetClickedFunc(func() bool {
		if axis == "Z" {
			t.onSelected(axisIdx, 3)
		}
		return false
	}))
	cells = append(cells, tview.NewTableCell("🏠").SetClickedFunc(func() bool {
		t.onSelected(axisIdx, 4)
		return false
	}))
	return cells
}

func (t *ToolheadPanelContent) buildMiscRow() []*tview.TableCell {
	cells := make([]*tview.TableCell, 0, 5)
	cells = append(cells, tview.NewTableCell(" Spd:").SetSelectable(false))
	cells = append(cells, tview.NewTableCell("100%").SetTextColor(AppConfig.Theme.PrimaryTextColor.Color()).SetClickedFunc(func() bool {
		t.setSpeed()
		return false
	}))
	cells = append(cells, tview.NewTableCell("").SetExpansion(1).SetSelectable(false))
	cells = append(cells, tview.NewTableCell("").SetSelectable(false))
	cells = append(cells, tview.NewTableCell("🏠 All"))
	return cells
}
