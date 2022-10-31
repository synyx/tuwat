package version

import (
	"bytes"
	"runtime"
	"strings"
	"text/template"
)

// These are mostly set during compilation from linker flags
var (
	application = "tuwat"
	version     = "dev"
	revision    string
	branch      string
	buildDate   string
)

// A rich representation of the version information which can also be readily serialized into a JSON representation
type VersionInfo struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Revision    string `json:"revision,omitempty"`
	Branch      string `json:"branch,omitempty"`
	BuildDate   string `json:"buildDate,omitempty"`
	GoVersion   string `json:"goVersion"`
	GoPlatform  string `json:"goPlatform"`
}

var versionInfoTmpl = `
{{.Application}}, version {{.Version}} (branch: {{.Branch}}, revision: {{.Revision}})
  build date:       {{.BuildDate}}
  go version:       {{.GoVersion}}
  platform:         {{.GoPlatform}}
`

func (v VersionInfo) Print() string {
	t := template.Must(template.New("version").Parse(versionInfoTmpl))
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", v); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

var Info VersionInfo

func init() {
	Info = VersionInfo{
		Application: application,
		Version:     version,
		Revision:    revision,
		Branch:      branch,
		BuildDate:   buildDate,
		GoVersion:   runtime.Version(),
		GoPlatform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}
