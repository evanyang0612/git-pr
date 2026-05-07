package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Push pushes the current branch to origin.
func Push(branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CreatePR uses the gh CLI to open a PR and returns the PR URL.
func CreatePR(base, head, title, body string) (string, error) {
	out, err := exec.Command("gh", "pr", "create",
		"--base", base,
		"--head", head,
		"--title", title,
		"--body", body,
	).Output()
	if err != nil {
		return "", fmt.Errorf("gh pr create failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GHAvailable returns true if gh CLI is installed and authenticated.
func GHAvailable() bool {
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}
	return exec.Command("gh", "auth", "status").Run() == nil
}

// OpenURL opens a URL in the default browser.
func OpenURL(url string) {
	for _, cmd := range []string{"open", "xdg-open", "start"} {
		if _, err := exec.LookPath(cmd); err == nil {
			exec.Command(cmd, url).Start()
			return
		}
	}
}
