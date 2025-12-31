// Command issue-bot is an autonomous GitHub issue resolver using Claude Code.
//
// It implements a test-first workflow:
//  1. Monitor accepted issues → analyze and create failing tests
//  2. Create PR with test cases → link to issue
//  3. Monitor accepted PRs → implement fix until tests pass
//
// Usage:
//
//	issue-bot [flags]
//
// Flags:
//
//	-poll      Poll interval in seconds (default: 15)
//	-timeout   Claude timeout in seconds (default: 300)
//	-dry-run   Print actions without executing
//	-once      Run once then exit (don't loop)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Labels used by the bot
const (
	LabelAccepted    = "accepted"
	LabelInProgress  = "bot-in-progress"
	LabelWaitingUser = "waiting-for-user"
	LabelBotTestPR   = "bot-test-pr"
	LabelBotFailed   = "bot-failed"
)

// Config holds bot configuration
type Config struct {
	PollInterval  time.Duration
	ClaudeTimeout time.Duration
	DryRun        bool
	Once          bool
	OwnerUsername string
	ProjectDir    string
}

// Issue represents a GitHub issue
type Issue struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []Label  `json:"labels"`
	Author Author   `json:"author"`
	State  string   `json:"state"`
}

// PR represents a GitHub pull request
type PR struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Labels    []Label  `json:"labels"`
	State     string   `json:"state"`
	HeadRef   string   `json:"headRefName"`
	Mergeable string   `json:"mergeable"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// Author represents a GitHub user
type Author struct {
	Login string `json:"login"`
}

// Comment represents a GitHub comment
type Comment struct {
	Author    Author    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Bot is the main issue bot
type Bot struct {
	cfg    Config
	logger *log.Logger
}

func main() {
	pollInterval := flag.Int("poll", 15, "Poll interval in seconds")
	claudeTimeout := flag.Int("timeout", 300, "Claude timeout in seconds")
	dryRun := flag.Bool("dry-run", false, "Print actions without executing")
	once := flag.Bool("once", false, "Run once then exit")
	flag.Parse()

	// Find project directory (where .git is)
	projectDir, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Failed to find project root: %v", err)
	}

	// Get repo owner from git remote
	owner := getRepoOwner()

	cfg := Config{
		PollInterval:  time.Duration(*pollInterval) * time.Second,
		ClaudeTimeout: time.Duration(*claudeTimeout) * time.Second,
		DryRun:        *dryRun,
		Once:          *once,
		OwnerUsername: owner,
		ProjectDir:    projectDir,
	}

	bot := &Bot{
		cfg:    cfg,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	if err := bot.checkDependencies(); err != nil {
		log.Fatalf("Dependency check failed: %v", err)
	}

	bot.ensureLabels()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		bot.logger.Println("Shutting down...")
		os.Exit(0)
	}()

	bot.logger.Printf("Issue Bot starting (poll=%s, timeout=%s, dry-run=%v)",
		cfg.PollInterval, cfg.ClaudeTimeout, cfg.DryRun)

	bot.run()
}

func (b *Bot) run() {
	for {
		// Pull latest changes
		b.gitPull()

		// Auto-accept owner issues
		b.autoAcceptOwnerIssues()

		// Check waiting issues for new feedback
		b.checkWaitingIssuesForFeedback()

		// Process accepted issues (Phase 1: Test creation)
		if issue := b.getNextAcceptedIssue(); issue != nil {
			b.processIssue(issue)
		}

		// Process accepted PRs (Phase 2: Implementation)
		if pr := b.getNextAcceptedPR(); pr != nil {
			b.processPR(pr)
		}

		if b.cfg.Once {
			b.logger.Println("Single run complete, exiting")
			return
		}

		b.logger.Printf("Sleeping for %s...", b.cfg.PollInterval)
		time.Sleep(b.cfg.PollInterval)
	}
}

// processIssue handles an accepted issue - analyzes it and creates test cases
func (b *Bot) processIssue(issue *Issue) {
	b.logger.Printf("Processing issue #%d: %s", issue.Number, issue.Title)

	if b.cfg.DryRun {
		b.logger.Printf("[DRY RUN] Would process issue #%d", issue.Number)
		return
	}

	// Add in-progress label
	b.addLabel("issue", issue.Number, LabelInProgress)

	// Fetch full issue context with comments
	context := b.fetchIssueContext(issue.Number)

	// Check if issue type is bug or feature
	isBug := b.hasLabel(issue.Labels, "bug")
	isFeature := b.hasLabel(issue.Labels, "enhancement")

	if !isBug && !isFeature {
		b.logger.Printf("Issue #%d has neither 'bug' nor 'enhancement' label, skipping", issue.Number)
		b.removeLabel("issue", issue.Number, LabelInProgress)
		return
	}

	// Phase 1a: For features, check documentation alignment
	if isFeature {
		if conflicts := b.checkDocAlignment(issue, context); conflicts != "" {
			b.commentOnIssue(issue.Number, fmt.Sprintf(`🤖 **Documentation Alignment Check**

⚠️ **Potential conflicts detected:**

%s

Please clarify how this feature should align with the project direction, or update the documentation/ADRs first.`, conflicts))
			b.removeLabel("issue", issue.Number, LabelInProgress)
			b.addLabel("issue", issue.Number, LabelWaitingUser)
			return
		}
	}

	// Phase 1b: Check if we have enough info to reproduce
	analysis := b.analyzeIssue(issue, context, isBug)
	if analysis == nil {
		b.removeLabel("issue", issue.Number, LabelInProgress)
		b.addLabel("issue", issue.Number, LabelBotFailed)
		return
	}

	if analysis.NeedsMoreInfo {
		b.commentOnIssue(issue.Number, fmt.Sprintf(`🤖 **Clarification Needed**

%s

Please provide the requested information so I can create accurate test cases.`, analysis.Questions))
		b.removeLabel("issue", issue.Number, LabelInProgress)
		b.addLabel("issue", issue.Number, LabelWaitingUser)
		return
	}

	// Phase 1c: Create test cases
	testResult := b.createTestCases(issue, analysis)
	if testResult == nil {
		b.removeLabel("issue", issue.Number, LabelInProgress)
		b.addLabel("issue", issue.Number, LabelBotFailed)
		return
	}

	// Phase 1d: Create PR with tests
	prNumber := b.createTestPR(issue, testResult)
	if prNumber == 0 {
		b.removeLabel("issue", issue.Number, LabelInProgress)
		b.addLabel("issue", issue.Number, LabelBotFailed)
		return
	}

	// Link PR to issue and mark waiting
	b.commentOnIssue(issue.Number, fmt.Sprintf(`🤖 **Test Cases Created**

I've created PR #%d with test cases that reproduce this issue.

**What happens next:**
1. Review the test cases in the PR
2. Add the `+"`accepted`"+` label to the PR when ready
3. I'll then implement the fix to make the tests pass

The focus now moves to the PR. I'll wait for your approval there.`, prNumber))

	b.removeLabel("issue", issue.Number, LabelInProgress)
	b.addLabel("issue", issue.Number, LabelWaitingUser)

	b.logger.Printf("Issue #%d: Created test PR #%d", issue.Number, prNumber)
}

// processPR handles an accepted PR - implements the fix
func (b *Bot) processPR(pr *PR) {
	b.logger.Printf("Processing PR #%d: %s", pr.Number, pr.Title)

	if b.cfg.DryRun {
		b.logger.Printf("[DRY RUN] Would process PR #%d", pr.Number)
		return
	}

	// Add in-progress label
	b.addLabel("pr", pr.Number, LabelInProgress)

	// Check out the PR branch
	if err := b.checkoutPRBranch(pr); err != nil {
		b.logger.Printf("Failed to checkout PR branch: %v", err)
		b.removeLabel("pr", pr.Number, LabelInProgress)
		b.addLabel("pr", pr.Number, LabelBotFailed)
		return
	}

	// Implement the fix
	result := b.implementFix(pr)
	if result == nil || !result.Success {
		errMsg := "Unknown error"
		if result != nil {
			errMsg = result.Error
		}
		b.commentOnPR(pr.Number, fmt.Sprintf(`🤖 **Implementation Failed**

❌ %s

Manual intervention may be required.`, errMsg))
		b.removeLabel("pr", pr.Number, LabelInProgress)
		b.addLabel("pr", pr.Number, LabelBotFailed)
		b.checkoutMain()
		return
	}

	// Push the fix
	if err := b.pushChanges(pr.HeadRef); err != nil {
		b.logger.Printf("Failed to push changes: %v", err)
		b.removeLabel("pr", pr.Number, LabelInProgress)
		b.addLabel("pr", pr.Number, LabelBotFailed)
		b.checkoutMain()
		return
	}

	b.commentOnPR(pr.Number, fmt.Sprintf(`🤖 **Implementation Complete**

✅ %s

**Commit:** %s

All tests should now pass. Please review and merge when ready.`, result.Summary, result.CommitSHA))

	b.removeLabel("pr", pr.Number, LabelInProgress)
	b.checkoutMain()

	b.logger.Printf("PR #%d: Implementation complete", pr.Number)
}

// IssueAnalysis holds the result of analyzing an issue
type IssueAnalysis struct {
	NeedsMoreInfo   bool
	Questions       string
	RootCause       string
	RelevantFiles   []string
	TestStrategy    string
	ExpectedBehavior string
}

// TestResult holds the result of creating test cases
type TestResult struct {
	Branch    string
	TestFiles []string
	Summary   string
}

// ImplementResult holds the result of implementing a fix
type ImplementResult struct {
	Success   bool
	CommitSHA string
	Summary   string
	Error     string
}

// analyzeIssue uses Claude to analyze the issue and determine what's needed
func (b *Bot) analyzeIssue(issue *Issue, context string, isBug bool) *IssueAnalysis {
	issueType := "feature request"
	if isBug {
		issueType = "bug report"
	}

	prompt := fmt.Sprintf(`You are analyzing GitHub issue #%d: %s

This is a %s.

## Issue Context

%s

## Your Task

Analyze this issue to understand:
1. What is being reported/requested
2. What files are relevant (verify they exist!)
3. Whether you have enough information to write test cases
4. What the expected behavior should be

## CRITICAL: Verify Everything

- Use Glob/Read to confirm files exist before listing them
- Do NOT guess or assume - verify first
- If information is missing, ask for it

## Output Format

Output your analysis in this exact format:

---ANALYSIS_RESULT---
NEEDS_MORE_INFO: <YES if you need clarification, NO if you have enough>
QUESTIONS: <If NEEDS_MORE_INFO is YES: numbered list of specific questions. If NO: N/A>
ROOT_CAUSE: <1-2 sentence description of the issue/feature>
RELEVANT_FILES: <comma-separated list of VERIFIED file paths>
TEST_STRATEGY: <How to test this - what test file to create/modify, what to test>
EXPECTED_BEHAVIOR: <What should happen when the fix is complete>
---END_ANALYSIS---`, issue.Number, issue.Title, issueType, context)

	output, err := b.runClaude(prompt)
	if err != nil {
		b.logger.Printf("Claude analysis failed: %v", err)
		return nil
	}

	section := extractSection(output, "---ANALYSIS_RESULT---", "---END_ANALYSIS---")
	if section == "" {
		b.logger.Printf("Could not extract analysis from Claude output")
		b.commentOnIssue(issue.Number, "🤖 **Analysis Failed**\n\nCould not parse Claude's analysis output. Manual intervention required.")
		return nil
	}

	analysis := &IssueAnalysis{
		NeedsMoreInfo:    extractField(section, "NEEDS_MORE_INFO") == "YES",
		Questions:        extractField(section, "QUESTIONS"),
		RootCause:        extractField(section, "ROOT_CAUSE"),
		TestStrategy:     extractField(section, "TEST_STRATEGY"),
		ExpectedBehavior: extractField(section, "EXPECTED_BEHAVIOR"),
	}

	filesStr := extractField(section, "RELEVANT_FILES")
	if filesStr != "" && filesStr != "N/A" {
		for _, f := range strings.Split(filesStr, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				analysis.RelevantFiles = append(analysis.RelevantFiles, f)
			}
		}
	}

	return analysis
}

// checkDocAlignment checks if a feature aligns with project documentation
func (b *Bot) checkDocAlignment(issue *Issue, context string) string {
	prompt := fmt.Sprintf(`You are checking if GitHub issue #%d conflicts with project documentation.

## Issue

%s

## Your Task

Read the following files (if they exist):
- README.md
- AGENTS.md
- Any ADR files in adr/ directory

Check if implementing this feature would:
1. Contradict stated project goals
2. Conflict with architectural decisions
3. Go against documented guidelines

## Output Format

---ALIGNMENT_CHECK---
HAS_CONFLICTS: <YES or NO>
CONFLICTS: <If YES: describe conflicts. If NO: N/A>
---END_ALIGNMENT_CHECK---`, issue.Number, context)

	output, err := b.runClaude(prompt)
	if err != nil {
		return "" // Assume no conflicts on error
	}

	section := extractSection(output, "---ALIGNMENT_CHECK---", "---END_ALIGNMENT_CHECK---")
	if section == "" {
		return ""
	}

	if extractField(section, "HAS_CONFLICTS") == "YES" {
		return extractField(section, "CONFLICTS")
	}

	return ""
}

// createTestCases uses Claude to create test cases for the issue
func (b *Bot) createTestCases(issue *Issue, analysis *IssueAnalysis) *TestResult {
	prompt := fmt.Sprintf(`You are creating test cases for GitHub issue #%d: %s

## Analysis

Root Cause: %s
Relevant Files: %s
Test Strategy: %s
Expected Behavior: %s

## Your Task

1. Create a new git branch: issue-%d-tests
2. Write test cases that:
   - For bugs: FAIL with current code (reproduce the bug)
   - For features: Define expected behavior (will fail until implemented)
3. Commit the test files
4. Do NOT implement the fix - only write tests

## Test Guidelines

- Use existing test patterns in the codebase
- Keep tests focused and minimal
- Tests should clearly fail for the right reason

## Output Format

After creating the tests, output:

---TEST_RESULT---
BRANCH: <branch name>
TEST_FILES: <comma-separated list of test files created/modified>
SUMMARY: <1-2 sentence summary of what the tests cover>
---END_TEST_RESULT---`,
		issue.Number, issue.Title,
		analysis.RootCause,
		strings.Join(analysis.RelevantFiles, ", "),
		analysis.TestStrategy,
		analysis.ExpectedBehavior,
		issue.Number)

	output, err := b.runClaude(prompt)
	if err != nil {
		b.logger.Printf("Claude test creation failed: %v", err)
		b.commentOnIssue(issue.Number, "🤖 **Test Creation Failed**\n\nClaude encountered an error while creating tests.")
		return nil
	}

	section := extractSection(output, "---TEST_RESULT---", "---END_TEST_RESULT---")
	if section == "" {
		b.logger.Printf("Could not extract test result from Claude output")
		b.commentOnIssue(issue.Number, "🤖 **Test Creation Failed**\n\nCould not parse test creation output.")
		return nil
	}

	result := &TestResult{
		Branch:  extractField(section, "BRANCH"),
		Summary: extractField(section, "SUMMARY"),
	}

	filesStr := extractField(section, "TEST_FILES")
	if filesStr != "" {
		for _, f := range strings.Split(filesStr, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				result.TestFiles = append(result.TestFiles, f)
			}
		}
	}

	return result
}

// createTestPR creates a PR with the test cases
func (b *Bot) createTestPR(issue *Issue, testResult *TestResult) int {
	// Push the branch
	cmd := exec.Command("git", "push", "-u", "origin", testResult.Branch)
	cmd.Dir = b.cfg.ProjectDir
	if err := cmd.Run(); err != nil {
		b.logger.Printf("Failed to push test branch: %v", err)
		return 0
	}

	// Create PR
	title := fmt.Sprintf("Test cases for #%d: %s", issue.Number, issue.Title)
	body := fmt.Sprintf(`## Summary

Test cases for issue #%d.

%s

## Test Files

%s

## Next Steps

1. Review the test cases
2. Add the `+"`accepted`"+` label when ready for implementation
3. Bot will implement the fix to make tests pass

Refs #%d

🤖 Generated with [Claude Code](https://claude.com/claude-code)`,
		issue.Number,
		testResult.Summary,
		strings.Join(testResult.TestFiles, "\n- "),
		issue.Number)

	cmd = exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--head", testResult.Branch,
		"--label", LabelBotTestPR)
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		b.logger.Printf("Failed to create PR: %v", err)
		return 0
	}

	// Extract PR number from URL
	prURL := strings.TrimSpace(string(output))
	parts := strings.Split(prURL, "/")
	if len(parts) > 0 {
		var prNum int
		fmt.Sscanf(parts[len(parts)-1], "%d", &prNum)
		return prNum
	}

	return 0
}

// implementFix uses Claude to implement the fix
func (b *Bot) implementFix(pr *PR) *ImplementResult {
	// Extract issue number from PR body (Refs #N)
	issueNum := 0
	re := regexp.MustCompile(`Refs #(\d+)`)
	if matches := re.FindStringSubmatch(pr.Body); len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &issueNum)
	}

	prompt := fmt.Sprintf(`You are implementing a fix for PR #%d: %s

## PR Description

%s

## Your Task

1. Run the existing tests to see them fail
2. Implement the minimal fix to make all tests pass
3. Run tests again to verify
4. Commit with message referencing the issue

## Commit Message Format

<Short description of fix>

Fixes #%d

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>

## Output Format

---IMPLEMENTATION_RESULT---
SUCCESS: <YES or NO>
COMMIT_SHA: <commit SHA if successful, N/A if not>
SUMMARY: <1-2 sentence summary of the fix>
ERROR: <error description if failed, N/A if successful>
---END_IMPLEMENTATION---`, pr.Number, pr.Title, pr.Body, issueNum)

	output, err := b.runClaude(prompt)
	if err != nil {
		return &ImplementResult{Success: false, Error: err.Error()}
	}

	section := extractSection(output, "---IMPLEMENTATION_RESULT---", "---END_IMPLEMENTATION---")
	if section == "" {
		return &ImplementResult{Success: false, Error: "Could not parse implementation output"}
	}

	return &ImplementResult{
		Success:   extractField(section, "SUCCESS") == "YES",
		CommitSHA: extractField(section, "COMMIT_SHA"),
		Summary:   extractField(section, "SUMMARY"),
		Error:     extractField(section, "ERROR"),
	}
}

// GitHub API helpers

func (b *Bot) getNextAcceptedIssue() *Issue {
	cmd := exec.Command("gh", "issue", "list",
		"--state", "open",
		"--label", LabelAccepted,
		"--json", "number,title,body,labels,author,state",
		"--jq", fmt.Sprintf(`.[] | select(.labels | map(.name) | (index("%s") | not) and (index("%s") | not) and (index("%s") | not))`,
			LabelInProgress, LabelBotFailed, LabelWaitingUser))
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil
	}

	// Parse first issue
	var issues []Issue
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var issue Issue
		if err := decoder.Decode(&issue); err != nil {
			break
		}
		// Check it has bug or enhancement label
		if b.hasLabel(issue.Labels, "bug") || b.hasLabel(issue.Labels, "enhancement") {
			issues = append(issues, issue)
		}
	}

	if len(issues) > 0 {
		return &issues[0]
	}
	return nil
}

func (b *Bot) getNextAcceptedPR() *PR {
	cmd := exec.Command("gh", "pr", "list",
		"--state", "open",
		"--label", LabelAccepted,
		"--label", LabelBotTestPR,
		"--json", "number,title,body,labels,state,headRefName,mergeable",
		"--jq", fmt.Sprintf(`.[] | select(.labels | map(.name) | (index("%s") | not) and (index("%s") | not))`,
			LabelInProgress, LabelBotFailed))
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil
	}

	var prs []PR
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var pr PR
		if err := decoder.Decode(&pr); err != nil {
			break
		}
		prs = append(prs, pr)
	}

	if len(prs) > 0 {
		return &prs[0]
	}
	return nil
}

func (b *Bot) fetchIssueContext(number int) string {
	// Get issue body
	cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", number),
		"--json", "body,comments",
		"--jq", `.body`)
	cmd.Dir = b.cfg.ProjectDir
	bodyOutput, _ := cmd.Output()

	// Get comments
	cmd = exec.Command("gh", "issue", "view", fmt.Sprintf("%d", number),
		"--json", "comments",
		"--jq", `.comments[] | "**\(.author.login)**: \(.body)\n"`)
	cmd.Dir = b.cfg.ProjectDir
	commentsOutput, _ := cmd.Output()

	var sb strings.Builder
	sb.WriteString("## Issue Description\n\n")
	sb.WriteString(strings.TrimSpace(string(bodyOutput)))
	sb.WriteString("\n\n")

	if len(commentsOutput) > 0 {
		sb.WriteString("## Comments\n\n")
		sb.WriteString(string(commentsOutput))
	}

	return sb.String()
}

func (b *Bot) autoAcceptOwnerIssues() {
	if b.cfg.OwnerUsername == "" {
		return
	}

	cmd := exec.Command("gh", "issue", "list",
		"--state", "open",
		"--author", b.cfg.OwnerUsername,
		"--json", "number,labels",
		"--jq", fmt.Sprintf(`.[] | select(.labels | map(.name) | index("%s") | not) | .number`, LabelAccepted))
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		var num int
		fmt.Sscanf(line, "%d", &num)
		if num > 0 {
			b.logger.Printf("Auto-accepting owner issue #%d", num)
			b.addLabel("issue", num, LabelAccepted)
		}
	}
}

func (b *Bot) checkWaitingIssuesForFeedback() {
	cmd := exec.Command("gh", "issue", "list",
		"--state", "open",
		"--label", LabelWaitingUser,
		"--json", "number")
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		return
	}

	var issues []struct{ Number int }
	json.Unmarshal(output, &issues)

	for _, issue := range issues {
		// Check if last comment is from user (not bot)
		cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", issue.Number),
			"--json", "comments",
			"--jq", `.comments | last | .body | contains("🤖") | not`)
		cmd.Dir = b.cfg.ProjectDir

		result, _ := cmd.Output()
		if strings.TrimSpace(string(result)) == "true" {
			b.logger.Printf("Issue #%d: User feedback detected, removing waiting label", issue.Number)
			b.removeLabel("issue", issue.Number, LabelWaitingUser)
		}
	}
}

func (b *Bot) addLabel(itemType string, number int, label string) {
	if b.cfg.DryRun {
		b.logger.Printf("[DRY RUN] Would add label '%s' to %s #%d", label, itemType, number)
		return
	}

	var cmd *exec.Cmd
	if itemType == "pr" {
		cmd = exec.Command("gh", "pr", "edit", fmt.Sprintf("%d", number), "--add-label", label)
	} else {
		cmd = exec.Command("gh", "issue", "edit", fmt.Sprintf("%d", number), "--add-label", label)
	}
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

func (b *Bot) removeLabel(itemType string, number int, label string) {
	if b.cfg.DryRun {
		return
	}

	var cmd *exec.Cmd
	if itemType == "pr" {
		cmd = exec.Command("gh", "pr", "edit", fmt.Sprintf("%d", number), "--remove-label", label)
	} else {
		cmd = exec.Command("gh", "issue", "edit", fmt.Sprintf("%d", number), "--remove-label", label)
	}
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

func (b *Bot) hasLabel(labels []Label, name string) bool {
	for _, l := range labels {
		if l.Name == name {
			return true
		}
	}
	return false
}

func (b *Bot) commentOnIssue(number int, body string) {
	if b.cfg.DryRun {
		b.logger.Printf("[DRY RUN] Would comment on issue #%d", number)
		return
	}

	cmd := exec.Command("gh", "issue", "comment", fmt.Sprintf("%d", number), "--body", body)
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

func (b *Bot) commentOnPR(number int, body string) {
	if b.cfg.DryRun {
		b.logger.Printf("[DRY RUN] Would comment on PR #%d", number)
		return
	}

	cmd := exec.Command("gh", "pr", "comment", fmt.Sprintf("%d", number), "--body", body)
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

// Git helpers

func (b *Bot) gitPull() {
	cmd := exec.Command("git", "pull", "--rebase", "origin", "main")
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

func (b *Bot) checkoutPRBranch(pr *PR) error {
	cmd := exec.Command("git", "fetch", "origin", pr.HeadRef)
	cmd.Dir = b.cfg.ProjectDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "checkout", pr.HeadRef)
	cmd.Dir = b.cfg.ProjectDir
	return cmd.Run()
}

func (b *Bot) checkoutMain() {
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = b.cfg.ProjectDir
	cmd.Run()
}

func (b *Bot) pushChanges(branch string) error {
	cmd := exec.Command("git", "push", "origin", branch)
	cmd.Dir = b.cfg.ProjectDir
	return cmd.Run()
}

// Claude integration

func (b *Bot) runClaude(prompt string) (string, error) {
	ctx := fmt.Sprintf("timeout %ds", int(b.cfg.ClaudeTimeout.Seconds()))

	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s claude -p %q --allowedTools 'Bash,Read,Write,Edit,Glob,Grep'",
		ctx, prompt))
	cmd.Dir = b.cfg.ProjectDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude failed: %w", err)
	}

	return string(output), nil
}

// Utility functions

func (b *Bot) checkDependencies() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found")
	}
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found")
	}

	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh not authenticated")
	}

	return nil
}

func (b *Bot) ensureLabels() {
	if b.cfg.DryRun {
		return
	}

	labels := map[string]string{
		LabelAccepted:    "0052CC",
		LabelInProgress:  "FFA500",
		LabelWaitingUser: "0E8A16",
		LabelBotTestPR:   "6F42C1",
		LabelBotFailed:   "FF0000",
	}

	for name, color := range labels {
		cmd := exec.Command("gh", "label", "create", name, "--color", color, "--force")
		cmd.Dir = b.cfg.ProjectDir
		cmd.Run()
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

func getRepoOwner() string {
	cmd := exec.Command("gh", "repo", "view", "--json", "owner", "--jq", ".owner.login")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func extractSection(output, startMarker, endMarker string) string {
	start := strings.Index(output, startMarker)
	if start == -1 {
		return ""
	}
	start += len(startMarker)

	end := strings.Index(output[start:], endMarker)
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(output[start : start+end])
}

func extractField(section, field string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*(.*)$`, regexp.QuoteMeta(field)))
	matches := re.FindStringSubmatch(section)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
