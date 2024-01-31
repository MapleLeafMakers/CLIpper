package ui

import (
	"clipper/ui/cmdinput"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/kirsle/configdir"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type ConfigColor string

func (c ConfigColor) Color() tcell.Color {
	return tcell.GetColor(string(c))
}

type ThemeConfig struct {
	BackgroundColor    ConfigColor `json:"backgroundColor"`
	BorderColor        ConfigColor `json:"borderColor"`
	TitleColor         ConfigColor `json:"titleColor"`
	GraphicsColor      ConfigColor `json:"graphicsColor"`
	PrimaryTextColor   ConfigColor `json:"primaryTextColor"`
	SecondaryTextColor ConfigColor `json:"secondaryTextColor"`
	TertiaryTextColor  ConfigColor `json:"tertiaryTextColor"`

	ConsoleBackgroundColor          ConfigColor `json:"consoleBackgroundColor"`
	ConsoleResponseColor            ConfigColor `json:"consoleResponseColor"`
	ConsoleCommandColor             ConfigColor `json:"consoleCommandColor"`
	ConsoleErrorColor               ConfigColor `json:"consoleErrorColor"`
	ConsoleTimestampBackgroundColor ConfigColor `json:"consoleTimestampBackgroundColor"`
	ConsoleTimestampTextColor       ConfigColor `json:"consoleTimestampTextColor"`

	InputTextColor        ConfigColor `json:"inputTextColor"`
	InputBackgroundColor  ConfigColor `json:"inputBackgroundColor"`
	InputPromptColor      ConfigColor `json:"inputPromptColor"`
	InputPlaceholderColor ConfigColor `json:"inputPlaceholderColor"`

	AutocompleteBackgroundColor ConfigColor `json:"autocompleteMenuBackgroundColor"`
	AutocompleteTextColor       ConfigColor `json:"autocompleteTextColor"`
	AutocompleteHelpColor       ConfigColor `json:"autocompleteHelpColor"`
}

type Config struct {
	LogIncoming           bool        `json:"logIncoming"`
	TimestampFormat       string      `json:"timestampFormat"`
	ConsoleFilterPatterns []string    `json:"consoleFilterPatterns"`
	Theme                 ThemeConfig `json:"theme"`
}

var DefaultConfig = &Config{
	LogIncoming:     false,
	TimestampFormat: "hh:mm:ss",
	ConsoleFilterPatterns: []string{
		"^(?:ok\\s+)?(B|C|T\\d*):", // Temperature updates
	},
	Theme: ThemeConfig{
		BackgroundColor:    "default",
		BorderColor:        "white",
		TitleColor:         "white",
		GraphicsColor:      "white",
		PrimaryTextColor:   "white",
		SecondaryTextColor: "yellow",
		TertiaryTextColor:  "green",

		ConsoleBackgroundColor:          "default",
		ConsoleResponseColor:            "white",
		ConsoleCommandColor:             "yellow",
		ConsoleErrorColor:               "red",
		ConsoleTimestampBackgroundColor: "darkslategray",
		ConsoleTimestampTextColor:       "white",

		InputTextColor:              "white",
		InputBackgroundColor:        "darkblue",
		InputPromptColor:            "gold",
		InputPlaceholderColor:       "darkslategray",
		AutocompleteBackgroundColor: "white",
		AutocompleteTextColor:       "darkblue",
		AutocompleteHelpColor:       "red",
	},
}

var AppConfig = &Config{}

func (c *Config) Load() error {
	configPath := configdir.LocalConfig("clipper")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")
	if _, err = os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		DefaultConfig.Save()
	}
	defaultCfgBytes, _ := json.Marshal(DefaultConfig)
	json.Unmarshal(defaultCfgBytes, c)
	cfgBytes, err := os.ReadFile(configFile)
	err = json.Unmarshal(cfgBytes, c)
	return err
}

func (c *Config) Save() error {
	configPath := configdir.LocalConfig("clipper")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")
	cfgBytes, _ := json.MarshalIndent(*c, "", "  ")
	err = os.WriteFile(configFile, cfgBytes, 0644)
	return err
}

func (c *Config) getField(path string) (reflect.Value, error) {
	pathParts := strings.Split(path, ".")
	currentVal := reflect.ValueOf(c)

	for _, part := range pathParts {
		// Check if it's a pointer and get its element
		if currentVal.Kind() == reflect.Ptr {
			currentVal = currentVal.Elem()
		}

		// Ensure current value is a struct
		if currentVal.Kind() != reflect.Struct {
			return reflect.ValueOf(nil), fmt.Errorf("not a struct")
		}

		// Get the field by name
		currentVal = currentVal.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, part)
		})

		if !currentVal.IsValid() {
			return reflect.ValueOf(nil), fmt.Errorf("no such field: %s in obj", part)
		}

		if !currentVal.CanSet() {
			return reflect.ValueOf(nil), fmt.Errorf("cannot set field %s", part)
		}
	}

	if currentVal.Kind() == reflect.Ptr {
		currentVal = currentVal.Elem()
	}
	return currentVal, nil
}

func (c *Config) Set(path string, value interface{}) error {
	var err error
	var ok bool
	var field reflect.Value
	var val interface{}
	field, err = c.getField(path)
	switch field.Kind() {
	case reflect.String:
		switch field.Type() {
		case reflect.TypeOf(ConfigColor("")):
			val = ConfigColor(value.(string))
		default:
			val = value
		}
	case reflect.Bool:
		val, ok = value.(bool)
		if !ok {
			val, err = strconv.ParseBool(value.(string))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, ok = value.(int)
		if !ok {
			val, err = strconv.ParseInt(value.(string), 10, 64)
		}
	case reflect.Float32, reflect.Float64:
		val, ok = value.(float64)
		if !ok {
			val, err = strconv.ParseFloat(value.(string), 64)
		}
	}
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(val))
	return nil
}

func (c *Config) GetKeys() []string {
	keys := c.keyPaths("", reflect.TypeOf(*c), reflect.ValueOf(*c))
	sort.Strings(keys)
	return keys
}

func (c *Config) keyPaths(prefix string, t reflect.Type, v reflect.Value) []string {
	var paths []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)
		// If the field is a struct, recursively get its fields
		if fieldVal.Kind() == reflect.Struct {
			nestedPaths := c.keyPaths(joinPrefix(prefix, field.Name), fieldVal.Type(), fieldVal)
			paths = append(paths, nestedPaths...)
		} else {
			// Append the field path to the paths slice
			path := joinPrefix(prefix, field.Name)
			paths = append(paths, path)
		}
	}

	return paths
}

// joinPrefix joins the prefix and field name with a dot, handling empty prefixes
func joinPrefix(prefix, name string) string {
	name = strings.ToLower(name[:1]) + name[1:]
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func NewSettingsCompleter() cmdinput.StaticTokenCompleter {
	keys := AppConfig.GetKeys()
	reg := make(map[string]cmdinput.TokenCompleter, len(keys))
	for _, key := range keys {
		field, _ := AppConfig.getField(key)
		switch field.Kind() {
		case reflect.Bool:
			reg[key] = cmdinput.NewBoolTokenCompleter("value", nil)
		case reflect.String:
			switch field.Type() {
			case reflect.TypeOf(ConfigColor("")):
				reg[key] = cmdinput.NewColorTokenCompleter("value", nil)
			default:
				reg[key] = cmdinput.AnythingCompleter{"value"}

			}
		default:
			reg[key] = cmdinput.AnythingCompleter{"value"}
		}
	}

	completer := cmdinput.StaticTokenCompleter{
		ContextKey: "setting",
		Registry:   reg,
	}
	return completer
}
