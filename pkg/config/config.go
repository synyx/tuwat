package config

import (
	"flag"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/synyx/gonagdash/pkg/connectors/alertmanager"
	"github.com/synyx/gonagdash/pkg/connectors/gitlabmr"
	"github.com/synyx/gonagdash/pkg/connectors/icinga2"
	"github.com/synyx/gonagdash/pkg/connectors/nagiosapi"
	"github.com/synyx/gonagdash/pkg/connectors/patchman"
)

var fVersion = flag.Bool("version", false, "Print version")
var fInstance = flag.String("instance", "0", "Running instance identifier")
var fMode = flag.String("mode", "dev", "Mode to use (dev, prod)")
var fEnvironment = flag.String("environment", "test", "(test, stage, prod)")
var fAddr = flag.String("addr", "127.0.0.1:8988", "Bind web application to port")
var fConfigFile = flag.String("conf", "/etc/gonagdash.toml", "Configuration file")

type Config struct {
	WebAddr       string
	Mode          string
	Environment   string
	JaegerUrl     string
	Instance      string
	PrintVersion  bool
	Connectors    []connectors.Connector
	WhereTemplate *template.Template
}
type MainConfig struct {
	WhereTemplate string
}
type ConnectorConfig struct {
	Main          MainConfig            `json:"main"`
	Alertmanagers []alertmanager.Config `toml:"alertmanager"`
	GitlabMRs     []gitlabmr.Config     `toml:"gitlabmr"`
	Icinga2s      []icinga2.Config      `toml:"icinga2"`
	NagiosAPIs    []nagiosapi.Config    `toml:"nagiosapi"`
	Patchmans     []patchman.Config     `toml:"patchman"`
}

func NewConfiguration() (*Config, error) {

	flag.Parse()

	cfg := &Config{
		PrintVersion: *fVersion,
	}

	if value, ok := os.LookupEnv("GND_MODE"); ok {
		cfg.Mode = value
	} else {
		cfg.Mode = *fMode
	}

	if value, ok := os.LookupEnv("GND_ENVIRONMENT"); ok {
		cfg.Environment = value
	} else {
		cfg.Environment = *fEnvironment
	}

	if value, ok := os.LookupEnv("GND_INSTANCE"); ok {
		cfg.Instance = value
	} else if *fInstance != "" {
		cfg.Instance = *fInstance
	} else {
		cfg.Instance = getHostname()
	}

	if value, ok := os.LookupEnv("GND_ADDR"); ok {
		cfg.WebAddr = value
	} else {
		cfg.WebAddr = *fAddr
	}

	err := cfg.loadFile(*fConfigFile)

	return cfg, err
}

func (cfg *Config) loadFile(file string) error {

	var connectorConfigs ConnectorConfig
	_, err := toml.DecodeFile(file, &connectorConfigs)
	if err != nil {
		panic(err)
	}

	for _, connectorConfig := range connectorConfigs.Alertmanagers {
		cfg.Connectors = append(cfg.Connectors, alertmanager.NewCollector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.GitlabMRs {
		cfg.Connectors = append(cfg.Connectors, gitlabmr.NewCollector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Icinga2s {
		cfg.Connectors = append(cfg.Connectors, icinga2.NewCollector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.NagiosAPIs {
		cfg.Connectors = append(cfg.Connectors, nagiosapi.NewCollector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Patchmans {
		cfg.Connectors = append(cfg.Connectors, patchman.NewCollector(connectorConfig))
	}

	cfg.WhereTemplate, err = template.New("where").
		Funcs(map[string]any{
			"first": func(m map[string]string, x ...string) string {
				for _, y := range x {
					if z, ok := m[y]; ok {
						return z
					}
				}
				return "NOT_FOUND"
			},
		}).
		Parse(connectorConfigs.Main.WhereTemplate)

	return err
}

func getHostname() string {
	if e, ok := os.LookupEnv("HOSTNAME"); ok {
		if e == "" {
			return "unknown"
		}
		return e
	}
	return "unknown"
}
