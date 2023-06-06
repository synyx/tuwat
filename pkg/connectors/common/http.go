package common

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/synyx/tuwat/pkg/version"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
)

// Client prepares a http client for a given configuration.
func (cfg *HTTPConfig) Client() *http.Client {

	// Create transport per connector, to only configure insecure SSL where needed.
	var tr http.RoundTripper = http.DefaultTransport.(*http.Transport).Clone()
	tr.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: cfg.Insecure}

	tr = &userAgentRoundTripper{rt: tr}

	// Use oauth2 if configured
	if cfg.OAuth2Creds.ClientID != "" {
		// This context is only used to get the Base transport within the oauth2 code.
		// This is promptly overridden by the transport needed here.
		ctx := context.Background()

		oauth2Transport := cfg.OAuth2Creds.Client(ctx).Transport
		oauth2Transport.(*oauth2.Transport).Base = tr
		tr = oauth2Transport
	}

	// Use bearer token if configured
	if cfg.BearerToken != "" {
		tr = &bearerAuthRoundTripper{cfg.BearerToken, tr}
	}

	// Use basic auth if configured
	if cfg.Username != "" {
		tr = &basicAuthRoundTripper{cfg.Username, cfg.Password, tr}
	}

	// Add tracing to the client connections
	tr = otelhttp.NewTransport(tr)

	client := &http.Client{Transport: tr}

	return client
}

type bearerAuthRoundTripper struct {
	bearer string
	rt     http.RoundTripper
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 {
		return rt.rt.RoundTrip(req)
	}

	req = cloneRequest(req)
	token := rt.bearer
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return rt.rt.RoundTrip(req)
}

type basicAuthRoundTripper struct {
	username string
	password string
	rt       http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)
	req.SetBasicAuth(rt.username, rt.password)

	return rt.rt.RoundTrip(req)
}

type userAgentRoundTripper struct {
	rt http.RoundTripper
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = cloneRequest(req)

	req.Header.Set("User-Agent", version.Info.Application+"/"+version.Info.Version)

	return rt.rt.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
