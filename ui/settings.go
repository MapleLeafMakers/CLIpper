package ui

import (
	"clipper/ui/cmdinput"
)

type Settings map[string]interface{}

var AppSettings = Settings{
	"logIncoming":           false,
	"theme.borderColor":     "#ffffff",
	"theme.timestampFormat": "hh:mm:ss",
}

func NewSettingsCompleter() cmdinput.StaticTokenCompleter {

	completer := cmdinput.StaticTokenCompleter{
		ContextKey: "setting",
		Registry: map[string]cmdinput.TokenCompleter{
			"logIncoming":           cmdinput.NewBoolTokenCompleter("value", nil),
			"theme.borderColor":     cmdinput.NewColorTokenCompleter("value", nil),
			"theme.timestampFormat": cmdinput.AnythingCompleter{"value"},
		},
	}
	return completer
}
