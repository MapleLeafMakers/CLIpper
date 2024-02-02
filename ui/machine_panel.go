package ui

import (
	"fmt"
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/orsinium-labs/enum"
	"strconv"
)

type MachineLimitValue struct {
	DisplayName string
	StatusKey   string
	ParamName   string
	Unit        string
}
type MachineLimit enum.Member[MachineLimitValue]

var (
	b                    = enum.NewBuilder[MachineLimitValue, MachineLimit]()
	Velocity             = b.Add(MachineLimit{Value: MachineLimitValue{DisplayName: "Velocity", ParamName: "VELOCITY", StatusKey: "max_velocity", Unit: "mm/s"}})
	Acceleration         = b.Add(MachineLimit{Value: MachineLimitValue{DisplayName: "Acceleration", ParamName: "ACCEL", StatusKey: "max_accel", Unit: "mm/s²"}})
	SquareCornerVelocity = b.Add(MachineLimit{Value: MachineLimitValue{DisplayName: "SCV", ParamName: "SQUARE_CORNER_VELOCITY", StatusKey: "square_corner_velocity", Unit: "mm/s"}})
	AccelToDecel         = b.Add(MachineLimit{Value: MachineLimitValue{DisplayName: "Acce. to Decel", ParamName: "ACCEL_TO_DECEL", StatusKey: "max_accel_to_decel", Unit: "mm/s²"}})

	MachineLimits = b.Enum()
)

type MachineLimitPanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	table     *tview.Table
	container *tview.Flex
	tabIndex  int
}

func NewMachineLimitPanel(tui *TUI) *MachineLimitPanelContent {
	content := MachineLimitPanelContent{
		tui:       tui,
		table:     tview.NewTable(),
		container: tview.NewFlex().SetDirection(tview.FlexRow),
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
	content.container.SetBorder(true).SetTitle("Machine [L[]imits")
	return &content
}

func (t MachineLimitPanelContent) GetRowCount() int    { return MachineLimits.Len() }
func (t MachineLimitPanelContent) GetColumnCount() int { return 2 }

func (t MachineLimitPanelContent) GetCell(row int, column int) *tview.TableCell {
	limit := MachineLimits.Members()[row]
	switch column {
	case 0:
		return tview.NewTableCell(" " + limit.Value.DisplayName).SetMaxWidth(0).SetExpansion(2).SetSelectable(true).SetClickedFunc(func() bool {
			t.onSelected(row, column)
			return false
		})
	case 1:
		toolhead := t.tui.State["toolhead"]
		currentValue := toolhead[limit.Value.StatusKey]

		return tview.NewTableCell(fmt.Sprintf("%s %-5s", strconv.FormatFloat(currentValue.(float64), 'f', 0, 64), limit.Value.Unit)).SetMaxWidth(12).SetAlign(tview.AlignRight).SetExpansion(0).SetSelectable(true).SetClickedFunc(func() bool {
			t.onSelected(row, column)
			return false
		})
	}
	return nil
}

func (t MachineLimitPanelContent) onSelected(row int, column int) {
	limit := MachineLimits.Members()[row]
	go t.tui.promptForInput(fmt.Sprintf("New limit for %s > ", limit.Value.DisplayName), "", func(entered bool, value string) {
		if entered {
			newLimit, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				t.tui.Output.WriteError(err.Error())
			} else {
				t.tui.ExecuteGcode(fmt.Sprintf("SET_VELOCITY_LIMIT %s=%d", limit.Value.ParamName, newLimit))
			}
		}
	})
}
