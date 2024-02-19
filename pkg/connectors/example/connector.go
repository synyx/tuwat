package example

import (
	"context"
	"fmt"
	html "html/template"
	"math/rand"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
)

type Connector struct {
	config Config
}

type Config struct {
	Tag string
}

func NewConnector(cfg *Config) *Connector {
	return &Connector{*cfg}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

// Collect returns a few example warnings, one of each type
// This eliminates the need to select a source that contains a given error type during development
func (c *Connector) Collect(_ context.Context) ([]connectors.Alert, error) {

	var alerts []connectors.Alert

	alerts = append(alerts, exampleOkAlert())
	alerts = append(alerts, exampleWarningAlert())
	alerts = append(alerts, exampleCriticalAlert())
	alerts = append(alerts, exampleUnknownAlert())
	alerts = append(alerts, exampleRandomAlert())

	return alerts, nil
}

func exampleOkAlert() connectors.Alert {
	alert := connectors.Alert{
		Labels: map[string]string{
			"Project":    "Example Alerts",
			"Culprit":    "Geras",
			"EscalateTo": "BuJo",
		},
		Start:       time.Now().Add(-time.Hour * 1),
		State:       connectors.OK,
		Description: "OK",
		Details:     "Nothing to worry about, just informing you that nothing bad happened",
		Links: []html.HTML{
			html.HTML("<a href=\"https://go.dev/\" target=\"_blank\" alt=\"Home\">🏠</a>"),
		},
	}
	return alert
}
func exampleWarningAlert() connectors.Alert {
	alert := connectors.Alert{
		Labels: map[string]string{
			"Project":    "Example Alerts",
			"Culprit":    "Geras",
			"EscalateTo": "BuJo",
		},
		Start:       time.Now().Add(-time.Hour * 2),
		State:       connectors.Warning,
		Description: "Warning",
		Details:     "This is a warning! Coffee is low!",
		Links: []html.HTML{
			html.HTML("<a href=\"https://go.dev/\" target=\"_blank\" alt=\"Home\">🏠</a>"),
		},
	}
	return alert
}

func exampleCriticalAlert() connectors.Alert {
	alert := connectors.Alert{
		Labels: map[string]string{
			"Project":    "Example Alerts",
			"Culprit":    "Geras",
			"EscalateTo": "BuJo",
		},
		Start:       time.Now().Add(-time.Hour * 3),
		State:       connectors.Critical,
		Description: "Critical",
		Details:     "This is a critical alert! We ran out of coffee!",
		Links: []html.HTML{
			html.HTML("<a href=\"https://go.dev/\" target=\"_blank\" alt=\"Home\">🏠</a>"),
		},
	}
	return alert
}

func exampleUnknownAlert() connectors.Alert {
	alert := connectors.Alert{
		Labels: map[string]string{
			"Project":    "Example Alerts",
			"Culprit":    "Geras",
			"EscalateTo": "BuJo",
		},
		Start:       time.Now().Add(-time.Hour * 4),
		State:       connectors.Unknown,
		Description: "Unknown",
		Details:     "We have no idea what happened here",
		Links: []html.HTML{
			html.HTML("<a href=\"https://go.dev/\" target=\"_blank\" alt=\"Home\">🏠</a>"),
		},
	}
	return alert
}

func exampleRandomAlert() connectors.Alert {
	alert := connectors.Alert{
		Labels: map[string]string{
			"Project": "Random Alerts",
			"Random":  "true",
		},
		Start:       time.Now().Add(-time.Duration(rand.Int31n(100)) * time.Minute),
		State:       connectors.State(rand.Int31n(3)),
		Description: "Randomized State",
		Details:     "We really have no idea what happened here",
	}
	return alert
}

func (c *Connector) String() string {
	return fmt.Sprintf("Example Connector")
}
