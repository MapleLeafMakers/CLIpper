package ui

type ServerInfo struct {
	KlippyConnected  bool     `json:"klippy_connected"`
	KlippyState      string   `json:"klippy_state"`
	Components       []string `json:"components"`
	FailedComponents []string `json:"failed_components"`
	Warnings         []string `json:"warnings"`
	WebsocketCount   int      `json:"websocket_count"`
	MoonrakerVersion string   `json:"moonraker_version"`
	APIVersion       []int    `json:"api_version"`
	APIVersionString string   `json:"api_version_string"`
}
