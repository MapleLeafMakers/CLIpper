package main

import (
	"clipper/build_info"
	"clipper/ui"
	"clipper/wsjsonrpc"
	"fmt"
	"github.com/bykof/gostradamus"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var buildVersion string = ""
var buildTime string = ""
var buildCommit = ""
var buildArch = ""
var buildOS = ""

var buildInfo = func() *build_info.BuildInfo {
	bt, err := strconv.Atoi(buildTime)
	if err != nil {
		bt = 0
	}

	return &build_info.BuildInfo{
		BuildArch:     buildArch,
		BuildOS:       buildOS,
		VersionString: buildVersion,
		CommitHash:    buildCommit,
		BuildTime:     gostradamus.FromUnixTimestamp(int64(bt)),
	}
}()

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

	rpcClient := wsjsonrpc.NewWebSocketClient(Url)
	defer func() {
		for len(rpcClient.Incoming) > 0 {
			<-rpcClient.Incoming
		}
		rpcClient.Close()
	}()

	tui := ui.NewTUI(rpcClient, buildInfo)
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
