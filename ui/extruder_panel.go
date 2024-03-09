package ui

import (
	"fmt"
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/orsinium-labs/enum"
	"strconv"
)

type ExtruderParameterValue struct {
	DisplayName  string
	StatusObject string
	StatusKey    string
	CommandName  string
	ParamName    string
	Unit         string
}
type ExtruderParameter enum.Member[ExtruderParameterValue]

var (
	extruderParametersBuilder = enum.NewBuilder[ExtruderParameterValue, ExtruderParameter]()
	ExtrusionFactor           = extruderParametersBuilder.Add(ExtruderParameter{Value: ExtruderParameterValue{DisplayName: "Extrusion Factor", CommandName: "M221", ParamName: "S", StatusObject: "gcode_move", StatusKey: "extrude_factor", Unit: "%"}})
	PressureAdvance           = extruderParametersBuilder.Add(ExtruderParameter{Value: ExtruderParameterValue{DisplayName: "Pressure Advance", CommandName: "SET_PRESSURE_ADVANCE", ParamName: "ADVANCE", StatusObject: "extruder", StatusKey: "pressure_advance", Unit: "mm/s"}})
	SmoothTime                = extruderParametersBuilder.Add(ExtruderParameter{Value: ExtruderParameterValue{DisplayName: "Smooth Time", CommandName: "SET_PRESSURE_ADVANCE", ParamName: "SMOOTH_TIME", StatusObject: "extruder", StatusKey: "smooth_time", Unit: "s"}})
	FilamentLength            = extruderParametersBuilder.Add(ExtruderParameter{Value: ExtruderParameterValue{DisplayName: "Filament Length", StatusKey: "filamentLength", Unit: "mm"}})
	ExtrusionFeedRate         = extruderParametersBuilder.Add(ExtruderParameter{Value: ExtruderParameterValue{DisplayName: "ExtrusionFeedRate", StatusKey: "extrusionFeedRate", Unit: "mm/s"}})
	ExtruderParameters        = extruderParametersBuilder.Enum()
)

type ExtruderPanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	table     *tview.Table
	container *tview.Flex
	tabIndex  float64
	settings  map[string]float64
}

func NewExtruderPanel(tui *TUI) *ExtruderPanelContent {
	content := ExtruderPanelContent{
		tui:       tui,
		table:     tview.NewTable(),
		container: tview.NewFlex().SetDirection(tview.FlexRow),
		settings: map[string]float64{
			"filamentLength":    50,
			"extrusionFeedRate": 5,
		},
	}

	table := tview.NewTable().SetSelectable(false, false)
	table.SetSelectedStyle(tcell.Style{})

	table.SetFocusFunc(func() {
		table.Select(0, 0)
		table.SetSelectable(true, false)

	})
	table.SetBlurFunc(func() {
		table.SetSelectable(false, false)
	})
	table.SetSelectedFunc(content.onSelected)
	content.table = table
	table.SetContent(content)
	content.container.AddItem(table, 0, 1, true)
	content.container.SetBorder(true).SetTitle("[E[]xtruder")
	return &content
}

func (t ExtruderPanelContent) GetRowCount() int    { return ExtruderParameters.Len() + 1 }
func (t ExtruderPanelContent) GetColumnCount() int { return 2 }

func (t ExtruderPanelContent) GetCell(row int, column int) *tview.TableCell {
	// Retract & Extrude buttons
	if row == len(ExtruderParameters.Members()) {
		if column == 0 {
			return tview.NewTableCell(" ⬆ Retract").SetMaxWidth(0).SetExpansion(2).SetSelectable(true).SetClickedFunc(func() bool {
				t.onSelected(row, column)
				return false
			})
		} else {
			return tview.NewTableCell(" ⬇ Extrude").SetMaxWidth(0).SetExpansion(2).SetSelectable(true).SetClickedFunc(func() bool {
				t.onSelected(row, column)
				return false
			})
		}
	}
	param := ExtruderParameters.Members()[row]
	switch column {
	case 0:
		return tview.NewTableCell(" " + param.Value.DisplayName).SetMaxWidth(0).SetExpansion(2).SetSelectable(true).SetClickedFunc(func() bool {
			t.onSelected(row, column)
			return false
		})
	case 1:
		switch param {
		case FilamentLength, ExtrusionFeedRate:
			if currentValue, ok := t.settings[param.Value.StatusKey]; ok {
				return tview.NewTableCell(
					fmt.Sprintf("%s %-5s", strconv.FormatInt(int64(currentValue), 10), param.Value.Unit)).SetMaxWidth(12).SetAlign(tview.AlignRight).SetExpansion(0).SetSelectable(true).SetClickedFunc(func() bool {
					t.onSelected(row, column)
					return false
				})
			}
		case PressureAdvance, SmoothTime:
			if statusObject, ok := t.tui.State[param.Value.StatusObject]; ok {
				if currentValue, ok := statusObject[param.Value.StatusKey]; ok {
					return tview.NewTableCell(
						fmt.Sprintf("%s %-5s", strconv.FormatFloat(currentValue.(float64), 'f', 3, 64), param.Value.Unit)).SetMaxWidth(12).SetAlign(tview.AlignRight).SetExpansion(0).SetSelectable(true).SetClickedFunc(func() bool {
						t.onSelected(row, column)
						return false
					})
				}
			}
		case ExtrusionFactor:
			if statusObject, ok := t.tui.State[param.Value.StatusObject]; ok {
				if currentValue, ok := statusObject[param.Value.StatusKey]; ok {
					return tview.NewTableCell(
						fmt.Sprintf("%s %-5s", strconv.FormatFloat(currentValue.(float64)*100, 'f', 0, 64), param.Value.Unit)).SetMaxWidth(12).SetAlign(tview.AlignRight).SetExpansion(0).SetSelectable(true).SetClickedFunc(func() bool {
						t.onSelected(row, column)
						return false
					})
				}
			}
		}
	}
	return nil
}

func (t ExtruderPanelContent) onSelected(row int, column int) {
	// Retract & Extrude buttons
	if row == len(ExtruderParameters.Members()) {
		dist := t.settings[FilamentLength.Value.StatusKey]
		speed := t.settings[ExtrusionFeedRate.Value.StatusKey] * 60

		if column == 0 {
			dist *= -1
		}

		t.tui.ExecuteGcode("M83")
		t.tui.ExecuteGcode(fmt.Sprintf("G92 E%d F%d", int64(dist), int64(speed)))
	} else {
		param := ExtruderParameters.Members()[row]
		go t.tui.promptForInput(fmt.Sprintf("New param for %s > ", param.Value.DisplayName), "", func(entered bool, value string) {
			if entered {
				switch param {
				case FilamentLength, ExtrusionFeedRate:
					newValue, err := strconv.ParseInt(value, 10, 32)
					if err != nil {
						t.tui.Output.WriteError(err.Error())
					} else {
						t.settings[param.Value.StatusKey] = float64(newValue)
					}
				case PressureAdvance, SmoothTime:
					newValue, err := strconv.ParseFloat(value, 32)
					if err != nil {
						t.tui.Output.WriteError(err.Error())
					} else {
						t.tui.ExecuteGcode(fmt.Sprintf("%s %s=%s", param.Value.CommandName, param.Value.ParamName, strconv.FormatFloat(newValue, 'f', 3, 64)))
					}
				default:
					newValue, err := strconv.ParseInt(value, 10, 32)
					if err != nil {
						t.tui.Output.WriteError(err.Error())
					} else {
						t.tui.ExecuteGcode(fmt.Sprintf("%s %s%d", param.Value.CommandName, param.Value.ParamName, newValue))
					}
				}
			}
		})
	}
}
