package version

import (
	"fmt"
	"runtime"
)

// These are mostly set during compilation from linker flags
var (
	application = "tuwat"
	version     = "dev"
	revision    string
	branch      string
	releaseDate string
)

// VersionInfo is a rich representation of the version information which can also be readily serialized into a JSON representation
type VersionInfo struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Revision    string `json:"revision,omitempty"`
	Branch      string `json:"branch,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	GoVersion   string `json:"goVersion"`
	GoPlatform  string `json:"goPlatform"`
}

func (v VersionInfo) HumanReadable() string {
	return fmt.Sprintf("%s v%s (release date: %s)", v.Application, v.Version, v.ReleaseDate)
}

var Info VersionInfo

func init() {
	Info = VersionInfo{
		Application: application,
		Version:     version,
		Revision:    revision,
		Branch:      branch,
		ReleaseDate: releaseDate,
		GoVersion:   runtime.Version(),
		GoPlatform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
