package connectors

type ExternalSilencer interface {
	Silence(labels map[string]string, externalId string) error
	Silenced(labels map[string]string) Silence
	String() string
}

type Silence struct {
	ExternalId string
	URL        string
	Silenced   bool
}
