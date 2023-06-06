package common

import (
	"golang.org/x/oauth2/clientcredentials"
)

type HTTPConfig struct {
	URL string

	// SSL
	Insecure bool

	// Basic Auth
	Username string
	Password string

	// OAuth2
	OAuth2Creds clientcredentials.Config

	// Bearer Token
	BearerToken string
}
