package ui

import (
	"bytes"
	"clipper/ui/cmdinput"
	"clipper/wsjsonrpc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type GcodeCommand struct {
	Name string
	Help string
}

func (g GcodeCommand) GetHelp() string { return g.Help }
func (g GcodeCommand) Call(ctx cmdinput.CommandContext) error {
	rawCommand := ctx["raw"].(string)
	tui := ctx["tui"].(*TUI)

	_, err := tui.RpcClient.Call("printer.gcode.script", map[string]interface{}{"script": rawCommand})
	if err != nil {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteResponse("Error: " + err.Error())
		})
	}
	return nil
}

func (g GcodeCommand) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

func NewGcodeCommand(name, help string) GcodeCommand {
	return GcodeCommand{name, help}
}

// /set command
type Command_Set struct{}

func (c Command_Set) GetHelp() string { return "Change configuration" }
func (c Command_Set) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	key, _ := ctx["setting"].(string)
	value := ctx["value"]
	// todo: figure out a generic way to set non-simple values
	if key == "consoleInputPatterns" {
		tokens := strings.SplitN(ctx["raw"].(string), " ", 3)
		rawJson := tokens[2]
		var mapVal map[string]interface{}
		var arrVal []interface{}
		err := json.Unmarshal([]byte(rawJson), &mapVal)
		if err != nil {
			err = json.Unmarshal([]byte(rawJson), &arrVal)
			if err != nil {
				return errors.New("Invalid value for " + key + ": " + rawJson)
			}
			value = arrVal
		} else {
			value = mapVal
		}
	}

	AppConfig.Set(key, value)
	tui.Output.WriteResponse(fmt.Sprintf("Set %s to %+v", key, value))
	AppConfig.Save()
	tui.UpdateTheme()
	return nil
}

func (c Command_Set) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return NewSettingsCompleter()
}

// /settings command
type Command_Settings struct{}

func (c Command_Settings) GetHelp() string { return "Display configuration" }
func (c Command_Settings) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	str, err := json.MarshalIndent(AppConfig, "", " ")
	if err != nil {
		panic(err)
	} else {
		tui.Output.WriteResponse(string(str))
	}
	return nil
}

func (c Command_Settings) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /Quit
type Command_Quit struct{}

func (c Command_Quit) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Close()
	tui.App.Stop()
	return nil
}
func (c Command_Quit) GetHelp() string { return "Quit the application" }
func (c Command_Quit) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /rpc

type Command_RPC struct{}

func (c Command_RPC) GetHelp() string { return "Make a moonraker RPC call" }
func (c Command_RPC) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	rawCommand := ctx["raw"].(string)
	parts := strings.SplitN(rawCommand, " ", 3)
	var payload map[string]interface{}
	if len(parts) == 2 {
		payload = map[string]interface{}{}
	} else {
		err := json.Unmarshal([]byte(parts[2]), &payload)
		if err != nil {
			return err
		}
	}
	resp, _ := tui.RpcClient.Call(ctx["method"].(string), payload)
	tui.App.QueueUpdateDraw(func() {
		tui.Output.WriteResponse(dumpToJson(resp))
	})

	return nil
}

func (c Command_RPC) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	reg := make(map[string]cmdinput.TokenCompleter, len(MoonrakerRPCMethods))
	for _, method := range MoonrakerRPCMethods {
		reg[method] = nil
	}
	return cmdinput.StaticTokenCompleter{
		ContextKey: "method",
		Registry:   reg,
	}
}

// /restart
type Command_Restart struct{}

func (c Command_Restart) GetHelp() string { return "Restart klippy" }
func (c Command_Restart) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Call("printer.restart", map[string]interface{}{})
	return nil
}

func (c Command_Restart) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /firmware_restart
type Command_FirmwareRestart struct{}

func (c Command_FirmwareRestart) GetHelp() string { return "Restart klippy and MCUs" }
func (c Command_FirmwareRestart) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Call("printer.firmware_restart", map[string]interface{}{})
	return nil
}

func (c Command_FirmwareRestart) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /estop
type Command_EStop struct{}

func (c Command_EStop) GetHelp() string { return "Emergency Stop!" }
func (c Command_EStop) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Call("printer.emergency_stop", map[string]interface{}{})
	return nil
}

func (c Command_EStop) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /print

type Command_Print struct{}

func (c Command_Print) GetHelp() string { return "Upload and print a gcode file" }
func (c Command_Print) Call(ctx cmdinput.CommandContext) error {

	filename := ctx["file"].(string)
	tui, _ := ctx["tui"].(*TUI)
	httpUrl := *tui.RpcClient.Url
	httpUrl.Scheme = "http"
	httpUrl.Path = "/server/files/upload"

	file, _ := os.Open(filename)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("print", "true")
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	r, _ := http.NewRequest("POST", httpUrl.String(), body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	_, err := client.Do(r)
	return err
}

func (c Command_Print) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return cmdinput.NewFileTokenCompleter("file", nil)
}

// /disconnect

type Command_Disconnect struct{}

func (c Command_Disconnect) GetHelp() string { return "Disconnect from the server" }
func (c Command_Disconnect) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Close()
	return nil
}

func (c Command_Disconnect) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /connect

type Command_Connect struct{}

func (c Command_Connect) GetHelp() string { return "Connect to a server" }
func (c Command_Connect) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	if tui.RpcClient.IsConnected {
		return errors.New("Already connected.")
	} else {
		raw := ctx["url"].(string)
		if !strings.Contains("/", strings.ToLower(raw)) {
			raw = "ws://" + raw + "/websocket"
		}
		serverUrl, err := url.Parse(raw)
		if err != nil {
			panic(err)
		}
		tui.App.QueueUpdateDraw(func() {
			tui.Output.WriteResponse(fmt.Sprintf("Connecting to %s", serverUrl.String()))
		})
		tui.RpcClient = wsjsonrpc.NewWebSocketClient(serverUrl)
		tui.RpcClient.SetOnConnectFunc(tui.onConnect)
		tui.RpcClient.SetOnDisconnectFunc(tui.onDisconnect)
		go tui.handleIncoming()

		if err := tui.RpcClient.Connect(); err != nil {
			return errors.New(fmt.Sprintf("Failed to connect: %v", err))
		}
	}
	return nil
}

func (c Command_Connect) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return cmdinput.AnythingCompleter{"url"}
}

type Command_About struct{}

func (c Command_About) GetHelp() string { return "About CLIpper" }
func (c Command_About) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	tui.App.QueueUpdateDraw(func() {
		tui.showAboutDialog()
	})
	return nil
}

func (c Command_About) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return cmdinput.AnythingCompleter{"url"}
}
