package config

import (
	"flag"
	"os"
)

var fVersion = flag.Bool("version", false, "Print version")
var fInstance = flag.String("instance", "0", "Running instance identifier")
var fMode = flag.String("mode", "dev", "Mode to use (dev, prod)")
var fEnvironment = flag.String("environment", "test", "(test, stage, prod)")
var fAddr = flag.String("addr", "127.0.0.1:8988", "Bind web application to port")

type Config struct {
	WebAddr      string
	Mode         string
	Environment  string
	JaegerUrl    string
	Instance     string
	PrintVersion bool
}

func NewConfiguration() *Config {

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

	return cfg
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
