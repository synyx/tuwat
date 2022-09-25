package connectors

type HTTPConfig struct {
	URL string

	// SSL
	Insecure bool

	// Basic Auth
	Username string
	Password string

	// OAuth2
	ClientId     string
	ClientSecret string
	TokenURL     string

	// Bearer Token
	BearerToken string
}
