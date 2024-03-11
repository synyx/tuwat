package config

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/alertmanager"
	"github.com/synyx/tuwat/pkg/connectors/example"
	"github.com/synyx/tuwat/pkg/connectors/github"
	"github.com/synyx/tuwat/pkg/connectors/gitlabmr"
	"github.com/synyx/tuwat/pkg/connectors/icinga2"
	"github.com/synyx/tuwat/pkg/connectors/nagiosapi"
	"github.com/synyx/tuwat/pkg/connectors/orderview"
	"github.com/synyx/tuwat/pkg/connectors/patchman"
	"github.com/synyx/tuwat/pkg/connectors/redmine"
)

var fVersion = flag.Bool("version", false, "Print version")
var fInstance = flag.String("instance", "0", "Running instance identifier")
var fEnvironment = flag.String("environment", "test", "(test, stage, prod)")
var fAddr = flag.String("addr", "0.0.0.0:8988", "Bind web application port")
var fMgmtAddr = flag.String("mgmtAddr", "127.0.0.1:8987", "Bind management port")
var fConfigFile = flag.String("conf", "/etc/tuwat.toml", "Configuration file")
var fDashboardDir = flag.String("dashboards", "/etc/tuwat.d", "Dashboard Configuration Directory")
var fOtelUrl = flag.String("otelUrl", "", "OTEL tracing endpoint URL")

type Config struct {
	WebAddr        string
	ManagementAddr string
	Environment    string
	OtelUrl        string
	Instance       string
	PrintVersion   bool
	Logger         zap.Config
	Connectors     []connectors.Connector
	WhereTemplate  *template.Template
	Interval       time.Duration
	Dashboards     map[string]*Dashboard
	Style          string
}

type Dashboard struct {
	Name   string
	Mode   DashboardMode
	Filter []Rule
}

type Rule struct {
	Description string
	What        RuleMatcher
	When        RuleMatcher
	Labels      map[string]RuleMatcher
}

type mainConfig struct {
	WhereTemplate string `toml:"where"`
	Interval      string `toml:"interval"`
	Style         string `toml:"style"`
}

type mainDashboardConfig struct {
	Mode DashboardMode `toml:"mode"`
}

type dashboardConfig struct {
	Main  mainDashboardConfig      `toml:"main"`
	Rules []map[string]interface{} `toml:"rule"`
}

type rootConfig struct {
	Main          mainConfig               `toml:"main"`
	Logger        zap.Config               `toml:"logger"`
	Rules         []map[string]interface{} `toml:"rule"`
	Alertmanagers []alertmanager.Config    `toml:"alertmanager"`
	GitlabMRs     []gitlabmr.Config        `toml:"gitlabmr"`
	Icinga2s      []icinga2.Config         `toml:"icinga2"`
	NagiosAPIs    []nagiosapi.Config       `toml:"nagiosapi"`
	Patchmans     []patchman.Config        `toml:"patchman"`
	GitHubIssues  []github.Config          `toml:"github"`
	Redmines      []redmine.Config         `toml:"redmine"`
	Orderview     []orderview.Config       `toml:"orderview"`
	Example       []example.Config         `toml:"example"`
}

func NewConfiguration() (config *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("panicked loading config: %w", errors.New(x))
			case error:
				err = fmt.Errorf("panicked loading config: %w", x)
			default:
				err = fmt.Errorf("panicked loading config: %w", errors.New(fmt.Sprint(x)))
			}
		}
	}()

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

	if value, ok := os.LookupEnv("TUWAT_OTEL_URL"); ok {
		cfg.OtelUrl = value
	} else {
		cfg.OtelUrl = *fOtelUrl
	}

	rootConfig := cfg.defaultConfiguration()

	if err := cfg.loadConfigFile(*fConfigFile, &rootConfig); errors.Is(err, os.ErrNotExist) {
		// ignore missing configuration and start with defaults
		err = nil
	} else if err != nil {
		return nil, err
	}

	if err := cfg.configureMain(&rootConfig); err != nil {
		return nil, err
	}

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

	// validate configuration
	if cfg.Style != "" && !slices.Contains([]string{"light", "dark"}, cfg.Style) {
		return nil, errors.New("configuration error: [main] style must be \"light\" or \"dark\"")
	}

	return cfg, err
}

func (cfg *Config) defaultConfiguration() rootConfig {
	var rootConfig rootConfig

	// Defaults for configuration
	rootConfig.Main.WhereTemplate = `{{with index .Labels "Cluster"}}{{.}}/{{end}}{{first .Labels "Project" "Namespace" "Hostname" "job" "cluster"}}`

	if cfg.Environment == "prod" {
		rootConfig.Logger = zap.NewProductionConfig()
	} else {
		rootConfig.Logger = zap.NewDevelopmentConfig()
	}

	rootConfig.Main.Interval = "1m"
	rootConfig.Main.Style = "dark"

	return rootConfig
}

func (cfg *Config) loadConfigFile(file string, rootConfig *rootConfig) error {
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("configuration file %s unreadable: %w", file, err)
	}

	// Fill configuration
	if _, err := toml.DecodeFile(file, &rootConfig); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) configureMain(rootConfig *rootConfig) (err error) {
	cfg.Style = rootConfig.Main.Style

	cfg.Logger = rootConfig.Logger

	// Add connectors
	for _, connectorConfig := range rootConfig.Alertmanagers {
		cfg.Connectors = append(cfg.Connectors, alertmanager.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.GitlabMRs {
		cfg.Connectors = append(cfg.Connectors, gitlabmr.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.Icinga2s {
		cfg.Connectors = append(cfg.Connectors, icinga2.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.NagiosAPIs {
		cfg.Connectors = append(cfg.Connectors, nagiosapi.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.Patchmans {
		cfg.Connectors = append(cfg.Connectors, patchman.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.GitHubIssues {
		cfg.Connectors = append(cfg.Connectors, github.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.Redmines {
		cfg.Connectors = append(cfg.Connectors, redmine.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.Orderview {
		cfg.Connectors = append(cfg.Connectors, orderview.NewConnector(&connectorConfig))
	}
	for _, connectorConfig := range rootConfig.Example {
		cfg.Connectors = append(cfg.Connectors, example.NewConnector(&connectorConfig))
	}

	// Add template for
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
		Parse(rootConfig.Main.WhereTemplate)
	if err != nil {
		return err
	}

	if cfg.Interval, err = time.ParseDuration(rootConfig.Main.Interval); err != nil {
		return err
	}

	// Add default dashboard, containing potentially all unfiltered alerts
	cfg.Dashboards = make(map[string]*Dashboard)
	var dashboard Dashboard
	for _, r := range rootConfig.Rules {
		dashboard.Filter = append(dashboard.Filter, parseRule(r))
	}
	cfg.Dashboards[""] = &dashboard

	return err
}

func (cfg *Config) loadDashboardConfig(file string) error {

	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("configuration file %s unreadable: %w", file, err)
	}

	var dashboard Dashboard
	var dashboardConfig dashboardConfig
	_, err := toml.DecodeFile(file, &dashboardConfig)
	if err != nil {
		panic(err)
	}

	// Excluding is the default mode, this mirrors a mindset of "everything
	// new has to be looked at, at least once".
	// `0` is the empty value, so in case Main.Mode is unset, it will still
	// be the default.
	dashboard.Mode = Excluding
	for _, r := range dashboardConfig.Rules {
		dashboard.Mode = dashboardConfig.Main.Mode
		dashboard.Filter = append(dashboard.Filter, parseRule(r))
	}

	name := filepath.Base(file)
	name = strings.TrimSuffix(name, ".toml")
	dashboard.Name = name
	cfg.Dashboards[name] = &dashboard

	return err
}

func parseRule(r map[string]interface{}) Rule {
	labels := make(map[string]RuleMatcher)
	if labelFilters, ok := r["label"]; ok {
		for n, l := range labelFilters.(map[string]interface{}) {
			labels[n] = ParseRuleMatcher(l.(string))
		}
	}
	var what RuleMatcher
	if w, ok := r["what"]; ok {
		what = ParseRuleMatcher(w.(string))
	}
	var when RuleMatcher
	if w, ok := r["when"]; ok {
		when = ParseRuleMatcher(w.(string))
	}

	br := Rule{
		Description: r["description"].(string),
		What:        what,
		When:        when,
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
