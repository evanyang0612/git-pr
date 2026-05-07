package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider   string `yaml:"provider"`
	APIKey     string `yaml:"api_key"`
	Model      string `yaml:"model"`
	BaseBranch string `yaml:"base_branch"`
	OllamaURL  string `yaml:"ollama_url"`
}

var defaults = Config{
	Provider:   "anthropic",
	Model:      "",
	BaseBranch: "main",
	OllamaURL:  "http://localhost:11434",
}

// DefaultModels per provider
var DefaultModels = map[string]string{
	"anthropic": "claude-sonnet-4-20250514",
	"openai":    "gpt-4o",
	"gemini":    "gemini-2.0-flash",
	"ollama":    "llama3",
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "git-pr", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &defaults, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &defaults, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := defaults
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Env vars override config file
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" && cfg.Provider == "anthropic" {
		cfg.APIKey = key
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" && cfg.Provider == "openai" {
		cfg.APIKey = key
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" && cfg.Provider == "gemini" {
		cfg.APIKey = key
	}

	// Set default model if not specified
	if cfg.Model == "" {
		cfg.Model = DefaultModels[cfg.Provider]
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	switch c.Provider {
	case "anthropic", "openai", "gemini":
		if c.APIKey == "" {
			return fmt.Errorf("api_key is required for provider %q\nRun: git pr config --provider %s --api-key <your-key>", c.Provider, c.Provider)
		}
	case "ollama":
		// no key needed
	default:
		return fmt.Errorf("unknown provider %q (choose: anthropic, openai, gemini, ollama)", c.Provider)
	}
	return nil
}
