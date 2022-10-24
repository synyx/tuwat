package nagiosapi

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
		"host":    alert.Labels["Hostname"],
		"service": alert.Description,
		"comment": fmt.Sprintf("%s: silenced via %s", user, buildinfo.Service),
		"author":  user,
		"expire":  duration / time.Minute,
	}
	endpoint := "/acknowledge_problem"

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

	decoder := json.NewDecoder(res.Body)

	var response response
	err = decoder.Decode(&response)
	if err != nil {
		return err
	}
	if !response.Success {
		// TODO(jo): the `content` map is overloaded with an error string
	}

	return nil
}
