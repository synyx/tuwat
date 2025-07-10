package config

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestParsingWhenWhatLabelRule(t *testing.T) {
	expected := Rule{
		Description: "Non-Escalated",
		When:        ParseRuleMatcher("< 86400"), // < 2d
		What:        ParseRuleMatcher(": Update"),
		Labels: map[string]RuleMatcher{
			"Type": ParseRuleMatcher("PullRequest"),
		},
	}

	cfg, err := config(updateToml)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Dashboards) != 1 {
		t.Fatal("Expected default dashboard")
	}

	filters := cfg.Dashboards[""].Filter
	if len(filters) != 1 {
		t.Fatal("Expected 1 filter")
	}
	filter := filters[0]

	if filter.Description != expected.Description {
		t.Errorf("Expected filter description to be %s, got %s", expected.Description, filter.Description)
	}
	if filter.What.String() != expected.What.String() {
		t.Errorf("Expected filter what to be %s, got %s", expected.What, filter.What)
	}
	if filter.When.String() != expected.When.String() {
		t.Errorf("Expected filter when to be %s, got %s", expected.When, filter.When)
	}
	if filter.Labels["Type"].String() != expected.Labels["Type"].String() {
		t.Errorf("Expected filter labels to be %s, got %s", expected.Labels["Type"], filter.Labels["Type"])
	}
}

func config(contents string) (*Config, error) {

	cfg := &Config{}
	rootConfig := cfg.defaultConfiguration()

	// Fill configuration
	if _, err := toml.Decode(contents, &rootConfig); err != nil {
		return nil, err
	}

	if err := cfg.configureMain(&rootConfig); err != nil {
		return nil, err
	}

	return cfg, nil
}

const updateToml = `# Generic Updates
[[rule]]
description = "Non-Escalated"
when = "< 86400"
what = ": Update"
[rule.label]
Type = "PullRequest"
`
