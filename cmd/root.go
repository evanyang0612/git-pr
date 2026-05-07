package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/evanyang0612/git-pr/internal/ai"
	"github.com/evanyang0612/git-pr/internal/config"
	"github.com/evanyang0612/git-pr/internal/git"
	"github.com/evanyang0612/git-pr/internal/github"
	"github.com/evanyang0612/git-pr/internal/prompt"
)

// ANSI colors
const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[1;33m"
	blue   = "\033[0;34m"
	cyan   = "\033[0;36m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

func color(c, s string) string { return c + s + reset }
func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, color(red, "❌ "+fmt.Sprintf(format, a...))+"\n")
	os.Exit(1)
}

var rootCmd = &cobra.Command{
	Use:   "git-pr [base-branch]",
	Short: "AI-powered PR generator",
	Long:  "Generate PR title and description from git diff using AI, then push and create the PR.",
	Args:  cobra.MaximumNArgs(1),
	Run:   run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func run(cmd *cobra.Command, args []string) {
	// ── Load config ───────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fatalf("loading config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		fatalf("%v", err)
	}

	// ── Determine base branch ─────────────────────────────
	base := cfg.BaseBranch
	if len(args) > 0 {
		base = args[0]
	}

	// ── Checks ────────────────────────────────────────────
	fmt.Println(color(bold, "🔍 Checking environment..."))

	if !git.IsInsideRepo() {
		fatalf("not inside a git repository")
	}
	if !github.GHAvailable() {
		fatalf("GitHub CLI not found or not authenticated\n   Install: https://cli.github.com\n   Then run: gh auth login")
	}

	current, err := git.CurrentBranch()
	if err != nil {
		fatalf("getting current branch: %v", err)
	}
	if current == base {
		fatalf("you are on %s — switch to a feature branch first", base)
	}

	fmt.Println(color(green, "✅ Environment check passed"))
	fmt.Printf("\n%s %s%s%s → %s%s%s\n\n",
		color(cyan, "📌 Branch:"),
		bold, current, reset,
		bold, base, reset,
	)

	// ── Collect git info ──────────────────────────────────
	fmt.Println(color(blue, "📦 Collecting git info..."))
	info, err := git.Collect(base)
	if err != nil {
		fatalf("collecting git info: %v", err)
	}

	// ── Ask for context ───────────────────────────────────
	context := askContext()

	// ── Generate PR content ───────────────────────────────
	provider, err := ai.New(cfg)
	if err != nil {
		fatalf("initializing AI provider: %v", err)
	}

	prContent := generate(provider, info, context)

	// ── Interactive loop ──────────────────────────────────
	for {
		printResult(prContent)

		fmt.Println(color(yellow+bold, "What would you like to do?"))
		fmt.Println("  [1] 🚀 Push + create PR")
		fmt.Println("  [2] ✏️  Edit then create PR")
		fmt.Println("  [3] 📤 Push only")
		fmt.Println("  [4] 🔄 Add context and regenerate")
		fmt.Println("  [5] ❌ Cancel")
		fmt.Println()

		choice := readLine("Choose [1-5]: ")

		switch choice {
		case "1":
			pushAndCreate(base, current, prContent.Title, prContent.Description)
			return
		case "2":
			edited := editInEditor(prContent)
			pushAndCreate(base, current, edited.Title, edited.Description)
			return
		case "3":
			pushOnly(current, base)
			return
		case "4":
			fmt.Println()
			fmt.Println(color(yellow+bold, "📎 Paste your context below, then press Ctrl+D when done:"))
			fmt.Println()
			extra := readMultiline()
			if extra != "" {
				context = context + "\n" + extra
			}
			prContent = generate(provider, info, context)
		case "5":
			fmt.Println(color(yellow, "Cancelled"))
			return
		default:
			fmt.Println(color(red, "Invalid choice, please try again."))
		}
	}
}

// ── Helpers ───────────────────────────────────────────────

func askContext() string {
	fmt.Println()
	fmt.Println(color(yellow+bold, "📎 Do you have any additional context? (spec, ticket, notes)"))
	fmt.Println("   Press Enter to skip, or paste your content and press Ctrl+D when done:")
	fmt.Println()
	return readMultiline()
}

func generate(provider ai.Provider, info *git.Info, context string) *ai.PRContent {
	fmt.Printf("%s (via %s)\n", color(blue, "🤖 Analyzing with AI..."), provider.Name())
	p := prompt.Build(info, context)
	content, err := provider.Generate(p)
	if err != nil {
		fatalf("generating PR content: %v", err)
	}
	return content
}

func printResult(pr *ai.PRContent) {
	fmt.Println()
	fmt.Println(color(bold, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println(color(green+bold, "✅ PR content generated"))
	fmt.Println(color(bold, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
	fmt.Println(color(cyan+bold, "📌 Title:"))
	fmt.Printf("   %s\n", pr.Title)
	fmt.Println()
	fmt.Println(color(cyan+bold, "📝 Description:"))
	fmt.Println(pr.Description)
	fmt.Println()
	fmt.Println(color(bold, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Println()
}

func pushAndCreate(base, current, title, body string) {
	fmt.Println()
	fmt.Println(color(blue, "🚀 Pushing branch..."))
	if err := github.Push(current); err != nil {
		fatalf("pushing branch: %v", err)
	}

	fmt.Println(color(blue, "🔗 Creating PR..."))
	url, err := github.CreatePR(base, current, title, body)
	if err != nil {
		fatalf("creating PR: %v", err)
	}

	fmt.Println()
	fmt.Println(color(green+bold, "🎉 PR created successfully!"))
	fmt.Printf("   %s\n\n", color(cyan, url))
	github.OpenURL(url)
}

func pushOnly(current, base string) {
	fmt.Println()
	fmt.Println(color(blue, "🚀 Pushing branch..."))
	if err := github.Push(current); err != nil {
		fatalf("pushing branch: %v", err)
	}
	fmt.Println(color(green, "✅ Push complete"))
	fmt.Printf("\nTo create PR manually: %s\n", color(cyan, "gh pr create --base "+base))
}

func editInEditor(pr *ai.PRContent) *ai.PRContent {
	f, err := os.CreateTemp("", "git-pr-*.md")
	if err != nil {
		fatalf("creating temp file: %v", err)
	}
	defer os.Remove(f.Name())

	fmt.Fprintf(f, "%s\n---\n%s", pr.Title, pr.Description)
	f.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	c := exec.Command(editor, f.Name())
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	c.Run()

	data, _ := os.ReadFile(f.Name())
	parts := strings.SplitN(string(data), "\n---\n", 2)
	title := strings.TrimSpace(parts[0])
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}
	return &ai.PRContent{Title: title, Description: body}
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readMultiline() string {
	var sb strings.Builder
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		sb.WriteString(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}
	return strings.TrimSpace(sb.String())
}
