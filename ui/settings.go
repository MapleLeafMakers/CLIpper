package ui

import (
	"clipper/ui/cmdinput"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MapleLeafMakers/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/kirsle/configdir"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type ConfigColor string

type ThemeConfig struct {
	BorderColor ConfigColor
}

type Config struct {
	LogIncoming     bool        `json:"logIncoming"`
	TimestampFormat string      `json:"timestampFormat"`
	Theme           ThemeConfig `json:"theme"`
}

var DefaultConfig = &Config{
	LogIncoming:     false,
	TimestampFormat: "hh:mm:ss",
	Theme: ThemeConfig{
		BorderColor: "#404040",
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
	cfgBytes, err := os.ReadFile(configFile)
	err = json.Unmarshal(cfgBytes, c)
	tview.Styles.BorderColor = tcell.GetColor(string(AppConfig.Theme.BorderColor))
	return err
}

func (c *Config) Save() error {
	configPath := configdir.LocalConfig("clipper")
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}
	configFile := filepath.Join(configPath, "config.json")
	log.Printf("Saving %#v", c)
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

func (c *Config) Set(path string, value string) error {
	updateTheme := strings.HasPrefix(path, "theme.")
	var err error
	var field reflect.Value
	var val interface{}
	field, err = c.getField(path)
	switch field.Kind() {
	case reflect.String:
		switch field.Type() {
		case reflect.TypeOf(ConfigColor("")):
			val = ConfigColor(value)
		default:
			val = value
		}
	case reflect.Bool:
		val, err = strconv.ParseBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err = strconv.ParseInt(value, 10, 64)
	case reflect.Float32, reflect.Float64:
		val, err = strconv.ParseFloat(value, 64)
	}
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(val))
	if updateTheme {
		log.Println("Updating Theme")
		tview.Styles.BorderColor = tcell.GetColor(string(AppConfig.Theme.BorderColor))
	}
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
	log.Println("Building settings completer")
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
		}
	}

	completer := cmdinput.StaticTokenCompleter{
		ContextKey: "setting",
		Registry:   reg,
	}
	log.Println("Built settings completer")
	return completer
}
