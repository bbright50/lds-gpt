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

	// Chat/generation model on the same Ollama server (e.g. "gemma3:27b",
	// "llama3:70b"). Reuses OLLAMA_URL — the only reason this is a separate
	// knob is that embedding and generation want different model sizes and
	// tunings. Empty means "don't construct a chat client" (runtime app
	// paths that need one check this and error up-front).
	OllamaChatModel string `mapstructure:"OLLAMA_CHAT_MODEL"`

	// Number of chunks per /api/embed request in Phase 6. Larger = fewer HTTP
	// round-trips + better GPU batching on the server, but each request uses
	// more memory and takes longer (longer timeouts may be needed). Defaults
	// to 32 when unset — tune up for a GPU-backed Ollama, down for CPU.
	EmbedBatchSize int `mapstructure:"EMBED_BATCH_SIZE"`

	// Number of /api/embed requests in flight to Ollama at once. Default is
	// 1 (fully serial) because stock Ollama servers serialise requests per
	// model unless OLLAMA_NUM_PARALLEL is raised; firing more concurrency
	// than the server supports only queues up on the server side. Bump to
	// OLLAMA_NUM_PARALLEL's value on GPU-backed multi-slot setups.
	EmbedConcurrency int `mapstructure:"EMBED_CONCURRENCY"`

	// Maximum characters per chunk before truncation. Must fit inside the
	// embed model's context window. Rough rule of thumb: window_tokens × 3.5
	// — so for mxbai-embed-large (512 tokens) stay under ~1800; for a
	// 2048-token model use ~6000. Defaults to 2000 when unset. Exceeding
	// the model's window causes Ollama to reject the entire batch with 400,
	// and until the per-item fallback kicks in you lose every sibling chunk
	// in that batch.
	EmbedMaxTextLen int `mapstructure:"EMBED_MAX_TEXT_LEN"`

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
