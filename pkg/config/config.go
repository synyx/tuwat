package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/alertmanager"
	"github.com/synyx/tuwat/pkg/connectors/gitlabmr"
	"github.com/synyx/tuwat/pkg/connectors/icinga2"
	"github.com/synyx/tuwat/pkg/connectors/nagiosapi"
	"github.com/synyx/tuwat/pkg/connectors/patchman"
)

var fVersion = flag.Bool("version", false, "Print version")
var fInstance = flag.String("instance", "0", "Running instance identifier")
var fEnvironment = flag.String("environment", "test", "(test, stage, prod)")
var fAddr = flag.String("addr", "127.0.0.1:8988", "Bind web application to port")
var fConfigFile = flag.String("conf", "/etc/tuwat.toml", "Configuration file")

type Config struct {
	WebAddr       string
	Environment   string
	JaegerUrl     string
	Instance      string
	PrintVersion  bool
	Connectors    []connectors.Connector
	WhereTemplate *template.Template
	Interval      time.Duration
	Filter        []Rule
}

type MainConfig struct {
	WhereTemplate string `toml:"where"`
	Interval      string `toml:"interval"`
}

type Rule struct {
	Description string                    `toml:"description"`
	What        *regexp.Regexp            `toml:"what"`
	Labels      map[string]*regexp.Regexp `toml:"labels"`
}

type ConnectorConfig struct {
	Main          MainConfig               `toml:"main"`
	Rules         []map[string]interface{} `toml:"rule"`
	Alertmanagers []alertmanager.Config    `toml:"alertmanager"`
	GitlabMRs     []gitlabmr.Config        `toml:"gitlabmr"`
	Icinga2s      []icinga2.Config         `toml:"icinga2"`
	NagiosAPIs    []nagiosapi.Config       `toml:"nagiosapi"`
	Patchmans     []patchman.Config        `toml:"patchman"`
}

func NewConfiguration() (*Config, error) {

	flag.Parse()

	cfg := &Config{
		PrintVersion: *fVersion,
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

	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("configuration file %s unreadable: %w", file, err)
	}

	var connectorConfigs ConnectorConfig
	_, err := toml.DecodeFile(file, &connectorConfigs)
	if err != nil {
		panic(err)
	}

	for _, connectorConfig := range connectorConfigs.Alertmanagers {
		cfg.Connectors = append(cfg.Connectors, alertmanager.NewConnector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.GitlabMRs {
		cfg.Connectors = append(cfg.Connectors, gitlabmr.NewConnector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Icinga2s {
		cfg.Connectors = append(cfg.Connectors, icinga2.NewConnector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.NagiosAPIs {
		cfg.Connectors = append(cfg.Connectors, nagiosapi.NewConnector(connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Patchmans {
		cfg.Connectors = append(cfg.Connectors, patchman.NewConnector(connectorConfig))
	}

	whereTemplate := connectorConfigs.Main.WhereTemplate
	if whereTemplate == "" {
		whereTemplate = `{{with index .Labels "Cluster"}}{{.}}/{{end}}{{first .Labels "Project" "Namespace" "Hostname" "job" "cluster"}}`
	}

	cfg.WhereTemplate, err = template.New("where").
		Funcs(map[string]any{
			"first": func(m map[string]string, x ...string) string {
				for _, y := range x {
					if z, ok := m[y]; ok && z != "" {
						return z
					}
				}
				return "NOT_FOUND"
			},
		}).
		Parse(connectorConfigs.Main.WhereTemplate)
	if err != nil {
		return err
	}

	if connectorConfigs.Main.Interval != "" {
		if cfg.Interval, err = time.ParseDuration(connectorConfigs.Main.Interval); err != nil {
			return err
		}
	} else {
		cfg.Interval = 1 * time.Minute
	}

	for _, r := range connectorConfigs.Rules {
		labels := make(map[string]*regexp.Regexp)
		if labelFilters, ok := r["label"]; ok {
			for n, l := range labelFilters.(map[string]interface{}) {
				labels[n] = regexp.MustCompile(l.(string))
			}
		}
		var what *regexp.Regexp
		if w, ok := r["what"]; ok {
			what = regexp.MustCompile(w.(string))
		}

		br := Rule{
			Description: r["description"].(string),
			What:        what,
			Labels:      labels,
		}
		cfg.Filter = append(cfg.Filter, br)
	}

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
