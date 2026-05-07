package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanyang0612/git-pr/internal/git"
)

const defaultTemplate = `## Why
{{why}}

## What
{{what}}

## Review deadline
_Please review by: _

## Others
{{others}}`

// LoadTemplate tries to load a custom PR template from the repo.
// Priority: .git-pr.md → .github/PULL_REQUEST_TEMPLATE.md → built-in default
func LoadTemplate() string {
	candidates := []string{
		".git-pr.md",
		".github/PULL_REQUEST_TEMPLATE.md",
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return defaultTemplate
}

// RepoRoot walks up from cwd to find the git root, so template lookup works
// regardless of which subdirectory the user is in.
func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Build constructs the full prompt sent to the AI.
func Build(info *git.Info, context string) string {
	// Change to repo root for template lookup
	root := findRepoRoot()
	if root != "" {
		prev, _ := os.Getwd()
		os.Chdir(root)
		defer os.Chdir(prev)
	}

	template := LoadTemplate()

	example := "- **`library/reservation_service.py`** — New `ReservationService` with:\n" +
		"  - `create(scheduled_at, payload)` — persists a reservation and registers an `at()` schedule in EventBridge targeting the Lambda handler\n" +
		"  - `get(reservation_id)` — retrieves by ID; raises `NotFoundException` if missing\n" +
		"  - `cancel(reservation_id)` — only allowed in `PENDING` status; deletes the schedule and transitions to `CANCELLED`"

	contextBlock := ""
	if strings.TrimSpace(context) != "" {
		contextBlock = fmt.Sprintf(`## Additional Context (spec / ticket / notes)
%s

Use this context to improve the accuracy of the Why section and overall description.`, context)
	}

	return fmt.Sprintf(`You are a senior software engineer writing a detailed, high-quality GitHub PR description. Analyze the code diff carefully and produce a description that reflects deep understanding of the changes — not just a surface-level summary.

## Branch Info
- Current branch: %s
- Base branch: %s

## Commit Messages
%s

## Changed Files
%s

## Code Diff (partial)
`+"```"+`
%s
`+"```"+`

%s

## Instructions

### Title
- Format: <type>(<scope>): <short description>
- Max 50 characters, in English
- Type: feat / fix / refactor / docs / chore / test / style / perf

### Description
Use the following PR template structure exactly. Replace each {{placeholder}} with real content.

%s

### Quality bar
- **Why**: Explain the motivation clearly — what was missing, broken, or needed. 2-4 sentences of real engineering rationale.
- **What**: Be specific. For each changed file, explain what was added/changed including method signatures, parameters, return values, status constraints, error handling, edge cases. Use nested bullets for sub-items. Use backticks for code identifiers and **bold** for file names.
- **Others**: Important implementation decisions, constraints, gotchas. Only write N/A if genuinely nothing to add.

### Example quality for "What":
%s

## Output Format
Respond ONLY with valid JSON. No extra text, no markdown fences:

{"title": "<type>(<scope>): <short description>", "description": "<full description with real newlines escaped as \\n>"}`,
		info.CurrentBranch,
		info.BaseBranch,
		info.Commits,
		info.DiffStat,
		info.FullDiff,
		contextBlock,
		template,
		example,
	)
}
