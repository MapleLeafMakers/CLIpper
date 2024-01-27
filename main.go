package main

import (
	"clipper/jsonrpcclient"
	"clipper/ui"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var version = "?"
var commit = ""

func configureLogger() {
	if os.Getenv("DEBUG") == "1" {
		logFile, err := os.OpenFile("./debug.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(io.Discard)
	}

}

func StartInteractive(url string) {
	versionString := version
	if commit != "" {
		versionString = versionString + "-" + commit[:7]
	}
	rpcClient := jsonrpcclient.NewClient(url)
	defer func() {
		for len(rpcClient.Incoming) > 0 {
			<-rpcClient.Incoming
		}
		rpcClient.Stop(false)
	}()

	tui := ui.NewTUI(rpcClient)
	if err := tui.App.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func main() {

	configureLogger()
	ui.AppConfig.Load()
	log.Printf("%#v", ui.AppConfig)
	args := os.Args[1:]
	var url string
	switch len(args) {
	case 2:
		url = "ws://" + args[0] + ":" + string(args[1]) + "/websocket"
	case 1:
		if strings.Contains(args[0], "://") {
			url = args[0]
		} else {
			url = "ws://" + args[0] + "/websocket"
		}
	default:
		url = ""
	}
	StartInteractive(url)
}
