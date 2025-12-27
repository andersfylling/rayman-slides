#!/usr/bin/env bash
#
# issue-bot.sh - Autonomous issue resolver using Claude Code
#
# Polls GitHub issues every minute and attempts to resolve them one at a time.
# Uses Claude Code in three phases:
#   1. Investigation - Analyze issue and post insights
#   2. Implementation - Fix the issue and commit
#   3. Closure - Close issue or mark as waiting-for-user
#
# Usage: ./scripts/issue-bot.sh
#
# Requirements:
#   - gh CLI (authenticated)
#   - claude CLI (Claude Code)
#
# Environment variables:
#   POLL_INTERVAL - Seconds between polls (default: 60)
#   DRY_RUN       - Set to 1 to print what would be done without running Claude
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
POLL_INTERVAL="${POLL_INTERVAL:-60}"
DRY_RUN="${DRY_RUN:-0}"

# Labels
LABEL_IN_PROGRESS="bot-in-progress"
LABEL_BOT_RESOLVED="bot-resolved"
LABEL_BOT_FAILED="bot-failed"
LABEL_WAITING_USER="waiting-for-user"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

error() {
    log "ERROR: $*" >&2
}

# Ensure we're in the project directory
cd "$PROJECT_DIR"

# Check dependencies
check_dependencies() {
    if ! command -v gh &> /dev/null; then
        error "gh CLI not found. Install from https://cli.github.com/"
        exit 1
    fi

    if ! command -v claude &> /dev/null; then
        error "claude CLI not found. Install Claude Code."
        exit 1
    fi

    # Check gh is authenticated
    if ! gh auth status &> /dev/null; then
        error "gh CLI not authenticated. Run 'gh auth login'"
        exit 1
    fi
}

# Get the next unprocessed issue
get_next_issue() {
    # Query open issues and filter for bug OR enhancement labels
    # Exclude issues already being processed or resolved by bot
    # Note: gh --label requires ALL labels, so we filter with jq instead
    gh issue list \
        --state open \
        --json number,title,body,labels \
        --jq ".[] | select(.labels | map(.name) | any(. == \"bug\" or . == \"enhancement\")) | select(.labels | map(.name) | index(\"$LABEL_IN_PROGRESS\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_RESOLVED\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_FAILED\") | not) | select(.labels | map(.name) | index(\"$LABEL_WAITING_USER\") | not)" \
        2>/dev/null | head -1
}

# Build prompt for Phase 1: Investigation
build_investigation_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_body="$3"

    cat <<EOF
You are investigating GitHub issue #${issue_number}: ${issue_title}

## Issue Description

${issue_body}

## Your Task: Investigation Only

Analyze this issue thoroughly. Do NOT make any code changes or commits yet.

1. Read the issue carefully and understand what is being reported
2. Explore the codebase to find relevant files and code sections
3. Identify the root cause or the components that need to be modified
4. Determine if this issue can be verified programmatically (tests, build, etc.) or requires manual user verification

## Output Format

When done investigating, output your findings in this exact format:

---INVESTIGATION_RESULT---
FILES: <comma-separated list of relevant file paths>
ROOT_CAUSE: <1-2 sentence description of the root cause or what needs to change>
APPROACH: <1-2 sentence description of how to fix it>
VERIFIABLE: <YES if can be tested programmatically, NO if requires manual user testing>
---END_INVESTIGATION---

Be concise but specific. This information will be posted as a comment on the issue.
EOF
}

# Fetch issue with all comments
fetch_issue_with_comments() {
    local issue_number="$1"

    # Get issue body
    local body=$(gh issue view "$issue_number" --json body --jq '.body // "No description provided."' 2>/dev/null)

    # Get all comments
    local comments=$(gh issue view "$issue_number" --json comments --jq '.comments[] | "**\(.author.login)** (\(.createdAt)):\n\(.body)\n"' 2>/dev/null)

    echo "## Issue Description"
    echo ""
    echo "$body"
    echo ""

    if [[ -n "$comments" ]]; then
        echo "## Comments"
        echo ""
        echo "$comments"
    fi
}

# Build prompt for Phase 2: Implementation
build_implementation_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_context="$3"
    local investigation="$4"

    cat <<EOF
You are implementing a fix for GitHub issue #${issue_number}: ${issue_title}

${issue_context}

## Investigation Summary

${investigation}

## Your Task: Implement the Fix

1. Make the necessary code changes based on the investigation
2. Test your changes (run \`make build\` at minimum, run tests if available)
3. Commit with a message that:
   - References the issue: "Fixes #${issue_number}" (for bugs) or "Closes #${issue_number}" (for features)
   - Explains WHY the change was needed, not what changed
   - Keep it concise (1-2 sentences)
4. Push the changes to main

## Commit Message Format

\`\`\`
<Short WHY explanation>

Fixes #${issue_number}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
\`\`\`

## Output Format

After implementing, output in this exact format:

---IMPLEMENTATION_RESULT---
STATUS: <SUCCESS or FAILED>
COMMIT: <commit SHA if successful, or N/A>
SUMMARY: <1-2 sentence summary of what was done>
ERROR: <error description if failed, or N/A>
---END_IMPLEMENTATION---

## Important

- Only commit if the fix is complete and tested
- If you cannot fix the issue, set STATUS to FAILED and explain why
- Do not make unrelated changes
- Keep changes minimal and focused
EOF
}

# Run Claude and capture output
run_claude() {
    local prompt="$1"
    local log_file="$2"

    local prompt_file=$(mktemp)
    echo "$prompt" > "$prompt_file"

    local output=""
    if output=$(claude -p "$(cat "$prompt_file")" --allowedTools "Bash,Read,Write,Edit,Glob,Grep" 2>&1 | tee "$log_file"); then
        rm -f "$prompt_file"
        echo "$output"
        return 0
    else
        local exit_code=$?
        rm -f "$prompt_file"
        echo "$output"
        return $exit_code
    fi
}

# Extract section from Claude output
extract_section() {
    local output="$1"
    local start_marker="$2"
    local end_marker="$3"

    # Use -- to prevent markers starting with --- from being interpreted as options
    echo "$output" | sed -n "/${start_marker}/,/${end_marker}/p" | grep -v -F -- "$start_marker" | grep -v -F -- "$end_marker"
}

# Extract field from section
extract_field() {
    local section="$1"
    local field="$2"

    # Use -- to prevent any pattern issues
    echo "$section" | grep -E -- "^${field}:" | sed "s/^${field}: *//"
}

# Process a single issue
process_issue() {
    local issue_json="$1"

    local issue_number=$(echo "$issue_json" | jq -r '.number')
    local issue_title=$(echo "$issue_json" | jq -r '.title')
    local issue_body=$(echo "$issue_json" | jq -r '.body // "No description provided."')

    log "Processing issue #${issue_number}: ${issue_title}"

    # Add in-progress label
    log "Adding '$LABEL_IN_PROGRESS' label..."
    if [[ "$DRY_RUN" != "1" ]]; then
        gh issue edit "$issue_number" --add-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
    fi

    if [[ "$DRY_RUN" == "1" ]]; then
        log "DRY RUN - Would process issue #${issue_number}"
        return 0
    fi

    # ============================================
    # PHASE 1: Investigation
    # ============================================
    log "Phase 1: Investigation..."

    local investigation_prompt=$(build_investigation_prompt "$issue_number" "$issue_title" "$issue_body")
    local investigation_output=$(run_claude "$investigation_prompt" "/tmp/claude-issue-${issue_number}-phase1.log")

    local investigation_section=$(extract_section "$investigation_output" "---INVESTIGATION_RESULT---" "---END_INVESTIGATION---")

    if [[ -z "$investigation_section" ]]; then
        log "Phase 1 failed: Could not extract investigation results"
        gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
        gh issue edit "$issue_number" --add-label "$LABEL_BOT_FAILED" 2>/dev/null || true
        gh issue comment "$issue_number" --body "🤖 **Bot Investigation Failed**

The issue bot could not complete the investigation phase. Manual intervention may be required.

See logs for details." 2>/dev/null || true
        return 1
    fi

    local files=$(extract_field "$investigation_section" "FILES")
    local root_cause=$(extract_field "$investigation_section" "ROOT_CAUSE")
    local approach=$(extract_field "$investigation_section" "APPROACH")
    local verifiable=$(extract_field "$investigation_section" "VERIFIABLE")

    # Post investigation comment
    log "Posting investigation findings..."
    gh issue comment "$issue_number" --body "🤖 **Bot Investigation Complete**

**Relevant Files:** \`${files}\`

**Root Cause:** ${root_cause}

**Approach:** ${approach}

**Programmatically Verifiable:** ${verifiable}

---
_Proceeding to implementation phase..._" 2>/dev/null || true

    # ============================================
    # PHASE 2: Implementation
    # ============================================
    log "Phase 2: Implementation..."

    # Re-fetch issue with all comments (including our investigation comment)
    log "Fetching fresh issue data with comments..."
    local issue_context=$(fetch_issue_with_comments "$issue_number")

    local implementation_prompt=$(build_implementation_prompt "$issue_number" "$issue_title" "$issue_context" "$investigation_section")
    local implementation_output=$(run_claude "$implementation_prompt" "/tmp/claude-issue-${issue_number}-phase2.log")

    local implementation_section=$(extract_section "$implementation_output" "---IMPLEMENTATION_RESULT---" "---END_IMPLEMENTATION---")

    if [[ -z "$implementation_section" ]]; then
        log "Phase 2 failed: Could not extract implementation results"
        gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
        gh issue edit "$issue_number" --add-label "$LABEL_BOT_FAILED" 2>/dev/null || true
        gh issue comment "$issue_number" --body "🤖 **Bot Implementation Failed**

The issue bot could not complete the implementation phase. Manual intervention may be required.

See logs for details." 2>/dev/null || true
        return 1
    fi

    local status=$(extract_field "$implementation_section" "STATUS")
    local commit_sha=$(extract_field "$implementation_section" "COMMIT")
    local summary=$(extract_field "$implementation_section" "SUMMARY")
    local impl_error=$(extract_field "$implementation_section" "ERROR")

    if [[ "$status" != "SUCCESS" ]]; then
        log "Phase 2 failed: Implementation unsuccessful"
        gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
        gh issue edit "$issue_number" --add-label "$LABEL_BOT_FAILED" 2>/dev/null || true
        gh issue comment "$issue_number" --body "🤖 **Bot Implementation Failed**

**Error:** ${impl_error}

Manual intervention is required." 2>/dev/null || true
        return 1
    fi

    # Post implementation comment
    log "Posting implementation results..."
    gh issue comment "$issue_number" --body "🤖 **Bot Implementation Complete**

**Commit:** ${commit_sha}

**Summary:** ${summary}

---
_Proceeding to closure phase..._" 2>/dev/null || true

    # ============================================
    # PHASE 3: Closure
    # ============================================
    log "Phase 3: Closure..."

    gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true

    if [[ "$verifiable" == "YES" ]]; then
        # Can be verified programmatically - close the issue
        log "Issue is programmatically verifiable - closing..."
        gh issue edit "$issue_number" --add-label "$LABEL_BOT_RESOLVED" 2>/dev/null || true
        gh issue close "$issue_number" --comment "🤖 **Issue Resolved**

This issue has been automatically fixed and verified. The fix has been pushed to main.

If you encounter any problems, please reopen this issue." 2>/dev/null || true
    else
        # Requires manual verification - add waiting-for-user label
        log "Issue requires manual verification - waiting for user..."
        gh issue edit "$issue_number" --add-label "$LABEL_WAITING_USER" 2>/dev/null || true
        gh issue comment "$issue_number" --body "🤖 **Awaiting User Verification**

The fix has been implemented and pushed to main, but this issue requires manual verification.

Please test the fix and:
- **Close this issue** if the fix works correctly
- **Reopen/comment** if there are still problems

The \`waiting-for-user\` label will remain until you verify." 2>/dev/null || true
    fi

    log "Issue #${issue_number} processing complete"
    return 0
}

# Main loop
main() {
    log "Issue Bot starting..."
    log "Project: $PROJECT_DIR"
    log "Poll interval: ${POLL_INTERVAL}s"
    log "Dry run: $DRY_RUN"

    check_dependencies

    # Ensure labels exist
    log "Ensuring bot labels exist..."
    if [[ "$DRY_RUN" != "1" ]]; then
        gh label create "$LABEL_IN_PROGRESS" --description "Bot is working on this issue" --color "FFA500" 2>/dev/null || true
        gh label create "$LABEL_BOT_RESOLVED" --description "Bot has addressed this issue" --color "00FF00" 2>/dev/null || true
        gh label create "$LABEL_BOT_FAILED" --description "Bot failed to resolve this issue" --color "FF0000" 2>/dev/null || true
        gh label create "$LABEL_WAITING_USER" --description "Bot fix applied, awaiting user verification" --color "0E8A16" 2>/dev/null || true
    fi

    log "Entering main loop (Ctrl+C to stop)..."

    while true; do
        # Pull latest changes first
        log "Pulling latest changes..."
        git pull --rebase origin main 2>/dev/null || true

        # Get next issue to process
        local issue_json=$(get_next_issue)

        if [[ -n "$issue_json" ]]; then
            process_issue "$issue_json" || true
        else
            log "No issues to process"
        fi

        log "Sleeping for ${POLL_INTERVAL}s..."
        sleep "$POLL_INTERVAL"
    done
}

# Handle Ctrl+C gracefully
trap 'log "Shutting down..."; exit 0' INT TERM

# Run main
main "$@"
