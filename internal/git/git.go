package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Info struct {
	CurrentBranch string
	BaseBranch    string
	Commits       string
	DiffStat      string
	FullDiff      string
}

func CurrentBranch() (string, error) {
	out, err := run("git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func IsInsideRepo() bool {
	_, err := run("git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func Collect(base string) (*Info, error) {
	current, err := CurrentBranch()
	if err != nil {
		return nil, err
	}

	commits, _ := run("git", "log",
		fmt.Sprintf("%s..%s", base, current),
		"--pretty=format:- %s (%an)")
	if commits == "" {
		commits = "(no commits found)"
	}

	diffStat, _ := runLimited("git", "diff",
		fmt.Sprintf("%s...%s", base, current),
		"--stat")
	if diffStat == "" {
		diffStat = "(no diff)"
	}

	// Only diff code files to keep prompt size reasonable
	codeExts := []string{
		"*.ts", "*.tsx", "*.js", "*.jsx",
		"*.py", "*.go", "*.rs", "*.java",
		"*.vue", "*.svelte", "*.rb", "*.php",
		"*.kt", "*.swift", "*.cs", "*.cpp",
	}
	args := []string{"diff", fmt.Sprintf("%s...%s", base, current), "--"}
	args = append(args, codeExts...)
	fullDiff, _ := runLimited(args[0], args[1:]...)

	return &Info{
		CurrentBranch: current,
		BaseBranch:    base,
		Commits:       commits,
		DiffStat:      diffStat,
		FullDiff:      fullDiff,
	}, nil
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

// runLimited runs a command and returns at most ~800 lines of output
func runLimited(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 800 {
		lines = lines[:800]
		lines = append(lines, "... (truncated)")
	}
	return strings.Join(lines, "\n"), nil
}
