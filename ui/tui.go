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
	"sync"
	"time"
)

type TUI struct {
	App               *tview.Application
	Root              *tview.Grid
	Input             *cmdinput.InputField
	Output            *LogContent
	RpcClient         *jsonrpcclient.Client
	TabCompleter      cmdinput.TabCompleter
	State             map[string]map[string]interface{}
	HostHeader        *tview.Table
	TemperaturesPanel *TemperaturePanelContent
	ToolheadPanel     *ToolheadPanelContent
	hostname          string
	LeftPanel         *tview.Flex
	LeftPanelSpacer   *tview.Box

	bellPending    bool // should a bell be rung on the next Draw
	focusedControl interface{}
}

func NewTUI(rpcClient *jsonrpcclient.Client) *TUI {
	tui := &TUI{
		RpcClient: rpcClient,
	}

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tui.buildInput()
	tui.buildOutput(100)
	tui.buildWindow()
	tui.App = tview.NewApplication().SetRoot(tui.Root, true).EnableMouse(true)
	tui.App.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if tui.bellPending {
			tui.bellPending = false
			err := screen.Beep()
			if err != nil {
				log.Println(err)
			}
		}
		return false
	})
	tui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlT:
			tui.SwitchFocus(tui.TemperaturesPanel.container)
		case tcell.KeyCtrlO:
			tui.SwitchFocus(tui.ToolheadPanel.container)
		case tcell.KeyCtrlC:
			tui.SwitchFocus(tui.Input)
		case tcell.KeyPgUp, tcell.KeyPgDn:
			tui.Output.table.InputHandler()(event, func(p tview.Primitive) {})
			return nil
		case tcell.KeyTab:
			return event
		default:
			return event
		}
		return nil // never happens
	})
	tui.buildLeftPanel()
	tui.SwitchFocus(tui.Input)
	tui.UpdateTheme()

	go tui.handleIncoming()
	go tui.connectOnStartup()
	return tui
}

func (tui *TUI) Bell() {
	tui.bellPending = true
}

func (tui *TUI) buildInput() {
	//tui.Input = tview.NewInputField()
	tui.Input = cmdinput.NewInputField().
		SetPlaceholder("Enter GCODE Commands or / commands").
		SetLabel("> ")

	tui.TabCompleter = cmdinput.NewTabCompleter()
	tui.TabCompleter.RegisterCommand("/set", Command_Set{})
	tui.TabCompleter.RegisterCommand("/settings", Command_Settings{})
	tui.TabCompleter.RegisterCommand("/quit", Command_Quit{})
	tui.TabCompleter.RegisterCommand("/rpc", Command_RPC{})
	tui.TabCompleter.RegisterCommand("/restart", Command_Restart{})
	tui.TabCompleter.RegisterCommand("/firmware_restart", Command_FirmwareRestart{})
	tui.TabCompleter.RegisterCommand("/estop", Command_EStop{})
	tui.TabCompleter.RegisterCommand("/print", Command_Print{})
	tui.TabCompleter.RegisterCommand("/disconnect", Command_Disconnect{})
	tui.TabCompleter.RegisterCommand("/connect", Command_Connect{})

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
				cmd, ok := ctx["cmd"]
				// ew
				if ok {
					tui.Output.WriteCommand(ctx["raw"].(string))
					cmd2, ok2 := cmd.(cmdinput.Command)
					if ok2 {

						go cmd2.Call(ctx)
					}
				} else {
					//not a registered command, send it as gcode.
					go (func() { NewGcodeCommand("", "").Call(ctx) })()
				}
				tui.Input.NewCommand()
			} else if err.Error() == "NoInput" {

			} else {
				tui.Output.WriteResponse(err.Error())
			}
		default:
		}
	})
}

func (tui *TUI) buildLeftPanel() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	tui.LeftPanel = flex
	tui.HostHeader = NewHostHeader(tui)
	tui.Root.AddItem(flex, 0, 0, 1, 1, 0, 0, true)
	flex.AddItem(tui.HostHeader, 1, 1, false)

	tempPanel := NewTemperaturePanel(tui)
	tui.TemperaturesPanel = &tempPanel

	toolheadPanel := NewToolheadPanel(tui)
	tui.ToolheadPanel = &toolheadPanel

	tui.LeftPanelSpacer = tview.NewBox()
	flex.AddItem(tui.LeftPanelSpacer, 0, 1, false)
}

func (tui *TUI) initialize() {
	log.Println("initializing")
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go tui.loadPrinterInfo(wg)
	go tui.subscribe(wg)
	go tui.loadGcodeHelp()
	log.Println("Waiting for subscribe")
	wg.Wait()
}

func (tui *TUI) buildOutput(numLines int) {

	output := tview.NewTable()
	ts := gostradamus.DateTimeFromTime(time.Time{})
	lines := make([]LogEntry, numLines)
	i := 0
	for i = 0; i < numLines-6; i++ {
		lines[i] = LogEntry{MsgTypeInternal, ts, ""}
	}
	lines[i+0] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]   ________    ____                     "}
	lines[i+1] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]  / ____/ /   /  _/___  ____  ___  _____"}
	lines[i+2] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow] / /   / /    / // __ \\/ __ \\/ _ \\/ ___/"}
	lines[i+3] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]/ /___/ /____/ // /_/ / /_/ /  __/ /    "}
	lines[i+4] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]\\____/_____/___/ .___/ .___/\\___/_/     "}
	lines[i+5] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]              /_/   /_/                 "}

	tui.Output = &LogContent{
		table:   output,
		entries: lines,
		tui:     tui,
	}
	output.SetContent(tui.Output)
	output.ScrollToEnd()
}

func (tui *TUI) buildWindow() {
	tui.Root = tview.NewGrid().
		SetBordersColor(tview.Styles.BorderColor).
		SetRows(0, 1).
		SetColumns(36, 0).
		SetBorders(true).
		AddItem(tui.Input, 1, 0, 1, 2, 0, 0, false).
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

func (tui *TUI) UpdateTheme() {
	tview.Styles.PrimitiveBackgroundColor = AppConfig.Theme.BackgroundColor.Color()
	tview.Styles.BorderColor = AppConfig.Theme.BorderColor.Color()
	tview.Styles.PrimaryTextColor = AppConfig.Theme.PrimaryTextColor.Color()
	tview.Styles.SecondaryTextColor = AppConfig.Theme.SecondaryTextColor.Color()
	tview.Styles.GraphicsColor = AppConfig.Theme.GraphicsColor.Color()
	tview.Styles.TitleColor = AppConfig.Theme.TitleColor.Color()

	// update the colors of everything, since there doesn't seem to be a better way
	if tui.Root != nil {
		tui.Root.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		tui.Root.SetBordersColor(tview.Styles.BorderColor)
	}

	tui.Output.table.SetBackgroundColor(AppConfig.Theme.ConsoleBackgroundColor.Color())

	tui.Input.SetLabelStyle(
		tcell.StyleDefault.Background(AppConfig.Theme.InputBackgroundColor.Color()).
			Foreground(AppConfig.Theme.InputPromptColor.Color()).Bold(true)).
		SetFieldBackgroundColor(AppConfig.Theme.InputBackgroundColor.Color()).
		SetFieldTextColor(AppConfig.Theme.InputTextColor.Color()).
		SetAutocompleteStyles(
			AppConfig.Theme.AutocompleteBackgroundColor.Color(),
			tcell.StyleDefault.Foreground(AppConfig.Theme.AutocompleteTextColor.Color()),
			tcell.StyleDefault.Foreground(AppConfig.Theme.AutocompleteBackgroundColor.Color()).Background(AppConfig.Theme.AutocompleteTextColor.Color())).
		SetPlaceholderStyle(tcell.StyleDefault.Background(AppConfig.Theme.InputBackgroundColor.Color()).Foreground(AppConfig.Theme.InputPlaceholderColor.Color()))

	if tui.LeftPanel != nil {
		tui.LeftPanel.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
		tui.LeftPanelSpacer.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

		if tui.ToolheadPanel != nil {
			tui.ToolheadPanel.container.SetBorderColor(tview.Styles.BorderColor)
			tui.ToolheadPanel.container.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.ToolheadPanel.table.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.ToolheadPanel.container.SetTitleColor(tview.Styles.TitleColor)
		}

		if tui.TemperaturesPanel != nil {
			tui.TemperaturesPanel.container.SetBorderColor(tview.Styles.BorderColor)
			tui.TemperaturesPanel.table.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.TemperaturesPanel.container.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.TemperaturesPanel.container.SetTitleColor(tview.Styles.TitleColor)
		}
	}

}

func (tui *TUI) subscribe(wg *sync.WaitGroup) {
	defer wg.Done()
	objList := getObjectList(tui.RpcClient)
	subs := make(map[string]interface{}, len(objList))
	for _, key := range objList {
		subs[key] = nil
	}
	log.Println("calling printer.objects.subscribe")
	resp, err := tui.RpcClient.Call("printer.objects.subscribe", map[string]interface{}{
		"objects": subs,
	})
	if err != nil {
		panic(err)
	}
	log.Println("got a response")
	asMap, _ := resp.(map[string]interface{})
	objectsAsMap, _ := asMap["status"].(map[string]interface{})
	state := make(map[string]map[string]interface{}, len(objectsAsMap))
	for k, v := range objectsAsMap {
		state[k] = v.(map[string]interface{})
	}
	tui.State = state
	log.Println("Queuing initServerUI")
	tui.App.QueueUpdateDraw(func() {
		tui.initializeServerUI()
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
	log.Println("handleIncoming")
	for {
		incoming := <-tui.RpcClient.Incoming
		switch incoming.Method {
		case "_client_connected":
			// rpcclient connected to server, re-init everything
			// show some indication of connection status
			tui.App.QueueUpdateDraw(func() {
				tui.Output.WriteResponse("Connected to " + tui.RpcClient.Url)
			})
			tui.initialize()

		case "_client_disconnected":
			// rpcclient disconnected (may have been intentional) from server, stop doing stuff
			// show some indication of connection status
			tui.App.QueueUpdateDraw(func() {
				tui.removeServerUI()
				tui.Output.WriteResponse("Disconnected.")
			})
		case "notify_status_update":
			status := incoming.Params[0].(map[string]interface{})
			statusMap, _ := toStatusMap(status)
			tui.App.QueueUpdateDraw(func() {
				tui.UpdateState(statusMap)
				if AppConfig.LogIncoming {
					out, _ := json.MarshalIndent(status, "", " ")
					log.Println(string(out))
					tui.Output.WriteResponse(string(out))
				}
			})

		case "notify_gcode_response":
			tui.App.QueueUpdateDraw(func() {
				for _, line := range incoming.Params {
					tui.Output.WriteResponse(line.(string))
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
			tui.TabCompleter.RegisterCommand(k, NewGcodeCommand(k, help.(string)))
		}
	})
}

func (tui *TUI) ExecuteGcode(gcode string) {
	log.Println("Executing GCODE:\n" + gcode)
}

func (tui *TUI) SwitchFocus(widget tview.Primitive) {
	tui.App.SetFocus(widget)
}

func (tui *TUI) connectOnStartup() {
	if tui.RpcClient.Url != "" && !tui.RpcClient.IsConnected {
		if err := tui.RpcClient.Start(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	} else if tui.RpcClient.IsConnected {
		tui.initialize()
	}
}

func (tui *TUI) initializeServerUI() {
	log.Println("Initializing Server UI")
	tui.TemperaturesPanel.loadSensors()
	tui.LeftPanel.RemoveItem(tui.LeftPanelSpacer)
	tui.LeftPanel.AddItem(tui.TemperaturesPanel.container, len(tui.TemperaturesPanel.sensors)+2, 0, false)
	tui.LeftPanel.AddItem(tui.ToolheadPanel.container, 5, 0, false)
	tui.LeftPanel.AddItem(tui.LeftPanelSpacer, 0, 1, false)
}

func (tui *TUI) removeServerUI() {
	tui.LeftPanel.RemoveItem(tui.TemperaturesPanel.container)
	tui.LeftPanel.RemoveItem(tui.ToolheadPanel.container)
}

func dumpToJson(obj any) string {
	out, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "<error>"
	}
	return string(out)
}
