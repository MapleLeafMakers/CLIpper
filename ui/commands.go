package ui

import (
	"clipper/ui/cmdinput"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type GcodeCommand struct {
	Name string
	Help string
}

func (g GcodeCommand) Call(ctx cmdinput.CommandContext) error {
	rawCommand := ctx["raw"].(string)
	tui := ctx["tui"].(*TUI)

	go tui.App.QueueUpdateDraw(func() {
		tui.Output.Write(rawCommand)
	})

	_, err := tui.RpcClient.Call("printer.gcode.script", map[string]interface{}{"script": rawCommand})
	if err != nil {
		tui.App.QueueUpdateDraw(func() {
			tui.Output.Write("Error: " + err.Error())
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

var Completer_Set = cmdinput.StaticTokenCompleter{
	"setting",
	map[string]cmdinput.TokenCompleter{
		"foo": cmdinput.NewBoolTokenCompleter("value", nil),
		"bar": cmdinput.NewBoolTokenCompleter("value", nil),
	},
}

func (c Command_Set) Call(ctx cmdinput.CommandContext) error {
	log.Printf("Set Command called with %+v", ctx)
	tui, _ := ctx["tui"].(*TUI)
	key, _ := ctx["setting"].(string)
	value := ctx["value"]
	tui.Settings[key] = value
	tui.Output.Write(fmt.Sprintf("Set %s to %+v", key, value))
	return nil
}

func (c Command_Set) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return NewSettingsCompleter()
}

// /settings command
type Command_Settings struct{}

func (c Command_Settings) Call(ctx cmdinput.CommandContext) error {
	tui, _ := ctx["tui"].(*TUI)
	str, err := json.MarshalIndent(tui.Settings, "", " ")
	if err != nil {
		panic(err)
	} else {
		tui.Output.Write(string(str))
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
	tui.App.Stop()
	return nil
}

func (c Command_Quit) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return nil
}

// /rpc
type Command_RPC struct{}

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
			log.Println("Returning Error", err)
			return err
		}
	}
	resp, _ := tui.RpcClient.Call(ctx["method"].(string), payload)
	tui.App.QueueUpdateDraw(func() {
		tui.Output.Write(dumpToJson(resp))
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

func (c Command_Print) Call(ctx cmdinput.CommandContext) error {

	//file := ctx["file"].(string)
	tui, _ := ctx["tui"].(*TUI)
	tui.RpcClient.Upload(ctx["file"].(string), true)
	return nil
}

func (c Command_Print) GetCompleter(ctx cmdinput.CommandContext) cmdinput.TokenCompleter {
	return cmdinput.NewFileTokenCompleter("file", nil)
}
