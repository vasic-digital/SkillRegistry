# CLAUDE.md - SkillRegistry Module


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

<!-- TODO: replace this block with the exact command(s) that exercise this
     module end-to-end against real dependencies, and the expected output.
     The commands must run the real artifact (built binary, deployed
     container, real service) — no in-process fakes, no mocks, no
     `httptest.NewServer`, no Robolectric, no JSDOM as proof of done. -->

```bash
# TODO
```

> **Note:** Earlier versions of this file were a verbatim copy of `ToolSchema/CLAUDE.md` and described the wrong module (ToolHandler / Git / Test / Lint tools). That was a documentation bug. This file now describes what actually lives in this module — agent *skill* registration and lifecycle.

## Overview

`digital.vasic.skillregistry` is a Go module that provides a skill-management framework for HelixAgent and its CLI agents. A "skill" is a declarative capability an agent can register, discover, and execute — things like `code_review`, `code_generation`, `semantic_search`. Skills carry metadata (category, triggers, tags), an optional handler function, a typed parameter/return schema, declared dependencies on other skills, per-skill timeouts, and execution metrics. Storage is pluggable (in-memory is complete; PostgreSQL is stubbed).

**Module path:** `digital.vasic.skillregistry` (Go package name is `agents` — the package name intentionally does not match the directory name; don't be surprised by imports like `import agents "digital.vasic.skillregistry"`.)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
go test ./... -short              # Unit tests only
```

## Package Structure

This is a flat module — all public types live in one package `agents` at the module root. Key source files:

| File | Purpose |
|------|---------|
| `types.go` | `Skill`, `SkillDefinition`, `SkillParameter`, `SkillResult`, `SkillStatus`, `SkillCategory`, `ExecutionStatus`, `SkillMetrics`, `SkillFilter` |
| `registry.go` | `SkillManager` — register/unregister, lookup, list, filter, enable/disable, execute, metrics |
| `storage.go` | `SkillStorage` interface + `NewInMemoryStorage()` + `NewPostgresStorage()` (latter is TODO) |
| `loader.go` | `Loader` — parse skills from YAML / JSON / Markdown (`SKILL.md` with YAML frontmatter); directory and recursive scans |
| `executor.go` | `SkillExecutor` — concurrency cap, pre/post hooks, input validation, timeout enforcement |
| `validator.go` | `SkillValidator` — field validation, dependency-cycle detection via `DependencyResolver` |

## Key types and interfaces

```go
type Skill struct {
    ID          string           // lowercase alphanumeric+hyphen, 1..100 chars
    Name        string           // 1..200 chars
    Description string           // 10..5000 chars
    Version     string           // semantic version
    Category    SkillCategory    // code | data | devops | testing | security | monitoring | general
    Status      SkillStatus      // active | inactive | disabled | error
    Triggers    []string
    Tags        []string
    Author      string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Metadata    map[string]any
    ContentPath string
    Definition  SkillDefinition
    Enabled     bool             // starts false on Register; call Enable() to set Status=active + Enabled=true
}

type SkillDefinition struct {
    Parameters   []SkillParameter
    Returns      SkillReturn
    Dependencies []string        // IDs of other skills that must exist and be registered
    Permissions  []string        // currently parsed but not enforced by the validator
    Timeout      time.Duration
    Handler      string          // key into the executor's registered handler map
    Examples     []SkillExample
    Config       map[string]any
}

type SkillExecutionContext struct {
    SkillID     string
    ExecutionID string
    Inputs      map[string]any
    UserID      string
    SessionID   string
    StartedAt   time.Time
    Timeout     time.Duration
    Environment map[string]string
    Metadata    map[string]any
}

type SkillResult struct {
    ExecutionID string
    SkillID     string
    Status      ExecutionStatus   // pending | running | success | failed | cancelled | timeout
    Output      map[string]any
    Error       string
    StartedAt   time.Time
    CompletedAt time.Time
    Duration    time.Duration
    Logs        []string
    Metadata    map[string]any
}

type SkillStorage interface {
    Save(ctx, skill) error
    Get(ctx, id) (*Skill, error)
    Load(ctx, id) (*Skill, error)
    LoadByName(ctx, name) (*Skill, error)
    Delete(ctx, id) error
    List(ctx) ([]*Skill, error)
    ListByCategory(ctx, category) ([]*Skill, error)
    Search(query) []*Skill
    Exists(ctx, id) (bool, error)
    Update(ctx, skill) error
    HealthCheck(ctx) error
    Close() error
    // + a handful of in-memory-friendly helpers: Count(), Clear(), GetAll(), GetByCategory(), GetByStatus()
}

type SkillHandler   func(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error)
type ExecutionHook  func(skill *Skill, ctx *SkillExecutionContext) error
```

**Constructors you will actually call:**

```go
func NewSkillManager(storage SkillStorage) *SkillManager
func NewInMemoryStorage() SkillStorage
func NewLoader() *Loader
func NewSkillExecutor() *SkillExecutor
func NewSkillExecutorWithConcurrency(maxConcurrent int) *SkillExecutor
func NewSkillValidator() *SkillValidator
```

## Registration & execution flow

1. **Define** a `Skill` (in code, or load from YAML / JSON / `SKILL.md`).
2. **Validate** — `SkillValidator.ValidateSkill(skill)` + `ValidateSkillDependencies(skill, available)` (latter does cycle detection).
3. **Register** — `SkillManager.Register(skill)`: persists via storage, adds to in-memory map, zeros metrics. Skill starts `Enabled=false`.
4. **Load from disk** — `Loader.LoadSkillsFromDirectory(dir)` or `LoadSkillsRecursive(rootDir)` (recognizes `SKILL.md` with YAML frontmatter).
5. **Enable** — `SkillManager.Enable(skillID)` sets `Status=active`, `Enabled=true`.
6. **Discover** — `Get(id)` / `List()` / `ListByCategory(cat)` / `Search(q)` / `Filter(&SkillFilter{...})`.
7. **Execute** — `SkillManager.Execute(skillID, ctx)` or `ExecuteWithTimeout(skillID, ctx, d)`. Executor acquires a semaphore (default cap 10), runs pre-hooks, dispatches to the handler registered under `SkillDefinition.Handler` (or the echo default if none), runs post-hooks, and records metrics (`TotalExecutions`, `SuccessfulRuns`, `FailedRuns`, `AverageDuration`, `LastExecutedAt`, `LastError`).

## Integration Seams

- **Upstream:** none (foundational module).
- **Downstream:** imported by HelixLLM for CLI-agent skill coordination. The HelixAgent side consumes skills via `internal/skills/` and tests in `internal/handlers/skills_handler_test.go`, `tests/integration/skills_*_test.go`.

The module does not depend on `digital.vasic.toolschema` — the two are peers with different concerns: ToolSchema provides low-level tool-execution primitives (Read, Git, Test); SkillRegistry provides agent-level capability metadata and a registry. A skill's handler *may* drive tool calls, but skills and tools are not the same thing.

## Known gaps

- `NewPostgresStorage` is wired but the implementation body is a `TODO`. Use `NewInMemoryStorage` for now; PostgreSQL persistence is not production-ready.
- `Permissions []string` on `SkillDefinition` is not validated or enforced — it is metadata only.
- Cycle-detection error messages do not distinguish a direct A→A loop from a deeper A→B→C→A chain. Minor UX gap.
- A skill without a handler registered in the executor falls through to a default echo handler that just returns the inputs. Tests will pass against this; real skills must register a real handler via `SkillExecutor.RegisterHandler(handlerType, fn)`.

## Acceptance demo

```bash
# Run the in-tree end-to-end test that exercises registration + execution + metrics
GOMAXPROCS=2 nice -n 19 go test -race -run 'TestSkillManager_.*' ./SkillRegistry -count=1 -v

# Expected tail:
# PASS: TestSkillManager_RegisterAndExecute
# PASS: TestSkillManager_MetricsAfterExecution
# ok  	digital.vasic.skillregistry	<duration>
```

A fuller demo (loads a YAML skill from disk, enables it, executes, checks metrics) belongs in `internal/skills/` on the HelixAgent side — add it there and reference it from this block once it exists.