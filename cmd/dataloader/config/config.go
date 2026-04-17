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

	Port     string `mapstructure:"SERVER_PORT"`     // the port to bind the listening server to
	Hostname string `mapstructure:"SERVER_HOSTNAME"` // the hostname to bind the listening server to

	// FalkorDB connection (Redis URL) and target graph name.
	FalkorDBURL   string `mapstructure:"FALKORDB_URL"`
	FalkorDBGraph string `mapstructure:"FALKORDB_GRAPH"`

	// Remote Ollama server used for Phase 6 embedding and query-time embedding.
	// The model's output dimensionality must match the 1024 hard-coded in
	// internal/falkor/schema.graphql (`@vector(dimensions: 1024)`) — e.g.
	// `mxbai-embed-large` or `snowflake-arctic-embed`. Picking a different
	// dimension requires a schema update + regen + full re-embed.
	OllamaURL   string `mapstructure:"OLLAMA_URL"`
	OllamaModel string `mapstructure:"OLLAMA_MODEL"`

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

	// .env loading is owned by dotenvx at the invocation boundary (see the
	// `dotenvx run --` wrapper in Taskfile.yml). Here we read exclusively
	// from the process environment so there is one source of truth; running
	// the binary directly without dotenvx still works as long as the vars
	// are exported.
	bindEnvVars(cfg)
	viper.AutomaticEnv()

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

	// OLLAMA_URL / OLLAMA_MODEL are checked at the embedding call sites
	// rather than here because `task load` (phases 1-5) does not need them.

	return nil
}
