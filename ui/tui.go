package ui

import (
	"clipper/jsonrpcclient"
	"clipper/ui/cmdinput"
	"encoding/json"
	"errors"
	"github.com/MapleLeafMakers/tview"
	"github.com/bykof/gostradamus"
	"github.com/gdamore/tcell/v2"
	"log"
	"reflect"
	"sync"
)

type LogEntry struct {
	Timestamp gostradamus.DateTime
	Message   string
}

type LogContent struct {
	tview.TableContentReadOnly
	entries []LogEntry
	table   *tview.Table
}

func (l *LogContent) Write(message string) {
	lines := cmdinput.WordWrap(cmdinput.Escape(message), 999) //arbitrary number to force it to only split on actual linebreaks for now
	for _, line := range lines {
		l.entries = append(l.entries, LogEntry{gostradamus.Now(), line})
	}
	l.table.ScrollToEnd()
}

func (l LogContent) GetCell(row, column int) *tview.TableCell {

	switch column {
	case 0:
		return tview.NewTableCell(l.entries[row].Timestamp.Format(" hh:mma")).SetBackgroundColor(tcell.NewRGBColor(64, 64, 64))
	case 1:
		return tview.NewTableCell(" " + l.entries[row].Message)
	}
	return nil
}

func (l LogContent) GetRowCount() int {
	return len(l.entries)
}

func (l LogContent) GetColumnCount() int {
	return 2
}

type TUI struct {
	App          *tview.Application
	Root         *tview.Grid
	Input        *cmdinput.InputField
	Output       *LogContent
	RpcClient    *jsonrpcclient.Client
	TabCompleter cmdinput.TabCompleter
	State        map[string]map[string]interface{}
	Settings     Settings
	HostHeader   *tview.TextView

	hostname string
}

func NewTUI(rpcClient *jsonrpcclient.Client) *TUI {
	tui := &TUI{
		RpcClient: rpcClient,
	}

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tui.Settings = AppSettings
	tui.buildInput()
	tui.buildOutput(100)
	tui.buildWindow()
	tui.App = tview.NewApplication().SetRoot(tui.Root, true).EnableMouse(true)

	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyPgUp, tcell.KeyPgDn:
			tui.Output.table.InputHandler()(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyTab:
			focused := reflect.TypeOf(tui.App.GetFocus())
			tui.Output.Write("Focused: " + focused.String())
			return event
		default:
			return event
		}
		return nil // never happens
	})

	go tui.initialize()
	return tui
}

func (tui *TUI) buildInput() {
	//tui.Input = tview.NewInputField()
	tui.Input = cmdinput.NewInputField().SetPlaceholder("Enter GCODE Commands or / commands").SetLabel("> ").SetLabelStyle(tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite).Bold(true))
	tui.TabCompleter = cmdinput.NewTabCompleter()
	tui.TabCompleter.RegisterCommand("/set", Command_Set{})
	tui.TabCompleter.RegisterCommand("/settings", Command_Settings{})
	tui.TabCompleter.RegisterCommand("/quit", Command_Quit{})
	tui.TabCompleter.RegisterCommand("/rpc", Command_RPC{})
	tui.TabCompleter.RegisterCommand("/restart", Command_Restart{})
	tui.TabCompleter.RegisterCommand("/firmware_restart", Command_FirmwareRestart{})
	tui.TabCompleter.RegisterCommand("/estop", Command_EStop{})
	tui.TabCompleter.RegisterCommand("/print", Command_Print{})

	tui.Input.SetAutocompleteFunc(func(currentText string, cursorPos int) (entries []string, menuOffset int) {
		ctx := cmdinput.CommandContext{
			"tui": tui,
			"raw": currentText,
		}
		return tui.TabCompleter.AutoComplete(currentText, cursorPos, ctx)
	})

	tui.Input.SetAutocompletedFunc(func(text string, index, source int) bool {
		closeMenu, fullText, cursorPos := tui.TabCompleter.OnAutoCompleted(text, index, source)
		tui.Input.SetText(fullText)
		tui.Input.SetCursor(cursorPos)
		return closeMenu
	})

	tui.Input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			ctx := cmdinput.CommandContext{
				"tui": tui,
				"raw": tui.Input.GetText(),
			}
			err := tui.TabCompleter.Parse(tui.Input.GetText(), ctx)
			if err == nil {
				log.Println("Executing", ctx)
				cmd, ok := ctx["cmd"]
				// ew
				if ok {
					cmd2, ok2 := cmd.(cmdinput.Command)
					if ok2 {
						go cmd2.Call(ctx)
					}
				} else {
					//not a registered command, send it as gcode.
					go (func() { NewGcodeCommand("", "").Call(ctx) })()
				}
				tui.Input.Clear()
			} else if err.Error() == "NoInput" {

			} else {
				panic(err)
			}
		default:
		}
	})
}

func (tui *TUI) buildLeftPanel() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	tui.HostHeader = tview.NewTextView().SetTextAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow).SetText(tui.hostname)
	tui.HostHeader.SetBackgroundColor(tcell.ColorDarkCyan)
	tui.Root.AddItem(flex, 0, 0, 1, 1, 0, 0, true)
	flex.AddItem(tui.HostHeader, 1, 1, false)
	tempPanel := NewTemperaturePanel(tui)
	flex.AddItem(tempPanel.container, len(tempPanel.sensors)+2, 0, false)
	//flex.SetFocusFunc(func() { log.Println("leftPanel focused") })
	toolheadPanel := NewToolheadPanel(tui)
	//toolheadPanel.container.SetFocusFunc(func() {
	//	flex.Focus(nil)
	//	log.Println("thp focused")
	//})
	flex.AddItem(toolheadPanel.container, 5, 0, true)
}

func (tui *TUI) initialize() {
	wg := new(sync.WaitGroup)
	go tui.handleIncoming()
	wg.Add(2)
	go tui.loadPrinterInfo(wg)
	go tui.subscribe(wg)
	wg.Wait()
	tui.App.QueueUpdateDraw(func() {
		tui.buildLeftPanel()
	})
	go tui.loadGcodeHelp()

}
func (tui *TUI) buildOutput(numLines int) {

	output := tview.NewTable()
	ts := gostradamus.Now()
	lines := make([]LogEntry, numLines)
	i := 0
	for i = 0; i < numLines-6; i++ {
		lines[i] = LogEntry{ts, ""}
	}
	lines[i+0] = LogEntry{ts, "[yellow]   ________    ____                     "}
	lines[i+1] = LogEntry{ts, "[yellow]  / ____/ /   /  _/___  ____  ___  _____"}
	lines[i+2] = LogEntry{ts, "[yellow] / /   / /    / // __ \\/ __ \\/ _ \\/ ___/"}
	lines[i+3] = LogEntry{ts, "[yellow]/ /___/ /____/ // /_/ / /_/ /  __/ /    "}
	lines[i+4] = LogEntry{ts, "[yellow]\\____/_____/___/ .___/ .___/\\___/_/     "}
	lines[i+5] = LogEntry{ts, "[yellow]              /_/   /_/                 "}

	tui.Output = &LogContent{
		table:   output,
		entries: lines,
	}
	output.SetContent(tui.Output)
	output.ScrollToEnd()
}

func (tui *TUI) buildWindow() {
	tui.Root = tview.NewGrid().
		SetRows(0, 1).
		SetColumns(36, 0).
		SetBorders(true).
		AddItem(tui.Input, 1, 0, 1, 2, 0, 0, true).
		AddItem(tui.Output.table, 0, 1, 1, 1, 0, 0, false)

}

func getObjectList(client *jsonrpcclient.Client) []string {
	resp, err := client.Call("printer.objects.list", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	objects, _ := resp.(map[string]interface{})
	objs := objects["objects"]
	objList, _ := objs.([]interface{})
	res := make([]string, 0, len(objList))
	for _, o := range objList {
		res = append(res, o.(string))
	}
	return res
}

func (tui *TUI) subscribe(wg *sync.WaitGroup) {
	defer wg.Done()
	objList := getObjectList(tui.RpcClient)
	subs := make(map[string]interface{}, len(objList))
	for _, key := range objList {
		subs[key] = nil
	}
	resp, err := tui.RpcClient.Call("printer.objects.subscribe", map[string]interface{}{
		"objects": subs,
	})
	if err != nil {
		panic(err)
	}
	asMap, _ := resp.(map[string]interface{})
	objectsAsMap, _ := asMap["status"].(map[string]interface{})
	state := make(map[string]map[string]interface{}, len(objectsAsMap))
	for k, v := range objectsAsMap {
		state[k] = v.(map[string]interface{})
	}
	log.Println("Queuing update for subs")
	tui.App.QueueUpdateDraw(func() {
		tui.State = state
		//log.Println("Subbed", state)
	})
}

func (tui *TUI) UpdateState(statusChanges map[string]map[string]interface{}) {
	for key, objectStatus := range statusChanges {
		for subKey, value := range objectStatus {
			tui.State[key][subKey] = value
			// TODO: Notify the relevant UI elements?
		}
	}
}

func toStatusMap(stat map[string]interface{}) (map[string]map[string]interface{}, error) {
	statusMap := make(map[string]map[string]interface{}, len(stat))
	for k, v := range stat {
		assertedV, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.New("NotAStatusMap")
		}
		statusMap[k] = assertedV
	}
	return statusMap, nil
}

func (tui *TUI) handleIncoming() {

	for {
		logIncoming, _ := AppSettings["logIncoming"].(bool)
		incoming := <-tui.RpcClient.Incoming
		switch incoming.Method {
		case "notify_status_update":
			status := incoming.Params[0].(map[string]interface{})
			statusMap, _ := toStatusMap(status)
			tui.App.QueueUpdateDraw(func() {
				tui.UpdateState(statusMap)
				if logIncoming {
					out, _ := json.MarshalIndent(status, "", " ")
					tui.Output.Write(string(out))
				}
			})

		case "notify_gcode_response":
			tui.App.QueueUpdateDraw(func() {
				for _, line := range incoming.Params {
					tui.Output.Write(line.(string))
				}
			})
		default:

		}
	}
}

func (tui *TUI) loadPrinterInfo(wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := tui.RpcClient.Call("printer.info", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	info, _ := resp.(map[string]interface{})
	tui.App.QueueUpdate(func() {
		tui.hostname = info["hostname"].(string)
	})
}

func (tui *TUI) loadGcodeHelp() {
	resp, err := tui.RpcClient.Call("printer.gcode.help", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	tui.App.QueueUpdate(func() {
		for k, help := range resp.(map[string]interface{}) {
			log.Println("Registered GCODE Command: " + k)
			tui.TabCompleter.RegisterCommand(k, NewGcodeCommand(k, help.(string)))
		}
	})
}

func dumpToJson(obj any) string {
	out, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "<error>"
	}
	return string(out)
}
