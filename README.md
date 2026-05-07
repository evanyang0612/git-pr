# git-pr

AI-powered PR title and description generator. Analyzes your `git diff` and commit messages to generate detailed, high-quality PR content — then pushes and creates the PR for you.

```
git pr          # base branch: main (default)
git pr develop  # specify base branch
```

## Install

```bash
brew install evanyang0612/tap/git-pr
```

## Setup

```bash
# Anthropic (Claude)
git pr config --provider anthropic --api-key sk-ant-...

# OpenAI (GPT-4o)
git pr config --provider openai --api-key sk-...

# Google (Gemini)
git pr config --provider gemini --api-key AIza...

# Ollama (local, no key needed)
git pr config --provider ollama --model llama3

# Set default base branch
git pr config --base develop
```

API keys can also be set via environment variables:
```bash
export ANTHROPIC_API_KEY=sk-ant-...
export OPENAI_API_KEY=sk-...
export GEMINI_API_KEY=AIza...
```

## Usage

```bash
# On your feature branch:
git pr            # generates PR, asks for optional context, shows interactive menu
git pr develop    # use develop as base instead of main
```

**Interactive menu after generation:**
```
[1] 🚀 Push + create PR
[2] ✏️  Edit then create PR
[3] 📤 Push only
[4] 🔄 Add context and regenerate
[5] ❌ Cancel
```

## Custom PR template

Create a `.git-pr.md` in your repo root to define your team's PR format:

```markdown
## Why
{{why}}

## What
{{what}}

## Review deadline
_Please review by: _

## Others
{{others}}
```

Falls back to `.github/PULL_REQUEST_TEMPLATE.md` if `.git-pr.md` doesn't exist.

## Config

Stored at `~/.config/git-pr/config.yaml`:

```yaml
provider: anthropic
api_key: sk-ant-...
model: claude-sonnet-4-20250514
base_branch: main
```

View current config:
```bash
git pr config
```

## Requirements

- [`gh`](https://cli.github.com) — GitHub CLI (for creating PRs)
- An API key from Anthropic, OpenAI, or Google — or [Ollama](https://ollama.com) running locally

## License

MIT
