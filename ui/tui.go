package ui

import (
	"clipper/build_info"
	"clipper/ui/cmdinput"
	"clipper/wsjsonrpc"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MapleLeafMakers/tview"
	"github.com/bykof/gostradamus"
	"github.com/gdamore/tcell/v2"
	"log"
	"regexp"
	"sync"
	"time"
)

type TUI struct {
	App               *tview.Application
	Pages             *tview.Pages
	Root              *tview.Grid
	Input             *cmdinput.InputField
	TempInput         *cmdinput.InputField
	Output            *LogContent
	RpcClient         *wsjsonrpc.RpcClient
	TabCompleter      cmdinput.TabCompleter
	State             map[string]map[string]interface{}
	HostHeader        *tview.Table
	TemperaturesPanel *TemperaturePanelContent
	ToolheadPanel     *ToolheadPanelContent
	PrintStatusPanel  *PrintStatusPanelContent
	hostname          string
	LeftPanel         *tview.Flex
	LeftPanelSpacer   *tview.Box
	ServerInfo        *ServerInfo
	BuildInfo         *build_info.BuildInfo

	bellPending    bool // should a bell be rung on the next Draw
	focusedControl interface{}
	mu             sync.Mutex
}

func NewTUI(rpcClient *wsjsonrpc.RpcClient, buildInfo *build_info.BuildInfo) *TUI {
	tui := &TUI{
		BuildInfo: buildInfo,
		RpcClient: rpcClient,
	}

	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tui.buildInput()
	tui.buildOutput(100)
	tui.buildWindow()
	tui.App = tview.NewApplication().SetRoot(tui.Pages, true).EnableMouse(true)
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
	if AppConfig.CheckForUpdatesOnStartup {
		time.AfterFunc(time.Second*5, func() {
			tui.checkForUpdates()
		})
	}
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
	tui.TabCompleter.RegisterCommand("/about", Command_About{})
	tui.TabCompleter.RegisterCommand("/updatecheck", Command_UpdateCheck{})
	tui.Input.SetAutocompleteFunc(func(currentText string, cursorPos int) (entries []cmdinput.Suggestion, menuOffset int) {
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
				tui.Output.WriteCommand(ctx["raw"].(string))
				if ok {
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
	tui.LeftPanel = tview.NewFlex().SetDirection(tview.FlexRow)
	tui.Root.AddItem(tui.LeftPanel, 0, 0, 1, 1, 0, 0, true)

	tui.HostHeader = NewHostHeader(tui)
	tui.LeftPanel.AddItem(tui.HostHeader, 1, 1, false)

	tui.PrintStatusPanel = NewPrintStatusPanel(tui)
	printStatusSize := 0
	if state, ok := tui.State["print_stats"]["state"]; ok && state != "standby" {
		printStatusSize = 2 + tui.PrintStatusPanel.GetRowCount()
	}
	tui.LeftPanel.AddItem(tui.PrintStatusPanel.container, printStatusSize, 0, false)

	tui.TemperaturesPanel = NewTemperaturePanel(tui)
	tui.LeftPanel.AddItem(tui.TemperaturesPanel.container, 0, 0, false)

	tui.ToolheadPanel = NewToolheadPanel(tui)
	tui.LeftPanel.AddItem(tui.ToolheadPanel.container, 0, 0, false)

	tui.LeftPanelSpacer = tview.NewBox()
	tui.LeftPanel.AddItem(tui.LeftPanelSpacer, 0, 1, false)
}

func (tui *TUI) initialize() {
	go tui.loadServerInfo()
}

func (tui *TUI) restoreCommandInput() {
	tui.Root.RemoveItem(tui.TempInput)
	tui.Root.AddItem(tui.Input, 1, 0, 1, 2, 0, 0, true)
	tui.App.SetFocus(tui.Input)
}

func (tui *TUI) promptForInput(prompt string, defaultValue string, callback func(bool, string)) {
	tui.TempInput = cmdinput.NewInputField().
		SetText(defaultValue).
		SetLabel(prompt).
		SetDoneFunc(func(key tcell.Key) {
			tui.restoreCommandInput()
			callback(key == tcell.KeyEnter, tui.TempInput.GetText())
		}).SetLabelStyle(
		tcell.StyleDefault.Background(AppConfig.Theme.InputBackgroundColor.Color()).
			Foreground(AppConfig.Theme.InputPromptColor.Color()).Bold(true)).
		SetFieldBackgroundColor(AppConfig.Theme.InputBackgroundColor.Color()).
		SetFieldTextColor(AppConfig.Theme.InputTextColor.Color()).
		SetPlaceholderStyle(tcell.StyleDefault.Background(AppConfig.Theme.InputBackgroundColor.Color()).Foreground(AppConfig.Theme.InputPlaceholderColor.Color()))

	tui.Root.RemoveItem(tui.Input)
	tui.Root.AddItem(tui.TempInput, 1, 0, 1, 2, 0, 0, true)
	tui.App.QueueUpdateDraw(func() {

		tui.App.SetFocus(tui.TempInput)
		// this is a dirty hack but it's the only thing that seems to work semi-consistently
		time.AfterFunc(time.Millisecond*50, func() {
			tui.TempInput.Select(0, len(tui.TempInput.GetText()))
		})
	})
}

func (tui *TUI) buildOutput(numLines int) {

	output := tview.NewTable()
	ts := gostradamus.DateTimeFromTime(time.Time{})
	lines := make([]LogEntry, numLines)
	i := 0
	for i = 0; i < numLines-4; i++ {
		lines[i] = LogEntry{MsgTypeInternal, ts, ""}
	}
	lines[i+0] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]┏┓┓ ┳ "}
	lines[i+1] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]┃ ┃ ┃┏┓┏┓┏┓┏┓"}
	lines[i+2] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]┗┛┗┛┻┣┛┣┛┗ ┛ "}
	lines[i+3] = LogEntry{MsgTypeInternal, gostradamus.Now(), "[yellow]     ┛ ┛ [-]" + tui.BuildInfo.VersionString}

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
		AddItem(tui.Input, 1, 0, 1, 2, 0, 0, true).
		AddItem(tui.Output.table, 0, 1, 1, 1, 0, 0, false)
	tui.Pages = tview.NewPages()
	tui.Pages.AddPage("main", tui.Root, true, true)
}

func getObjectList(client *wsjsonrpc.RpcClient) []string {
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
		SetPlaceholderStyle(tcell.StyleDefault.Background(AppConfig.Theme.InputBackgroundColor.Color()).Foreground(AppConfig.Theme.InputPlaceholderColor.Color())).
		SetAutocompleteStyles(
			AppConfig.Theme.AutocompleteBackgroundColor.Color(),
			tcell.StyleDefault.Foreground(AppConfig.Theme.AutocompleteTextColor.Color()),
			tcell.StyleDefault.Foreground(AppConfig.Theme.AutocompleteBackgroundColor.Color()).Background(AppConfig.Theme.AutocompleteTextColor.Color()),
			tcell.StyleDefault.Foreground(AppConfig.Theme.AutocompleteHelpColor.Color()))

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

		if tui.PrintStatusPanel != nil {
			tui.PrintStatusPanel.container.SetBorderColor(tview.Styles.BorderColor)
			tui.PrintStatusPanel.table.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.PrintStatusPanel.container.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
			tui.PrintStatusPanel.container.SetTitleColor(tview.Styles.TitleColor)
		}
	}

}

func (tui *TUI) subscribe() {
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
	tui.State = state
	tui.App.QueueUpdateDraw(func() {
		tui.initializeServerUI()
	})
}

func (tui *TUI) UpdateState(statusChanges map[string]map[string]interface{}) {
	for key, objectStatus := range statusChanges {
		switch key {
		// handle some special cases here
		case "print_stats":
			newState, ok := objectStatus["state"].(string)
			if ok {
				// print_stats.state changed, might need to show/hide the print panel
				if newState == "standby" {
					tui.hidePrintStatus()
				} else if tui.State["print_stats"]["state"] == "standby" {
					tui.showPrintStatus()
				}
			}
		}

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

func (tui *TUI) checkForUpdates() {
	build_info.CheckForUpdates(tui.BuildInfo.VersionString, tui.onUpdateCheck)
}

func (tui *TUI) handleIncoming() {

	for {
		incoming, ok := <-tui.RpcClient.Incoming
		if !ok {
			return
		}
		switch incoming.Method {
		case "notify_status_update":
			params, ok := incoming.Params.([]interface{})
			if !ok {
				panic(fmt.Sprintf("Unexpected non-array params, %#v", params))
			}
			status := params[0].(map[string]interface{})
			statusMap, _ := toStatusMap(status)
			tui.App.QueueUpdateDraw(func() {
				tui.UpdateState(statusMap)
			})

		case "notify_gcode_response":
			params, ok := incoming.Params.([]interface{})

			if !ok {
				panic(fmt.Sprintf("Unexpected non-array params, %#v", params))
			}

			filtered := make([]string, 0, len(params))
			for _, line := range params {
				if passesConsoleFilters(line.(string)) {
					filtered = append(filtered, line.(string))
				}
			}
			tui.App.QueueUpdateDraw(func() {
				for _, line := range filtered {
					tui.Output.WriteResponse(line)
				}
			})

		case "notify_klippy_ready":
			log.Println("notify_klippy_ready!")
			tui.ServerInfo.KlippyConnected = true
			tui.ServerInfo.KlippyState = "ready"
			go tui.loadServerInfo()

		case "notify_klippy_shutdown":
			log.Println("notify_klippy_shutdown!")
		case "notify_klippy_disconnected":
			go tui.loadServerInfo()
			tui.App.QueueUpdateDraw(func() {
				tui.removeServerUI()
			})
			log.Println("notify_klippy_disconnected!")

		case "notify_filelist_changed":
		case "notify_update_response":
		case "notify_update_refreshed":
		case "notify_cpu_throttled":
		case "notify_history_changed":
		case "notify_user_created":
		case "notify_user_deleted":
		case "notify_user_logged_out":
		case "notify_service_state_changed":
		case "notify_job_queue_changed":
		case "notify_button_event":
		case "notify_announcement_update":
		case "notify_announcement_dismissed":
		case "notify_announcement_wake":
		case "notify_sudo_alert":
		case "notify_webcams_changed":
		case "notify_active_spool_set":
		case "notify_spoolman_status_changed":
		case "notify_agent_event":
		case "sensors:sensor_update":
		default:

		}
	}
}

func passesConsoleFilters(consoleOutput string) bool {
	for _, pattern := range AppConfig.ConsoleFilterPatterns {
		matches, err := regexp.MatchString(pattern, consoleOutput)
		if err != nil {
			panic(err)
		}
		if matches {
			return false
		}
	}
	return true
}

func (tui *TUI) loadServerInfo() {
	resp, err := tui.RpcClient.Call("server.info", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	var si ServerInfo
	err = json.Unmarshal(bytes, &si)
	if err != nil {
		panic(err)
	}
	tui.mu.Lock()
	tui.ServerInfo = &si
	tui.mu.Unlock()
	if si.KlippyConnected && si.KlippyState == "ready" {
		log.Println("Klippy connected, loading server ui")
		tui.loadPrinterInfo()
		tui.loadGcodeHelp()
		tui.subscribe()
	} else {
		log.Println("Klippy not ready, need UI for this")
	}
	tui.App.QueueUpdateDraw(func() {})
}

func (tui *TUI) loadPrinterInfo() {
	resp, err := tui.RpcClient.Call("printer.info", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	info, _ := resp.(map[string]interface{})
	tui.App.QueueUpdate(func() {
		tui.mu.Lock()
		tui.hostname = info["hostname"].(string)
		tui.mu.Unlock()
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
	go func() {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteCommand(gcode)
		})
		_, err := tui.RpcClient.Call("printer.gcode.script", map[string]interface{}{"script": gcode})
		if err != nil {
			tui.App.QueueUpdateDraw(func() {
				tui.Output.WriteError(err.Error())
			})
		}
	}()
}

func (tui *TUI) SwitchFocus(widget tview.Primitive) {
	tui.App.SetFocus(widget)
}

func (tui *TUI) onConnect() {
	tui.App.QueueUpdateDraw(func() {
		tui.Output.WriteResponse("Connected!")
	})
	tui.initialize()
}

func (tui *TUI) onDisconnect() {
	tui.App.QueueUpdateDraw(func() {
		tui.removeServerUI()
		tui.Output.WriteResponse("Disconnected.")
	})
}

func (tui *TUI) connectOnStartup() {
	tui.RpcClient.SetOnConnectFunc(tui.onConnect)
	tui.RpcClient.SetOnDisconnectFunc(tui.onDisconnect)
	if tui.RpcClient.Url != nil && !tui.RpcClient.IsConnected {
		log.Printf("Connecting to %#v", tui.RpcClient.Url)
		if err := tui.RpcClient.Connect(); err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	} else if tui.RpcClient.IsConnected {
		tui.initialize()
	}
}

func (tui *TUI) initializeServerUI() {
	tui.TemperaturesPanel.loadSensors()
	if tui.State["print_stats"]["state"] != "standby" {
		tui.showPrintStatus()
	}
	tui.LeftPanel.ResizeItem(tui.TemperaturesPanel.container, tui.TemperaturesPanel.GetRowCount()+2, 0)
	tui.LeftPanel.ResizeItem(tui.ToolheadPanel.container, tui.ToolheadPanel.GetRowCount()+2, 0)
}

func (tui *TUI) removeServerUI() {
	tui.LeftPanel.ResizeItem(tui.TemperaturesPanel.container, 0, 0)
	tui.LeftPanel.ResizeItem(tui.ToolheadPanel.container, 0, 0)
}

func (tui *TUI) hidePrintStatus() {
	tui.LeftPanel.ResizeItem(tui.PrintStatusPanel.container, 0, 0)
}

func (tui *TUI) showPrintStatus() {
	tui.LeftPanel.ResizeItem(tui.PrintStatusPanel.container, 2+tui.PrintStatusPanel.GetRowCount(), 0)
}

func (tui *TUI) showAboutDialog() {
	logo := ("┏┓┓ ┳        \n" +
		"┃ ┃ ┃┏┓┏┓┏┓┏┓\n" +
		"┗┛┗┛┻┣┛┣┛┗ ┛ \n" +
		"     ┛ ┛     ")
	modal := tview.NewModal().AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		tui.Pages.RemovePage("about-modal")
	})
	modal.SetText(fmt.Sprintf(`%s
%s

[::i]Copyright © 2024, MapleLeafMakers[::-]

For new versions, bug reports or information on contributing, see the [:::http://github.com/MapleLeafMakers/CLIpper]Github Repository[:::-]

CLIpper is free and open source software licensed under the GNU General Public License Version 3.

[::i]%s-%s
[::i](%s / %s / %s)[::-]
`, logo, tui.BuildInfo.VersionString, tui.BuildInfo.VersionString, tui.BuildInfo.CommitHash,
		tui.BuildInfo.BuildTime.Format("YYYY-MM-DD, hh:mm:ss"), tui.BuildInfo.BuildOS, tui.BuildInfo.BuildArch))

	tui.Pages.AddPage("about-modal", modal, false, true)
}

func (tui *TUI) onUpdateCheck(updateAvailable bool, newVersion string, newVersionUrl string, err error) {
	if err != nil {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteResponse("Update check failed: " + err.Error())
		})
	} else if updateAvailable {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteInternal(fmt.Sprintf("[::b]%s of CLIpper is available! [:::%s]Get it now![:::-]", newVersion, newVersionUrl))
		})
	} else {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteResponse("You are running the latest release of CLIpper")
		})
	}
}

func dumpToJson(obj any) string {
	out, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "<error>"
	}
	return string(out)
}
