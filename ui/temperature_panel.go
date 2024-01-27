package ui

import (
	"github.com/MapleLeafMakers/tview"
	"sort"
	"strconv"
	"strings"
)

type TempSensor struct {
	StatusKey   string
	DisplayName string
	Type        string
	Temperature float64
	Target      float64
}

type TemperaturePanelContent struct {
	tview.TableContentReadOnly
	tui       *TUI
	sensors   []TempSensor
	table     *tview.Table
	container *tview.Flex
}

func (t TemperaturePanelContent) GetCell(row, column int) *tview.TableCell {
	// TODO: this could be way more efficient by not recreating new cells every time.
	switch column {
	case 0:
		icon := getIcon(t.sensors[row].Type)
		return tview.NewTableCell(" " + icon)
	case 1:
		return tview.NewTableCell(t.sensors[row].DisplayName).SetExpansion(1)
	case 2:
		s := t.tui.State[t.sensors[row].StatusKey]
		sensor := s
		return tview.NewTableCell(strconv.FormatFloat(sensor["temperature"].(float64), 'f', 1, 64) + "°C").SetAlign(tview.AlignRight)
	case 3:
		s := t.tui.State[t.sensors[row].StatusKey]
		sensor := s
		target, ok := sensor["target"]
		if ok && target.(float64) > 0 {
			activity, okActivity := sensor["power"].(float64)
			activityIcon := ""
			if okActivity {
				activityIcon = getHeaterActivityIcon(activity)
			}
			return tview.NewTableCell(strconv.Itoa(int(target.(float64))) + " " + activityIcon).SetAlign(tview.AlignRight)
		} else {
			return tview.NewTableCell("")
		}
	}
	return nil
}

func getHeaterActivityIcon(activity float64) string {
	levels := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	index := int(activity * float64(len(levels)-1))
	return "[red]" + string(levels[index])
}

func (t TemperaturePanelContent) GetRowCount() int {
	return len(t.sensors)
}

func (t TemperaturePanelContent) GetColumnCount() int {
	return 4
}

func NewTemperaturePanel(tui *TUI) TemperaturePanelContent {
	sensors := getSensors(tui.State["heaters"])
	content := TemperaturePanelContent{
		tui:       tui,
		sensors:   sensors,
		container: tview.NewFlex().SetDirection(tview.FlexRow),
	}
	table := tview.NewTable()

	content.table = table
	table.SetContent(content)
	content.container.AddItem(table, 0, 1, false)
	content.container.SetBorder(true).SetTitle("[T[]emperatures")
	return content
}

func getSensors(state map[string]interface{}) []TempSensor {
	sensors_ := state["available_sensors"].([]interface{})
	sensors := make([]string, len(sensors_))
	for i, s := range sensors_ {
		sensors[i] = s.(string)
	}
	sort.Strings(sensors)
	results := make([]TempSensor, len(sensors))
	for i, sensorKey := range sensors {
		keyParts := strings.Split(sensorKey, " ")
		var sType, sName string
		if len(keyParts) == 1 {
			sType = keyParts[0]
			sName = toDisplayName(keyParts[0])
		} else {
			sType = keyParts[0]
			sName = toDisplayName(keyParts[1])
		}

		results[i] = TempSensor{
			StatusKey:   sensorKey,
			DisplayName: sName,
			Type:        sType,
		}
	}
	return results
}

func toDisplayName(key string) string {
	words := strings.Split(key, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}

func getIcon(sType string) string {
	switch sType {
	case "heater_bed":
		return "🛏"
	case "heater_generic":
		return "♨"
	case "extruder":
		return "⛊"
	default:
		return "🌡"
	}
}
