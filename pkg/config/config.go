package config

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/alertmanager"
	"github.com/synyx/tuwat/pkg/connectors/github"
	"github.com/synyx/tuwat/pkg/connectors/gitlabmr"
	"github.com/synyx/tuwat/pkg/connectors/icinga2"
	"github.com/synyx/tuwat/pkg/connectors/nagiosapi"
	"github.com/synyx/tuwat/pkg/connectors/patchman"
)

var fVersion = flag.Bool("version", false, "Print version")
var fInstance = flag.String("instance", "0", "Running instance identifier")
var fEnvironment = flag.String("environment", "test", "(test, stage, prod)")
var fAddr = flag.String("addr", "127.0.0.1:8988", "Bind web application port")
var fMgmtAddr = flag.String("mgmtAddr", "127.0.0.1:8987", "Bind management port")
var fConfigFile = flag.String("conf", "/etc/tuwat.toml", "Configuration file")
var fDashboardDir = flag.String("dashboards", "/etc/tuwat.d", "Dashboard Configuration Directory")

type Config struct {
	WebAddr        string
	ManagementAddr string
	Environment    string
	JaegerUrl      string
	Instance       string
	PrintVersion   bool
	Connectors     []connectors.Connector
	WhereTemplate  *template.Template
	Interval       time.Duration
	Dashboards     map[string]*Dashboard
}

type MainConfig struct {
	WhereTemplate string `toml:"where"`
	Interval      string `toml:"interval"`
}

type Dashboard struct {
	Name   string
	Filter []Rule
}

type Rule struct {
	Description string                    `toml:"description"`
	What        *regexp.Regexp            `toml:"what"`
	Labels      map[string]*regexp.Regexp `toml:"labels"`
}

type DashboardConfig struct {
	Rules []map[string]interface{} `toml:"rule"`
}

type ConnectorConfig struct {
	Main          MainConfig               `toml:"main"`
	Rules         []map[string]interface{} `toml:"rule"`
	Alertmanagers []alertmanager.Config    `toml:"alertmanager"`
	GitlabMRs     []gitlabmr.Config        `toml:"gitlabmr"`
	Icinga2s      []icinga2.Config         `toml:"icinga2"`
	NagiosAPIs    []nagiosapi.Config       `toml:"nagiosapi"`
	Patchmans     []patchman.Config        `toml:"patchman"`
	GitHubIssues  []github.Config          `toml:"github"`
}

func NewConfiguration() (*Config, error) {

	flag.Parse()

	cfg := &Config{
		PrintVersion: *fVersion,
	}

	if value, ok := os.LookupEnv("TUWAT_ENVIRONMENT"); ok {
		cfg.Environment = value
	} else {
		cfg.Environment = *fEnvironment
	}

	if value, ok := os.LookupEnv("TUWAT_INSTANCE"); ok {
		cfg.Instance = value
	} else if *fInstance != "" {
		cfg.Instance = *fInstance
	} else {
		cfg.Instance = getHostname()
	}

	if value, ok := os.LookupEnv("TUWAT_ADDR"); ok {
		cfg.WebAddr = value
	} else {
		cfg.WebAddr = *fAddr
	}

	if value, ok := os.LookupEnv("TUWAT_MANAGEMENT_ADDR"); ok {
		cfg.ManagementAddr = value
	} else {
		cfg.ManagementAddr = *fMgmtAddr
	}

	if value, ok := os.LookupEnv("TUWAT_CONF"); ok {
		*fConfigFile = value
	}

	if value, ok := os.LookupEnv("TUWAT_DASHBOARD_DIR"); ok {
		*fDashboardDir = value
	}

	err := cfg.loadMainConfig(*fConfigFile)

	err = filepath.WalkDir(*fDashboardDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Ext(path) == ".toml" {
			return cfg.loadDashboardConfig(path)
		} else {
			return err
		}
	})
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}

	return cfg, err
}

func (cfg *Config) loadMainConfig(file string) error {

	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("configuration file %s unreadable: %w", file, err)
	}

	var connectorConfigs ConnectorConfig
	_, err := toml.DecodeFile(file, &connectorConfigs)
	if err != nil {
		panic(err)
	}

	for _, connectorConfig := range connectorConfigs.Alertmanagers {
		cfg.Connectors = append(cfg.Connectors, alertmanager.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.GitlabMRs {
		cfg.Connectors = append(cfg.Connectors, gitlabmr.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Icinga2s {
		cfg.Connectors = append(cfg.Connectors, icinga2.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.NagiosAPIs {
		cfg.Connectors = append(cfg.Connectors, nagiosapi.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.Patchmans {
		cfg.Connectors = append(cfg.Connectors, patchman.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range connectorConfigs.GitHubIssues {
		cfg.Connectors = append(cfg.Connectors, github.NewConnector(&connectorConfig))
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

	cfg.Dashboards = make(map[string]*Dashboard)
	var dashboard Dashboard
	for _, r := range connectorConfigs.Rules {
		dashboard.Filter = append(dashboard.Filter, parseRule(r))
	}

	return err
}

func (cfg *Config) loadDashboardConfig(file string) error {

	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("configuration file %s unreadable: %w", file, err)
	}

	var dashboard Dashboard
	var dashboardConfig DashboardConfig
	_, err := toml.DecodeFile(file, &dashboardConfig)
	if err != nil {
		panic(err)
	}

	for _, r := range dashboardConfig.Rules {
		dashboard.Filter = append(dashboard.Filter, parseRule(r))
	}

	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".toml")
	dashboard.Name = name
	cfg.Dashboards[name] = &dashboard

	return err
}

func parseRule(r map[string]interface{}) Rule {
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
	return br
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
