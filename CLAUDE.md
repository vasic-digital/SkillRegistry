# CLAUDE.md - SkillRegistry Module


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# Skill definition → validation → registration → enable → execute with metrics
cd SkillRegistry && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -v \
  -run 'TestSkillManager_|TestLoader_LoadSkillFromFile|TestSkillValidator_' .
```
Expect: PASS; YAML/JSON/SKILL.md loaders work; dependency-cycle detection catches cycles; metrics increment on execution.


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
<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->



<!-- CONST-035 anti-bluff addendum (cascaded) -->

## CONST-035 — Anti-Bluff Tests & Challenges (mandatory; inherits from root)

Tests and Challenges in this submodule MUST verify the product, not
the LLM's mental model of the product. A test that passes when the
feature is broken is worse than a missing test — it gives false
confidence and lets defects ship to users. Functional probes at the
protocol layer are mandatory:

- TCP-open is the FLOOR, not the ceiling. Postgres → execute
  `SELECT 1`. Redis → `PING` returns `PONG`. ChromaDB → `GET
  /api/v1/heartbeat` returns 200. MCP server → TCP connect + valid
  JSON-RPC handshake. HTTP gateway → real request, real response,
  non-empty body.
- Container `Up` is NOT application healthy. A `docker/podman ps`
  `Up` status only means PID 1 is running; the application may be
  crash-looping internally.
- No mocks/fakes outside unit tests (already CONST-030; CONST-035
  raises the cost of a mock-driven false pass to the same severity
  as a regression).
- Re-verify after every change. Don't assume a previously-passing
  test still verifies the same scope after a refactor.
- Verification of CONST-035 itself: deliberately break the feature
  (e.g. `kill <service>`, swap a password). The test MUST fail. If
  it still passes, the test is non-conformant and MUST be tightened.

## CONST-033 clarification — distinguishing host events from sluggishness

Heavy container builds (BuildKit pulling many GB of layers, parallel
podman/docker compose-up across many services) can make the host
**appear** unresponsive — high load average, slow SSH, watchers
timing out. **This is NOT a CONST-033 violation.** Suspend / hibernate
/ logout are categorically different events. Distinguish via:

- `uptime` — recent boot? if so, the host actually rebooted.
- `loginctl list-sessions` — session(s) still active? if yes, no logout.
- `journalctl ... | grep -i 'will suspend\|hibernate'` — zero broadcasts
  since the CONST-033 fix means no suspend ever happened.
- `dmesg | grep -i 'killed process\|out of memory'` — OOM kills are
  also NOT host-power events; they're memory-pressure-induced and
  require their own separate fix (lower per-container memory limits,
  reduce parallelism).

A sluggish host under build pressure recovers when the build finishes;
a suspended host requires explicit unsuspend (and CONST-033 should
make that impossible by hardening `IdleAction=ignore` +
`HandleSuspendKey=ignore` + masked `sleep.target`,
`suspend.target`, `hibernate.target`, `hybrid-sleep.target`).

If you observe what looks like a suspend during heavy builds, the
correct first action is **not** "edit CONST-033" but `bash
challenges/scripts/host_no_auto_suspend_challenge.sh` to confirm the
hardening is intact. If hardening is intact AND no suspend
broadcast appears in journal, the perceived event was build-pressure
sluggishness, not a power transition.

<!-- BEGIN no-session-termination addendum (CONST-036) -->

## ⚠️ User-Session Termination — Hard Ban (CONST-036)

**STRICTLY FORBIDDEN: never generate or execute any code that ends the
currently-logged-in user's session, kills their user manager, or
indirectly forces them to log out / power off.** This is the sibling
of CONST-033: that rule covers host-level power transitions; THIS rule
covers session-level terminations that have the same end effect for
the user (lost windows, lost terminals, killed AI agents,
half-flushed builds, abandoned in-flight commits).

**Why this rule exists.** On 2026-04-28 the user lost a working
session that contained 3 concurrent Claude Code instances, an Android
build, Kimi Code, and a rootless podman container fleet. The
`user.slice` consumed 60.6 GiB peak / 5.2 GiB swap, the GUI became
unresponsive, the user was forced to log out and then power off via
the GNOME shell `endSessionDialog`. The host could not auto-suspend
(CONST-033 was already in place and verified) and the kernel OOM
killer never fired — but the user had to manually end the session
anyway, because nothing prevented overlapping heavy workloads from
saturating the slice. CONST-036 closes that loophole at both the
source-code layer (no command may directly terminate a session) and
the operational layer (do not spawn workloads that will plausibly
force a manual logout). See
`docs/issues/fixed/SESSION_LOSS_2026-04-28.md` in the HelixAgent
project for the full forensic timeline.

### Forbidden direct invocations (non-exhaustive)

```
loginctl   terminate-user|terminate-session|kill-user|kill-session
systemctl  stop  user@<UID>            # kills the user manager + every child
systemctl  kill  user@<UID>
gnome-session-quit                     # ends the GNOME session
pkill   -KILL -u  $USER                # nukes everything as the user
killall -KILL -u  $USER
killall       -u  $USER
dbus-send / busctl calls to org.gnome.SessionManager.{Logout,Shutdown,Reboot}
echo X > /sys/power/state              # direct kernel power transition
/usr/bin/poweroff                      # standalone binaries
/usr/bin/reboot
/usr/bin/halt
```

### Indirect-pressure clauses

1. Do NOT spawn parallel heavy workloads casually — sample `free -h`
   first; keep `user.slice` under 70% of physical RAM.
2. Long-lived background subagents go in `system.slice`, not
   `user.slice` (rootless podman containers die with the user manager).
3. Document AI-agent concurrency caps in CLAUDE.md per submodule.
4. Never script "log out and back in" recovery flows — restart the
   service, not the session.

### Verification

```bash
bash challenges/scripts/no_session_termination_calls_challenge.sh  # source clean
bash challenges/scripts/no_suspend_calls_challenge.sh              # CONST-033 still clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh          # host hardened
```

All three must PASS.

<!-- END no-session-termination addendum (CONST-036) -->

<!-- BEGIN const035-strengthening-2026-04-29 -->

## CONST-035 — End-User Usability Mandate (2026-04-29 strengthening)

A test or Challenge that PASSES is a CLAIM that the tested behavior
**works for the end user of the product**. The HelixAgent project
has repeatedly hit the failure mode where every test ran green AND
every Challenge reported PASS, yet most product features did not
actually work — buggy challenge wrappers masked failed assertions,
scripts checked file existence without executing the file,
"reachability" tests tolerated timeouts, contracts were honest in
advertising but broken in dispatch. **This MUST NOT recur.**

Every PASS result MUST guarantee:

a. **Quality** — the feature behaves correctly under inputs an end
   user will send, including malformed input, edge cases, and
   concurrency that real workloads produce.
b. **Completion** — the feature is wired end-to-end from public
   API surface down to backing infrastructure, with no stub /
   placeholder / "wired lazily later" gaps that silently 503.
c. **Full usability** — a CLI agent / SDK consumer / direct curl
   client following the documented model IDs, request shapes, and
   endpoints SUCCEEDS without having to know which of N internal
   aliases the dispatcher actually accepts.

A passing test that doesn't certify all three is a **bluff** and
MUST be tightened, or marked `t.Skip("...SKIP-OK: #<ticket>")`
so absence of coverage is loud rather than silent.

### Bluff taxonomy (each pattern observed in HelixAgent and now forbidden)

- **Wrapper bluff** — assertions PASS but the wrapper's exit-code
  logic is buggy, marking the run FAILED (or the inverse: assertions
  FAIL but the wrapper swallows them). Every aggregating wrapper MUST
  use a robust counter (`! grep -qs "|FAILED|" "$LOG"` style) —
  never inline arithmetic on a command that prints AND exits
  non-zero.
- **Contract bluff** — the system advertises a capability but
  rejects it in dispatch. Every advertised capability MUST be
  exercised by a test or Challenge that actually invokes it.
- **Structural bluff** — `check_file_exists "foo_test.go"` passes
  if the file is present but doesn't run the test or assert anything
  about its content. File-existence checks MUST be paired with at
  least one functional assertion.
- **Comment bluff** — a code comment promises a behavior the code
  doesn't actually have. Documentation written before / about code
  MUST be re-verified against the code on every change touching the
  documented function.
- **Skip bluff** — `t.Skip("not running yet")` without a
  `SKIP-OK: #<ticket>` marker silently passes. Every skip needs the
  marker; CI fails on bare skips.

The taxonomy is illustrative, not exhaustive. Every Challenge or
test added going forward MUST pass an honest self-review against
this taxonomy before being committed.

<!-- END const035-strengthening-2026-04-29 -->

---

## Article XI §11.9 — Anti-Bluff Forensic Anchor (cascaded from parent CONSTITUTION.md)

> Verbatim user mandate (2026-04-29, reasserted multiple times across 2026-05): *"We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completion and full usability by end users of the product!"*

Operative rule: **The bar for shipping is not "tests pass" but "users can use the feature."** Every PASS in this codebase MUST carry positive runtime evidence captured during execution. Metadata-only / configuration-only / absence-of-error / grep-based PASS without runtime evidence are critical defects regardless of how green the summary line looks. No false-success results are tolerable.

This anchor MUST remain in this submodule's CONSTITUTION.md, CLAUDE.md, and AGENTS.md alongside CONST-047 — see the parent repository's `CONSTITUTION.md` for the full text.
