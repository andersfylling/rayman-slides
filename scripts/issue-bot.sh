#!/usr/bin/env bash
#
# issue-bot.sh - Autonomous issue resolver using Claude Code
#
# Polls GitHub issues every 15 seconds and attempts to resolve them one at a time.
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
#   POLL_INTERVAL  - Seconds between polls (default: 15)
#   CLAUDE_TIMEOUT - Max seconds for Claude to run per phase (default: 300)
#   DRY_RUN        - Set to 1 to print what would be done without running Claude
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
POLL_INTERVAL="${POLL_INTERVAL:-15}"
DRY_RUN="${DRY_RUN:-0}"

# Labels
LABEL_ACCEPTED="accepted"
LABEL_IN_PROGRESS="bot-in-progress"
LABEL_BOT_RESOLVED="bot-resolved"
LABEL_BOT_FAILED="bot-failed"
LABEL_WAITING_USER="waiting-for-user"

# Auto-accept issues from the project owner
OWNER_USERNAME="andersfylling"

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

# Auto-accept issues from the project owner
auto_accept_owner_issues() {
    log "Checking for owner issues to auto-accept..."

    # Get open issues from owner without accepted label
    local owner_issues=$(gh issue list \
        --state open \
        --author "$OWNER_USERNAME" \
        --json number,labels \
        --jq ".[] | select(.labels | map(.name) | index(\"$LABEL_ACCEPTED\") | not) | .number" 2>/dev/null)

    for issue_number in $owner_issues; do
        log "Issue #${issue_number}: Auto-accepting (created by $OWNER_USERNAME)"
        gh issue edit "$issue_number" --add-label "$LABEL_ACCEPTED" 2>/dev/null || true
    done
}

# Check waiting-for-user issues for new user feedback and remove label if found
check_waiting_issues_for_feedback() {
    log "Checking waiting-for-user issues for new feedback..."

    # Get all open issues with waiting-for-user label
    local waiting_issues=$(gh issue list \
        --state open \
        --label "$LABEL_WAITING_USER" \
        --json number \
        --jq '.[].number' 2>/dev/null)

    for issue_number in $waiting_issues; do
        # Check if the last comment is NOT from the bot (doesn't contain 🤖)
        local last_comment_is_bot=$(gh issue view "$issue_number" --json comments \
            --jq '.comments | last | .body | contains("🤖")' 2>/dev/null)

        if [[ "$last_comment_is_bot" == "false" ]]; then
            log "Issue #${issue_number}: User feedback detected, removing waiting-for-user label"
            gh issue edit "$issue_number" --remove-label "$LABEL_WAITING_USER" 2>/dev/null || true
        fi
    done
}

# Get the next unprocessed issue (must have 'accepted' label)
get_next_issue() {
    # Query open issues with accepted label and bug/enhancement labels
    # Exclude issues already being processed or resolved by bot
    gh issue list \
        --state open \
        --label "$LABEL_ACCEPTED" \
        --json number,title,body,labels \
        --jq ".[] | select(.labels | map(.name) | any(. == \"bug\" or . == \"enhancement\")) | select(.labels | map(.name) | index(\"$LABEL_IN_PROGRESS\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_RESOLVED\") | not) | select(.labels | map(.name) | index(\"$LABEL_BOT_FAILED\") | not) | select(.labels | map(.name) | index(\"$LABEL_WAITING_USER\") | not)" \
        2>/dev/null | head -1
}

# Build prompt for Phase 0: Conflict Check
build_conflict_check_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_context="$3"

    cat <<EOF
You are checking if GitHub issue #${issue_number} conflicts with the project's architectural decisions.

${issue_context}

## Your Task: Conflict Check Only

Check if implementing this issue would conflict with or contradict:
1. The project's ADRs (Architecture Decision Records) in the adr/ directory
2. The AGENTS.md file (project guidelines for AI agents)
3. The README.md file (project overview and goals)

Read these files and determine if the requested feature/fix:
- Contradicts any architectural decisions
- Goes against the project's stated goals or principles
- Would require changing fundamental project decisions

## Output Format

Output your findings in this exact format:

---CONFLICT_CHECK_RESULT---
HAS_CONFLICTS: <YES if there are conflicts or concerns, NO if the issue aligns with project direction>
CONFLICTS: <If HAS_CONFLICTS is YES: describe each conflict/concern, one per line. If NO: write "None">
RECOMMENDATION: <If HAS_CONFLICTS is YES: what should be clarified or decided. If NO: "Proceed with implementation">
---END_CONFLICT_CHECK---

Be thorough but concise. Only flag genuine conflicts, not minor implementation details.
EOF
}

# Check if conflict check was already done
check_conflict_check_complete() {
    local issue_number="$1"

    gh issue view "$issue_number" --json comments \
        --jq '.comments[] | select(.body | contains("Conflict Check Complete")) | .body' 2>/dev/null | head -1 | grep -q "Conflict Check Complete"
}

# Build prompt for Phase 1: Investigation
build_investigation_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_context="$3"

    cat <<EOF
You are investigating GitHub issue #${issue_number}: ${issue_title}

${issue_context}

## Your Task: Investigation Only

Analyze this issue thoroughly. Do NOT make any code changes or commits yet.

1. Read the issue carefully and understand what is being reported
2. Explore the codebase to find relevant files and code sections
3. Identify the root cause or the components that need to be modified
4. Determine if there are multiple valid approaches to fix this
5. Determine if this issue can be verified programmatically (tests, build, etc.) or requires manual user verification

## Output Format

When done investigating, output your findings in this exact format:

---INVESTIGATION_RESULT---
FILES: <comma-separated list of relevant file paths>
ROOT_CAUSE: <1-2 sentence description of the root cause or what needs to change>
NEEDS_DECISION: <YES if there are multiple valid approaches and user should choose, NO if there's one clear approach>
APPROACH: <If NEEDS_DECISION is NO: 1-2 sentence description of how to fix it>
APPROACHES: <If NEEDS_DECISION is YES: numbered list of approaches, one per line, format "1. Description", "2. Description", etc.>
VERIFIABLE: <YES if can be tested programmatically, NO if requires manual user testing>
---END_INVESTIGATION---

IMPORTANT:
- If the user has already indicated a preference in the comments, set NEEDS_DECISION to NO and use their preferred approach.
- Only set NEEDS_DECISION to YES if there are genuinely different approaches with meaningful trade-offs AND the user hasn't expressed a preference.

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

# Check if investigation phase was already completed
get_existing_investigation() {
    local issue_number="$1"

    # Look for our investigation comment
    local investigation_comment=$(gh issue view "$issue_number" --json comments \
        --jq '.comments[] | select(.body | contains("Bot Investigation Complete")) | .body' 2>/dev/null | head -1)

    if [[ -z "$investigation_comment" ]]; then
        return 1
    fi

    # Extract the fields from the comment
    local files=$(echo "$investigation_comment" | grep -oP '(?<=\*\*Relevant Files:\*\* `)[^`]+' || echo "")
    local root_cause=$(echo "$investigation_comment" | grep -oP '(?<=\*\*Root Cause:\*\* ).*' || echo "")
    local approach=$(echo "$investigation_comment" | grep -oP '(?<=\*\*Approach:\*\* ).*' || echo "")
    local verifiable=$(echo "$investigation_comment" | grep -oP '(?<=\*\*Programmatically Verifiable:\*\* )(YES|NO)' || echo "NO")

    # Return as a structured format
    echo "FILES: ${files}"
    echo "ROOT_CAUSE: ${root_cause}"
    echo "APPROACH: ${approach}"
    echo "VERIFIABLE: ${verifiable}"
}

# Check if implementation phase was already completed
check_implementation_complete() {
    local issue_number="$1"

    gh issue view "$issue_number" --json comments \
        --jq '.comments[] | select(.body | contains("Bot Implementation Complete")) | .body' 2>/dev/null | head -1 | grep -q "Bot Implementation Complete"
}

# Check if awaiting user verification comment was already posted
check_awaiting_verification() {
    local issue_number="$1"

    gh issue view "$issue_number" --json comments \
        --jq '.comments[] | select(.body | contains("Awaiting User Verification")) | .body' 2>/dev/null | head -1 | grep -q "Awaiting User Verification"
}

# Check if there's user feedback after the last bot implementation/verification comment
# Returns 0 (true) if there's user feedback that needs addressing
check_user_feedback_pending() {
    local issue_number="$1"

    # Get all comments with timestamps and authors
    local comments_json=$(gh issue view "$issue_number" --json comments 2>/dev/null)

    # Find the timestamp of the last bot comment (Implementation Complete or Awaiting Verification)
    local last_bot_time=$(echo "$comments_json" | jq -r '
        .comments
        | map(select(.body | contains("Bot Implementation Complete") or contains("Awaiting User Verification")))
        | last
        | .createdAt // empty
    ')

    if [[ -z "$last_bot_time" ]]; then
        return 1  # No bot comment found, no feedback to check
    fi

    # Check if there are any non-bot comments after the last bot comment
    local user_feedback=$(echo "$comments_json" | jq -r --arg bot_time "$last_bot_time" '
        .comments
        | map(select(.createdAt > $bot_time and (.body | contains("🤖") | not)))
        | length
    ')

    if [[ "$user_feedback" -gt 0 ]]; then
        return 0  # User feedback exists
    fi
    return 1  # No user feedback
}

# Build prompt for Phase 2: Implementation
build_implementation_prompt() {
    local issue_number="$1"
    local issue_title="$2"
    local issue_context="$3"
    local investigation="$4"
    local verifiable="$5"

    # Determine issue reference keyword based on verifiability
    # "Fixes/Closes" auto-closes on GitHub, "Refs" does not
    local issue_keyword="Refs"
    local keyword_note="Use 'Refs' (NOT 'Fixes' or 'Closes') because this issue requires manual user verification."
    if [[ "$verifiable" == "YES" ]]; then
        issue_keyword="Fixes"
        keyword_note="Use 'Fixes' because this issue can be verified programmatically."
    fi

    cat <<EOF
You are implementing a fix for GitHub issue #${issue_number}: ${issue_title}

${issue_context}

## Investigation Summary

${investigation}

## Your Task: Implement the Fix

1. Make the necessary code changes based on the investigation
2. Test your changes (run \`make build\` at minimum, run tests if available)
3. Commit with a message that:
   - References the issue with "${issue_keyword} #${issue_number}"
   - ${keyword_note}
   - IMPORTANT: Do NOT use "Fixes", "Closes", or "Resolves" unless VERIFIABLE was YES - these keywords auto-close the issue on GitHub!
   - Explains WHY the change was needed, not what changed
   - Keep it concise (1-2 sentences)
4. Push the changes to main

## Commit Message Format

\`\`\`
<Short WHY explanation>

${issue_keyword} #${issue_number}

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
- CRITICAL: Only use "Fixes/Closes/Resolves" if VERIFIABLE is YES, otherwise use "Refs"
EOF
}

# Run Claude and capture output
run_claude() {
    local prompt="$1"
    local log_file="$2"
    local timeout_seconds="${CLAUDE_TIMEOUT:-300}"  # Default 5 minutes

    local prompt_file=$(mktemp)
    echo "$prompt" > "$prompt_file"

    local output=""
    if output=$(timeout "${timeout_seconds}s" claude -p "$(cat "$prompt_file")" --allowedTools "Bash,Read,Write,Edit,Glob,Grep" 2>&1 | tee "$log_file"); then
        rm -f "$prompt_file"
        echo "$output"
        return 0
    else
        local exit_code=$?
        rm -f "$prompt_file"
        if [[ $exit_code -eq 124 ]]; then
            log "Claude timed out after ${timeout_seconds}s"
        fi
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
    # PHASE DETECTION: Check what's already done
    # ============================================
    local investigation_section=""
    local verifiable=""
    local skip_to_phase=0  # Start from Phase 0 (Conflict Check)

    # Check for pending user feedback first - if exists, start fresh from Phase 0
    if check_user_feedback_pending "$issue_number"; then
        log "Phase detection: User feedback pending - restarting from Phase 0"
        skip_to_phase=0
    # Check if implementation is already complete
    elif check_implementation_complete "$issue_number"; then
        investigation_section=$(get_existing_investigation "$issue_number" || echo "")
        verifiable=$(extract_field "$investigation_section" "VERIFIABLE")
        [[ -z "$verifiable" ]] && verifiable="NO"
        log "Phase detection: Implementation already complete, skipping to Phase 3"
        skip_to_phase=3
    # Check if investigation is already complete
    elif investigation_section=$(get_existing_investigation "$issue_number"); then
        log "Phase detection: Investigation already complete, skipping to Phase 2"
        skip_to_phase=2
        verifiable=$(extract_field "$investigation_section" "VERIFIABLE")
        [[ -z "$verifiable" ]] && verifiable="NO"
    # Check if conflict check is already complete
    elif check_conflict_check_complete "$issue_number"; then
        log "Phase detection: Conflict check already complete, skipping to Phase 1"
        skip_to_phase=1
    fi

    # ============================================
    # PHASE 0: Conflict Check
    # ============================================
    if [[ $skip_to_phase -le 0 ]]; then
        log "Phase 0: Conflict Check..."

        # Fetch issue with all comments
        log "Fetching issue data with comments..."
        local issue_context=$(fetch_issue_with_comments "$issue_number")

        local conflict_prompt=$(build_conflict_check_prompt "$issue_number" "$issue_title" "$issue_context")
        local conflict_output=$(run_claude "$conflict_prompt" "/tmp/claude-issue-${issue_number}-phase0.log")

        local conflict_section=$(extract_section "$conflict_output" "---CONFLICT_CHECK_RESULT---" "---END_CONFLICT_CHECK---")

        if [[ -z "$conflict_section" ]]; then
            log "Phase 0 failed: Could not extract conflict check results"
            gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
            gh issue edit "$issue_number" --add-label "$LABEL_BOT_FAILED" 2>/dev/null || true
            gh issue comment "$issue_number" --body "🤖 **Bot Conflict Check Failed**

The issue bot could not complete the conflict check phase. Manual intervention may be required.

See logs for details." 2>/dev/null || true
            return 1
        fi

        local has_conflicts=$(extract_field "$conflict_section" "HAS_CONFLICTS")
        local conflicts=$(extract_field "$conflict_section" "CONFLICTS")
        local recommendation=$(extract_field "$conflict_section" "RECOMMENDATION")

        if [[ "$has_conflicts" == "YES" ]]; then
            log "Conflicts found - waiting for owner decision"
            gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
            gh issue edit "$issue_number" --add-label "$LABEL_WAITING_USER" 2>/dev/null || true
            gh issue comment "$issue_number" --body "🤖 **Conflict Check Complete - Decision Needed**

⚠️ **Potential conflicts with project architecture detected:**

${conflicts}

**Recommendation:** ${recommendation}

---
@${OWNER_USERNAME} Please review and either:
- Approve proceeding (reply \"approved\" or \"proceed\")
- Reject the issue
- Provide guidance on how to resolve the conflicts

The bot will continue once approved." 2>/dev/null || true
            return 0
        fi

        # No conflicts - post result and continue
        log "No conflicts found - proceeding..."
        gh issue comment "$issue_number" --body "🤖 **Conflict Check Complete**

✅ No conflicts with project ADRs, AGENTS.md, or README.md detected.

---
_Proceeding to investigation phase..._" 2>/dev/null || true
    else
        log "Phase 0: Skipped (already complete)"
    fi

    # ============================================
    # PHASE 1: Investigation
    # ============================================
    if [[ $skip_to_phase -le 1 ]]; then
        log "Phase 1: Investigation..."

        # Fetch issue with all comments (important for re-investigations after user feedback)
        log "Fetching issue data with comments..."
        local issue_context=$(fetch_issue_with_comments "$issue_number")

        local investigation_prompt=$(build_investigation_prompt "$issue_number" "$issue_title" "$issue_context")
        local investigation_output=$(run_claude "$investigation_prompt" "/tmp/claude-issue-${issue_number}-phase1.log")

        investigation_section=$(extract_section "$investigation_output" "---INVESTIGATION_RESULT---" "---END_INVESTIGATION---")

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
        local needs_decision=$(extract_field "$investigation_section" "NEEDS_DECISION")
        local approach=$(extract_field "$investigation_section" "APPROACH")
        local approaches=$(extract_field "$investigation_section" "APPROACHES")
        verifiable=$(extract_field "$investigation_section" "VERIFIABLE")

        # Check if user needs to make a decision
        if [[ "$needs_decision" == "YES" ]]; then
            log "Multiple approaches found - waiting for user decision"
            gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
            gh issue edit "$issue_number" --add-label "$LABEL_WAITING_USER" 2>/dev/null || true
            gh issue comment "$issue_number" --body "🤖 **Bot Investigation Complete - Decision Needed**

**Relevant Files:** \`${files}\`

**Root Cause:** ${root_cause}

**Possible Approaches:**
${approaches}

**Programmatically Verifiable:** ${verifiable}

---
⚠️ **Please reply with your preferred approach number (e.g., \"Use approach 1\" or \"Go with option 2\").**

The bot will continue implementation once you've decided." 2>/dev/null || true
            return 0
        fi

        # Post investigation comment and continue to implementation
        log "Posting investigation findings..."
        gh issue comment "$issue_number" --body "🤖 **Bot Investigation Complete**

**Relevant Files:** \`${files}\`

**Root Cause:** ${root_cause}

**Approach:** ${approach}

**Programmatically Verifiable:** ${verifiable}

---
_Proceeding to implementation phase..._" 2>/dev/null || true
    else
        log "Phase 1: Skipped (already complete)"
    fi

    # ============================================
    # PHASE 2: Implementation
    # ============================================
    if [[ $skip_to_phase -le 2 ]]; then
        log "Phase 2: Implementation..."

        # Re-fetch issue with all comments (including our investigation comment)
        log "Fetching fresh issue data with comments..."
        local issue_context=$(fetch_issue_with_comments "$issue_number")

        local implementation_prompt=$(build_implementation_prompt "$issue_number" "$issue_title" "$issue_context" "$investigation_section" "$verifiable")
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
    else
        log "Phase 2: Skipped (already complete)"
    fi

    # ============================================
    # PHASE 3: Awaiting User Verification
    # ============================================
    log "Phase 3: Awaiting user verification..."

    gh issue edit "$issue_number" --remove-label "$LABEL_IN_PROGRESS" 2>/dev/null || true
    gh issue edit "$issue_number" --add-label "$LABEL_WAITING_USER" 2>/dev/null || true

    # Only post comment if not already posted
    if ! check_awaiting_verification "$issue_number"; then
        local verification_note=""
        if [[ "$verifiable" == "YES" ]]; then
            verification_note="The fix includes programmatic verification (build/tests pass)."
        else
            verification_note="This fix requires manual testing to verify."
        fi

        gh issue comment "$issue_number" --body "🤖 **Awaiting User Verification**

The fix has been implemented and pushed to main.

${verification_note}

Please test the fix and:
- **Close this issue** if the fix works correctly
- **Comment** if there are still problems" 2>/dev/null || true
    else
        log "Phase 3: Verification comment already exists, skipping"
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
        gh label create "$LABEL_ACCEPTED" --description "Issue accepted for bot processing" --color "0052CC" 2>/dev/null || true
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

        # Auto-accept issues from project owner
        auto_accept_owner_issues

        # Check if any waiting-for-user issues have new feedback
        check_waiting_issues_for_feedback

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
