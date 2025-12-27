#!/usr/bin/env bash
#
# issue-bot.sh - Autonomous issue resolver using Claude Code
#
# Polls GitHub issues every minute and attempts to resolve them one at a time.
# Uses Claude Code with -p flag to fix bugs and implement features.
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
STATE_FILE="$PROJECT_DIR/.issue-bot-state"
POLL_INTERVAL="${POLL_INTERVAL:-60}"
DRY_RUN="${DRY_RUN:-0}"

# Labels
LABEL_IN_PROGRESS="bot-in-progress"
LABEL_BOT_RESOLVED="bot-resolved"
LABEL_BOT_FAILED="bot-failed"

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
    # Query open issues with bug or enhancement labels
    # Exclude issues already being processed or resolved by bot
    gh issue list \
        --state open \
        --label "bug,enhancement" \
        --json number,title,body,labels \
        --jq ".[] | select(.labels | map(.name) | index(\"$LABEL_IN_PROGRESS\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_RESOLVED\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_FAILED\") | not)" \
        2>/dev/null | head -1
}

# Build the prompt for Claude
build_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_body="$3"

    cat <<EOF
You are fixing GitHub issue #${issue_number}: ${issue_title}

## Issue Description

${issue_body}

## Instructions

1. Analyze the issue and understand what needs to be fixed or implemented
2. Make the necessary code changes
3. Test your changes if possible (run \`make build\` at minimum)
4. Commit with a message that:
   - References the issue: "Fixes #${issue_number}" (for bugs) or "Closes #${issue_number}" (for features)
   - Explains WHY the change was needed, not what changed
   - Keep it concise (1-2 sentences)
   - Example: "Fixes #${issue_number} - Player velocity wasn't preserved during jump due to intent clearing"
5. Push the changes to main

## Commit Message Format

\`\`\`
<Short WHY explanation>

Fixes #${issue_number}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
\`\`\`

## Important

- Only commit if the fix is complete and tested
- If you cannot fix the issue, explain why and do NOT commit
- Do not make unrelated changes
- Keep changes minimal and focused

Begin by reading the issue carefully and exploring the relevant code.
EOF
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

    # Build prompt
    local prompt=$(build_prompt "$issue_number" "$issue_title" "$issue_body")

    if [[ "$DRY_RUN" == "1" ]]; then
        log "DRY RUN - Would run Claude with prompt:"
        echo "---"
        echo "$prompt"
        echo "---"
        return 0
    fi

    # Run Claude Code
    log "Starting Claude Code..."
    local claude_exit_code=0

    # Create a temp file for the prompt (handles special characters better)
    local prompt_file=$(mktemp)
    echo "$prompt" > "$prompt_file"

    # Run Claude with the prompt
    if claude -p "$(cat "$prompt_file")" --allowedTools "Bash,Read,Write,Edit,Glob,Grep" 2>&1 | tee /tmp/claude-issue-${issue_number}.log; then
        claude_exit_code=0
    else
        claude_exit_code=$?
    fi

    rm -f "$prompt_file"

    # Check if changes were pushed
    local current_sha=$(git rev-parse HEAD)
    local remote_sha=$(git rev-parse origin/main 2>/dev/null || echo "")

    if [[ "$current_sha" == "$remote_sha" ]] || git log --oneline -1 | grep -q "#${issue_number}"; then
        # Changes were committed and likely pushed
        log "Issue #${issue_number} appears to be resolved"
        gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
        gh issue edit "$issue_number" --add-label "$LABEL_BOT_RESOLVED" 2>/dev/null || true

        # Add a comment
        gh issue comment "$issue_number" --body "🤖 This issue was automatically addressed by the issue bot. Please verify the fix and close if resolved." 2>/dev/null || true
    else
        log "Issue #${issue_number} may not have been fully resolved"
        gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true

        if [[ $claude_exit_code -ne 0 ]]; then
            gh issue edit "$issue_number" --add-label "$LABEL_BOT_FAILED" 2>/dev/null || true
            gh issue comment "$issue_number" --body "🤖 The issue bot attempted to fix this but encountered an error. Manual intervention may be required." 2>/dev/null || true
        fi
    fi

    return $claude_exit_code
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
