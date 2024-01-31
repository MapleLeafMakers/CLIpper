package build_info

import (
	"encoding/json"
	"errors"
	"github.com/bykof/gostradamus"
	"golang.org/x/mod/semver"
	"io"
	"net/http"
)

type BuildInfo struct {
	BuildTime     gostradamus.DateTime
	VersionString string
	CommitHash    string
	BuildArch     string
	BuildOS       string
}

func getLatestVersion() (string, error) {
	url := "https://api.github.com/repos/MapleLeafMakers/CLIpper/releases/latest"
	httpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CLIpper")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-Github-Api-Version", "2022-11-28")
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return "", getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}
	var response map[string]interface{}
	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		return "", jsonErr
	}
	return response["tag_name"].(string), nil
}

func CheckForUpdates(currentVersion string, callback func(bool, string, string, error)) {
	if !semver.IsValid(currentVersion) {
		callback(false, "", "", errors.New(currentVersion+" is not a valid version number"))
		return
	}

	latest, err := getLatestVersion()
	if err != nil {
		callback(false, "", "", err)
		return
	}

	if !semver.IsValid(latest) {
		callback(false, "", "", errors.New(latest+" is not a valid version number"))
		return
	}

	if semver.Compare(latest, currentVersion) == 1 {
		callback(true, latest, "https://github.com/MapleLeafMakers/CLIpper/releases/latest", nil)
	} else {
		callback(false, latest, "https://github.com/MapleLeafMakers/CLIpper/releases/latest", nil)
	}
}
