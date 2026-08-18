package version

import (
	"fmt"
	"path"
	"runtime"
	"runtime/debug"
)

// These are mostly set during compilation from linker flags
var (
	application = "tuwat"
	version     = "dev"
	revision    string
	branch      string
	releaseDate string
)

// ApplicationInfo is a rich representation of the version information which can also be readily serialized into a JSON representation
type ApplicationInfo struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Revision    string `json:"revision,omitempty"`
	Branch      string `json:"branch,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	GoVersion   string `json:"goVersion"`
	GoPlatform  string `json:"goPlatform"`
}

// HumanReadable prepares a string for human-readable version information, mostly for the applications' `-version`
func (v *ApplicationInfo) HumanReadable() string {
	// in-development
	if v.Version == "dev" && v.Branch != "" {
		return fmt.Sprintf("%s %s %s (%s, build date: %s)", v.Application, v.Version, v.Revision, v.Branch, v.ReleaseDate)
	}
	if v.Version == "dev" && v.Branch == "" {
		return fmt.Sprintf("%s %s %s (build date: %s)", v.Application, v.Version, v.Revision, v.ReleaseDate)
	}

	if v.ReleaseDate != "" {
		return fmt.Sprintf("%s %s (release date: %s)", v.Application, v.Version, v.ReleaseDate)
	}

	return fmt.Sprintf("%s %s", v.Application, v.Version)
}

var Info ApplicationInfo

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		version = pick(version, info.Main.Version)
		application = pick(application, path.Base(info.Main.Path))
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = pick(revision, setting.Value)
			case "vcs.time":
				releaseDate = pick(releaseDate, setting.Value)
			}
		}
	}

	Info = ApplicationInfo{
		Application: application,
		Version:     version,
		Revision:    revision,
		Branch:      branch,
		ReleaseDate: releaseDate,
		GoVersion:   runtime.Version(),
		GoPlatform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// pick picks the first NonZero value
func pick(primary, fallback string) string {
	if primary == "" {
		return fallback
	}

	return primary
}
