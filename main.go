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
	"runtime/debug"
	"strings"
	"time"
)

var buildVersion = "devel"

var buildInfo = func() *build_info.BuildInfo {
	bi := build_info.BuildInfo{
		VersionString: buildVersion,
		BuildTime:     gostradamus.DateTimeFromTime(time.Now()),
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			switch s.Key {
			case "GOARCH":
				bi.BuildArch = s.Value
			case "GOOS":
				bi.BuildOS = s.Value
			case "vcs.revision":
				bi.CommitHash = s.Value[:min(len(s.Value), 7)]
			case "vcs.time":
				bt, err := time.Parse(time.RFC3339, s.Value)
				if err == nil {
					bi.BuildTime = gostradamus.DateTimeFromTime(bt)
				}
			}
		}
	}

	return &bi
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
	err := ui.AppConfig.Load()
	if err != nil {
		fmt.Println("could not load configuration:", err)
		os.Exit(1)
	}
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
