# CI and Testing Strategy

**Status:** Accepted

## Context

Need automated testing in CI. Also need a way to ensure comprehensive test coverage through defined test personas that guide test generation.

## Options Considered

### CI Platform

**GitHub Actions:**
- Native to GitHub, no external service
- Free for public repos, generous free tier for private
- YAML workflow files in `.github/workflows/`
- Matrix builds for multiple Go versions/OS

**GitLab CI:**
- Good if using GitLab
- Similar YAML config
- Would require mirroring or migration

**CircleCI / Travis CI:**
- External services
- Additional account/config
- No clear advantage over GitHub Actions

**Decision:** GitHub Actions. Native integration, no external dependencies.

### Test Organization

**Flat test files:**
```
internal/collision/tiles_test.go
internal/collision/aabb_test.go
```

**Separate test directory:**
```
test/
  unit/
  integration/
  e2e/
```

**Decision:** Go convention - `*_test.go` files alongside code, with `integration/` folder for cross-package tests.

## Decision

### GitHub Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go build ./...
```

### Test Personas

Personas define perspectives for test generation. Each persona asks different questions about the code.

#### 1. The Happy Path User
**Focus:** Normal usage, expected inputs
**Questions:**
- Does the basic functionality work?
- Do default values behave correctly?
- Does the common workflow succeed?

**Example tests:**
```go
func TestCollision_PlayerOnGround(t *testing.T)
func TestInput_WalkRight(t *testing.T)
func TestRoom_CreateAndJoin(t *testing.T)
```

#### 2. The Edge Case Explorer
**Focus:** Boundaries, limits, transitions
**Questions:**
- What happens at zero? At max values?
- What about empty inputs? Single items? Exactly at limits?
- State transitions and boundary conditions?

**Example tests:**
```go
func TestCollision_AtMapEdge(t *testing.T)
func TestCollision_ExactlyOnTileBoundary(t *testing.T)
func TestSnapshot_EmptyEntityList(t *testing.T)
func TestSnapshot_MaxEntities(t *testing.T)
func TestRoom_CodeAtCharsetBoundary(t *testing.T)
```

#### 3. The Saboteur
**Focus:** Invalid inputs, error conditions
**Questions:**
- What if input is nil? Negative? Wrong type?
- What if network fails mid-operation?
- What if state is corrupted?

**Example tests:**
```go
func TestCollision_NilTileMap(t *testing.T)
func TestInput_InvalidIntent(t *testing.T)
func TestSnapshot_MalformedDelta(t *testing.T)
func TestRoom_LookupNonexistent(t *testing.T)
func TestTransport_ConnectionDropped(t *testing.T)
```

#### 4. The Concurrency Tester
**Focus:** Race conditions, deadlocks, ordering
**Questions:**
- What if two goroutines access this simultaneously?
- Is this operation atomic when it needs to be?
- Can operations happen out of order?

**Example tests:**
```go
func TestServer_ConcurrentPlayerJoin(t *testing.T)
func TestRoom_ConcurrentCreateLookup(t *testing.T)
func TestSnapshot_ConcurrentReadWrite(t *testing.T)
func TestInput_ConcurrentBuffer(t *testing.T)
```

#### 5. The Performance Skeptic
**Focus:** Benchmarks, allocations, scaling
**Questions:**
- Does it allocate unnecessarily?
- Does it scale with input size?
- Are hot paths optimized?

**Example tests:**
```go
func BenchmarkCollision_CheckTile(b *testing.B)
func BenchmarkSnapshot_DeltaCompression(b *testing.B)
func BenchmarkRender_HalfBlock100Entities(b *testing.B)
```

#### 6. The Replayer
**Focus:** Determinism, reproducibility
**Questions:**
- Given same inputs, do we get same outputs?
- Can we replay a sequence and get identical results?
- Is random state seeded correctly?

**Example tests:**
```go
func TestGame_DeterministicTick(t *testing.T)
func TestCollision_ReplaySequence(t *testing.T)
func TestRoom_CodeGenerationSeeded(t *testing.T)
```

### Test File Structure

```
internal/
  collision/
    tiles.go
    tiles_test.go          # Unit tests
    aabb.go
    aabb_test.go
  server/
    server.go
    server_test.go
  ...

integration/
  client_server_test.go    # Cross-package integration
  lobby_flow_test.go
  game_session_test.go
```

### Persona Coverage Matrix

When generating tests, ensure each package has coverage from relevant personas:

| Package | Happy | Edge | Saboteur | Concurrency | Perf | Replay |
|---------|-------|------|----------|-------------|------|--------|
| collision | ✓ | ✓ | ✓ | - | ✓ | ✓ |
| input | ✓ | ✓ | ✓ | ✓ | - | ✓ |
| network | ✓ | ✓ | ✓ | ✓ | ✓ | - |
| sync | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| lobby | ✓ | ✓ | ✓ | ✓ | - | - |
| server | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| render | ✓ | ✓ | ✓ | - | ✓ | - |

### Test Generation Prompt Template

When asking AI to generate tests:

```
Generate tests for [package/function] using these personas:

1. Happy Path: Test normal usage with valid inputs
2. Edge Cases: Test boundaries (zero, max, empty, single)
3. Saboteur: Test invalid inputs and error conditions
4. Concurrency: Test parallel access (if applicable)
5. Performance: Add benchmarks for hot paths
6. Replay: Verify deterministic behavior (if applicable)

Existing code: [paste code]
```

## Consequences

- **Automated testing** - Every PR runs tests automatically
- **Race detection** - `-race` flag catches concurrency bugs
- **Structured coverage** - Personas ensure we don't miss test categories
- **AI-friendly** - Persona framework guides test generation prompts
- **Living documentation** - Coverage matrix shows what's tested
