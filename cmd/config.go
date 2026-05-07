package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/evanyang0612/git-pr/internal/config"
)

var (
	cfgProvider   string
	cfgAPIKey     string
	cfgModel      string
	cfgBaseBranch string
	cfgOllamaURL  string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set configuration (provider, API key, model, base branch)",
	Example: `  git pr config --provider anthropic --api-key sk-ant-...
  git pr config --provider openai --api-key sk-...
  git pr config --provider gemini --api-key AIza...
  git pr config --provider ollama --model llama3
  git pr config --base develop`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fatalf("loading config: %v", err)
		}

		changed := false

		if cfgProvider != "" {
			cfg.Provider = cfgProvider
			changed = true
		}
		if cfgAPIKey != "" {
			cfg.APIKey = cfgAPIKey
			changed = true
		}
		if cfgModel != "" {
			cfg.Model = cfgModel
			changed = true
		}
		if cfgBaseBranch != "" {
			cfg.BaseBranch = cfgBaseBranch
			changed = true
		}
		if cfgOllamaURL != "" {
			cfg.OllamaURL = cfgOllamaURL
			changed = true
		}

		if !changed {
			// Print current config
			fmt.Printf("provider:    %s\n", cfg.Provider)
			fmt.Printf("model:       %s\n", cfg.Model)
			fmt.Printf("base_branch: %s\n", cfg.BaseBranch)
			if cfg.APIKey != "" {
				masked := cfg.APIKey
				if len(masked) > 8 {
					masked = masked[:8] + "..."
				}
				fmt.Printf("api_key:     %s\n", masked)
			}
			if cfg.OllamaURL != "" {
				fmt.Printf("ollama_url:  %s\n", cfg.OllamaURL)
			}
			return
		}

		if err := config.Save(cfg); err != nil {
			fatalf("saving config: %v", err)
		}
		fmt.Println(color(green, "✅ Config saved (~/.config/git-pr/config.yaml)"))
	},
}

func init() {
	configCmd.Flags().StringVar(&cfgProvider, "provider", "", "AI provider (anthropic, openai, gemini, ollama)")
	configCmd.Flags().StringVar(&cfgAPIKey, "api-key", "", "API key for the provider")
	configCmd.Flags().StringVar(&cfgModel, "model", "", "Model to use (e.g. gpt-4o, claude-sonnet-4-20250514)")
	configCmd.Flags().StringVar(&cfgBaseBranch, "base", "", "Default base branch (e.g. main, develop)")
	configCmd.Flags().StringVar(&cfgOllamaURL, "ollama-url", "", "Ollama base URL (default: http://localhost:11434)")
}
