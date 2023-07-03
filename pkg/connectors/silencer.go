package connectors

type ExternalSilencer interface {
	String() string

	Silence(labels map[string]string, id string) error
	Silenced(labels map[string]string) Silence

	Silences() []Silence
	SetSilence(id string, labels map[string]string)
	DeleteSilence(id string)
}

type Silence struct {
	ExternalId string
	URL        string
	Silenced   bool
	Labels     map[string]string
}
