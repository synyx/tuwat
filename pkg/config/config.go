package config

type Config struct {
	WebAddr     string
	Mode        string
	Environment string
	JaegerUrl   string
	Instance    string
}

func NewConfiguration() *Config {
	return &Config{
		WebAddr:     "127.0.0.0:8988",
		Mode:        "dev",
		Environment: "test",
		Instance:    "gonagdash",
	}
}
