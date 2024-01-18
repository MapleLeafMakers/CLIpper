package main

import (
	"clipper/jsonrpcclient"
	"clipper/ui"
	"fmt"
	"log"
	"os"
)

var version = "?"
var commit = ""

func main() {
	versionString := version
	if commit != "" {
		versionString = versionString + "-" + commit[:7]
	}
	// Initialize JSON-RPC WebSocket jsonrpcclient
	rpcClient := jsonrpcclient.NewClient("ws://trident/websocket")
	defer rpcClient.Close()

	if err := rpcClient.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Load some text for our viewport

	tui := ui.NewTUI(rpcClient, versionString)
	//tui.Send(tea.QuitMsg{})
	if _, err := tui.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
