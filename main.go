package main

import (
	"clipper/ui"
	"clipper/wsjsonrpc"
	"fmt"
	"io"
	"log"
	"net/url"
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

func StartInteractive(serverUrl string) {
	var Url *url.URL
	var err error

	if serverUrl == "" {
		Url = nil
	} else {
		Url, err = url.Parse(serverUrl)
		if err != nil {
			panic(err)
		}
	}

	versionString := version
	//if commit != "" {
	//	versionString = versionString + "-" + commit[:7]
	//}

	rpcClient := wsjsonrpc.NewWebSocketClient(Url)
	defer func() {
		for len(rpcClient.Incoming) > 0 {
			<-rpcClient.Incoming
		}
		rpcClient.Close()
	}()

	tui := ui.NewTUI(rpcClient, versionString)
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
