# Issue Bot Workflow

## Decision Diagram

```mermaid
flowchart TD
    START([Poll GitHub]) --> CHECK_ISSUES{Check for<br/>accepted issues}
    CHECK_ISSUES -->|None| CHECK_PRS{Check for<br/>accepted PRs}
    CHECK_ISSUES -->|Found| ISSUE_TYPE{Issue type?}

    %% Issue Type Branch
    ISSUE_TYPE -->|Bug| BUG_CHECK{Enough info<br/>to reproduce?}
    ISSUE_TYPE -->|Feature| DOC_CHECK{Aligns with<br/>docs/ADRs?}

    %% Bug Path
    BUG_CHECK -->|No| ASK_QUESTIONS[Ask follow-up<br/>questions]
    ASK_QUESTIONS --> ADD_WAITING[Add 'waiting-for-user'<br/>label to issue]
    ADD_WAITING --> END_CYCLE([End cycle])

    BUG_CHECK -->|Yes| CREATE_TEST[Create test case<br/>that reproduces bug]

    %% Feature Path
    DOC_CHECK -->|No| CONFLICT_COMMENT[Comment: conflicts<br/>with project direction]
    CONFLICT_COMMENT --> ADD_WAITING

    DOC_CHECK -->|Yes| CREATE_TEST_FEATURE[Create test case<br/>for expected behavior]

    %% Test Creation
    CREATE_TEST --> TEST_PASSES{Test fails as<br/>expected?}
    CREATE_TEST_FEATURE --> TEST_PASSES

    TEST_PASSES -->|No| REFINE_TEST[Refine test or<br/>ask for clarification]
    REFINE_TEST --> ADD_WAITING

    TEST_PASSES -->|Yes| CREATE_PR[Create PR with<br/>test cases]
    CREATE_PR --> LINK_PR[Comment PR link<br/>on issue]
    LINK_PR --> LABEL_ISSUE[Add 'waiting-for-user'<br/>to issue]
    LABEL_ISSUE --> END_CYCLE

    %% PR Processing
    CHECK_PRS -->|None| END_CYCLE
    CHECK_PRS -->|Found| IMPLEMENT[Implement fix<br/>to pass tests]
    IMPLEMENT --> TESTS_PASS{All tests<br/>pass?}

    TESTS_PASS -->|No| DEBUG[Debug and<br/>retry]
    DEBUG --> TESTS_PASS

    TESTS_PASS -->|Yes| PUSH_FIX[Push fix to PR]
    PUSH_FIX --> REQUEST_REVIEW[Request review]
    REQUEST_REVIEW --> END_CYCLE

    %% Styling
    classDef startEnd fill:#e1f5fe,stroke:#01579b
    classDef decision fill:#fff3e0,stroke:#e65100
    classDef action fill:#e8f5e9,stroke:#2e7d32
    classDef waiting fill:#fce4ec,stroke:#c2185b

    class START,END_CYCLE startEnd
    class CHECK_ISSUES,CHECK_PRS,ISSUE_TYPE,BUG_CHECK,DOC_CHECK,TEST_PASSES,TESTS_PASS decision
    class ASK_QUESTIONS,CREATE_TEST,CREATE_TEST_FEATURE,REFINE_TEST,CREATE_PR,LINK_PR,IMPLEMENT,DEBUG,PUSH_FIX,REQUEST_REVIEW,CONFLICT_COMMENT action
    class ADD_WAITING,LABEL_ISSUE waiting
```

## Sequence Diagram

```mermaid
sequenceDiagram
    participant U as User
    participant GH as GitHub
    participant Bot as Issue Bot
    participant Claude as Claude Code
    participant Repo as Repository

    Note over U,Repo: Phase 1: Issue Triage & Reproduction

    U->>GH: Create issue (bug/feature)
    U->>GH: Add 'accepted' label

    loop Poll every 15s
        Bot->>GH: Check for accepted issues
    end

    GH-->>Bot: Issue #123 (accepted)
    Bot->>GH: Fetch issue + comments

    alt Bug Report
        Bot->>Claude: Analyze bug report
        Claude-->>Bot: Need more info?

        alt Insufficient info
            Bot->>GH: Comment: ask follow-up questions
            Bot->>GH: Add 'waiting-for-user' label
            U->>GH: Reply with details
            GH-->>Bot: New comment detected
            Bot->>GH: Remove 'waiting-for-user'
        end

        Bot->>Claude: Create failing test case
        Claude->>Repo: Write test that reproduces bug

    else Feature Request
        Bot->>Claude: Check against docs/ADRs
        Claude-->>Bot: Aligned?

        alt Conflicts with docs
            Bot->>GH: Comment: explain conflict
            Bot->>GH: Add 'waiting-for-user' label
        else Aligned
            Bot->>Claude: Create test for expected behavior
            Claude->>Repo: Write test for new feature
        end
    end

    Note over U,Repo: Phase 2: PR Creation

    Bot->>Repo: Create branch 'issue-123-tests'
    Bot->>Repo: Commit test cases
    Bot->>GH: Create PR (tests only)
    Bot->>GH: Comment on issue: "PR #456 created"
    Bot->>GH: Add 'waiting-for-user' to issue

    U->>GH: Review PR
    U->>GH: Add 'accepted' label to PR

    Note over U,Repo: Phase 3: Implementation

    loop Poll every 15s
        Bot->>GH: Check for accepted PRs
    end

    GH-->>Bot: PR #456 (accepted)
    Bot->>Claude: Implement fix to pass tests

    loop Until tests pass
        Claude->>Repo: Write implementation
        Claude->>Repo: Run tests
        alt Tests fail
            Claude->>Repo: Debug and fix
        end
    end

    Bot->>Repo: Push implementation commits
    Bot->>GH: Request review on PR

    U->>GH: Approve & merge PR
    GH->>GH: Auto-close issue #123
```

## State Transitions

```mermaid
stateDiagram-v2
    [*] --> Open: Issue created

    Open --> Accepted: User adds 'accepted' label

    Accepted --> Investigating: Bot starts analysis

    Investigating --> WaitingForUser: Need more info
    Investigating --> TestCreation: Have enough info

    WaitingForUser --> Investigating: User replies

    TestCreation --> PRCreated: Tests written

    PRCreated --> WaitingForUser: Add label to issue

    state PRCreated {
        [*] --> PROpen
        PROpen --> PRAccepted: User adds 'accepted'
        PRAccepted --> Implementing: Bot starts fix
        Implementing --> PRReady: Tests pass
        PRReady --> PRMerged: User merges
    }

    PRMerged --> Closed: Auto-close issue

    Closed --> [*]
```

## Labels Used

| Label | Applied To | Meaning |
|-------|-----------|---------|
| `accepted` | Issue | Issue approved for bot processing |
| `accepted` | PR | PR approved for bot to implement fix |
| `waiting-for-user` | Issue | Bot waiting for user response/review |
| `bot-in-progress` | Issue/PR | Bot actively working |
| `bot-test-pr` | PR | PR contains only test cases (no fix yet) |
| `bot-failed` | Issue/PR | Bot encountered unrecoverable error |

## Key Changes from Current Bot

1. **Two-phase approach**: Test cases first, implementation second
2. **PR-centric workflow**: Focus moves from issue to PR once tests exist
3. **Explicit acceptance gates**: Both issues AND PRs need `accepted` label
4. **Documentation alignment check**: Features validated against ADRs/README
5. **Follow-up questions**: Bot can ask for clarification on bugs
6. **Issue-PR linking**: PR referenced in issue comment, issue gets `waiting-for-user`
