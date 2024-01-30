package ui

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	err := AppConfig.Load()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(AppConfig, DefaultConfig) {
		t.Error("Didn't load default", AppConfig, DefaultConfig)
	}

	AppConfig.Theme.BorderColor = "#ff0000"

	if AppConfig.Theme.BorderColor != "#ff0000" {
		t.Error("Didn't set nested")
	}
}

func TestConfig_Save(t *testing.T) {
	err := AppConfig.Load()
	if err != nil {
		t.Error(err)
	}
	err = AppConfig.Save()
	if err != nil {
		t.Error(err)
	}

}

func TestConfig_Set(t *testing.T) {
	err := AppConfig.Load()
	if err != nil {
		t.Error(err)
	}
	AppConfig.Set("logIncoming", "true")
	AppConfig.Set("theme.borderColor", "#00FF00")
	if AppConfig.Theme.BorderColor != "#00FF00" {
		t.Errorf("Nested update didn't .Set() %#v", AppConfig)
	}
	if !AppConfig.LogIncoming {
		t.Errorf("Top level bool didn't .Set() %#v", AppConfig)
	}
}

func TestConfig_Load_NewDefaults(t *testing.T) {
	c := Config{LogIncoming: true, TimestampFormat: "hh:mma"}
	json.Unmarshal([]byte("{\"logIncoming\": false}"), &c)
	if c.TimestampFormat != "hh:mma" {
		t.Errorf("TimestampFormat wasn't preserved")
	}
	if c.LogIncoming {
		t.Errorf("LogIncoming didn't update")
	}

}
