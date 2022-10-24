package icinga2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/synyx/tuwat/pkg/buildinfo"
	"github.com/synyx/tuwat/pkg/connectors"
)

func (c *Connector) createSilencer(alert connectors.Alert) connectors.SilencerFunc {

	return func(ctx context.Context, duration time.Duration, user string) error {

		return c.Silence(ctx, alert, duration, user)
	}
}

func (c *Connector) Silence(ctx context.Context, alert connectors.Alert, duration time.Duration, user string) error {
	payload := map[string]interface{}{
		"type":          alert.Labels["type"],
		"comment":       fmt.Sprintf("%s: silenced via %s", user, buildinfo.Service),
		"author":        user,
		"child_options": 1,
		"start_time":    time.Now().Unix(),
		"expiry":        time.Now().Add(duration).Unix(),
	}
	endpoint := "/v1/actions/acknowledge-problem"

	switch alert.Labels["type"] {
	case "Service":
		payload["filter"] = `host.name=="` + alert.Labels["Hostname"] + `" && service.name=="` + alert.Details + `"`
	case "Host":
		payload["filter"] = `host.name=="` + alert.Labels["Hostname"] + `"`
	}

	err := c.post(ctx, endpoint, payload)

	return err
}

func (c *Connector) post(ctx context.Context, endpoint string, content map[string]interface{}) error {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(content); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL+endpoint, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
