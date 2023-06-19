package config

import (
	"bytes"
	"errors"

	"github.com/BurntSushi/toml"
)

const (
	// Excluding is the default mode, this mirrors a mindset of "everything
	// new has to be looked at, at least once".
	// `0` is the empty value, so in case it is used in a context with
	// uninitialized data, it will still be the default.
	Excluding DashboardMode = iota
	Including
)

type DashboardMode uint8

func (s DashboardMode) String() string {
	switch s {
	case Including:
		return "including"
	case Excluding:
		return "excluding"
	default:
		return "unknown"
	}
}

func (s DashboardMode) MarshalTOML() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	encoder := toml.NewEncoder(buf)
	if err := encoder.Encode(s.String()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *DashboardMode) UnmarshalTOML(data interface{}) (err error) {
	switch data.(string) {
	case "including":
		fallthrough
	case "include":
		*s = Including
		return nil
	case "excluding":
		fallthrough
	case "exclude":
		*s = Excluding
		return nil
	default:
		return errors.New("unknown dashboard mode: " + data.(string))
	}
}
