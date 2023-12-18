package connectors

import "context"

type ExternalSilencer interface {
	String() string

	Silences() []Silence
	Silenced(labels map[string]string) Silence
	SetSilence(id string, labels map[string]string)
	DeleteSilence(id string)
	Refresh(ctx context.Context) error
}

type Silence struct {
	ExternalId string
	URL        string
	Silenced   bool
	Labels     map[string]string
}
