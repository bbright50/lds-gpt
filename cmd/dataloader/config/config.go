package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/spf13/viper"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type Config struct {
	Env Environment `mapstructure:"ENV"`

	// AWS specific configuration
	AWSRegion string `mapstructure:"AWS_REGION"`

	Port     string `mapstructure:"SERVER_PORT"`     // the port to bind the listening server to
	Hostname string `mapstructure:"SERVER_HOSTNAME"` // the hostname to bind the listening server to

	// FalkorDB connection (Redis URL) and target graph name.
	FalkorDBURL   string `mapstructure:"FALKORDB_URL"`
	FalkorDBGraph string `mapstructure:"FALKORDB_GRAPH"`

	// data directory containing scraped scripture JSON files
	DataDir string `mapstructure:"DATA_DIR"`
}

func bindEnvVars(cfg *Config) {
	t := reflect.TypeOf(*cfg)
	for i := range t.NumField() {
		field := t.Field(i)
		if tag := field.Tag.Get("mapstructure"); tag != "" {
			viper.BindEnv(tag)
		}
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// pull in environment variables
	bindEnvVars(cfg)

	// Look for .env file in current directory
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("config file not found, using environment variables...")
	}
	viper.AutomaticEnv()

	// build the config object
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// print configuration
	cfgJson, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	fmt.Println("config: ", string(cfgJson))

	return cfg, nil
}

func Validate(cfg *Config) error {
	if cfg.Env == "" {
		return fmt.Errorf("ENV is required")
	}

	envs := []Environment{EnvironmentDevelopment, EnvironmentProduction}
	if !slices.Contains(envs, cfg.Env) {
		return fmt.Errorf("ENV must be one of %v", envs)
	}

	if cfg.FalkorDBURL == "" {
		return fmt.Errorf("FALKORDB_URL is required")
	}

	if cfg.FalkorDBGraph == "" {
		return fmt.Errorf("FALKORDB_GRAPH is required")
	}

	if cfg.DataDir == "" {
		return fmt.Errorf("DATA_DIR is required")
	}

	return nil
}
