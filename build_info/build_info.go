package build_info

import (
	"github.com/bykof/gostradamus"
)

type BuildInfo struct {
	BuildTime     gostradamus.DateTime
	VersionString string
	CommitHash    string
	BuildArch     string
	BuildOS       string
}
