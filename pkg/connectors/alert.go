package connectors

import (
	html "html/template"
	"time"
)

type Alert struct {
	Labels      map[string]string
	Start       time.Time
	State       State
	Description string
	Details     string
	Links       []html.HTML
	Silence     SilencerFunc
}
