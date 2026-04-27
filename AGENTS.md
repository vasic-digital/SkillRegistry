# SkillRegistry Developer Guide

This document provides guidelines for developers working on the SkillRegistry module.

## Module Structure

```
SkillRegistry/
├── types.go              # Core types and error definitions
├── loader.go             # Skill loading from files
├── executor.go           # Skill execution engine
├── validator.go          # Skill validation
├── manager.go            # Skill management
├── storage.go            # Storage interface
├── storage_memory.go     # In-memory storage implementation
├── storage_postgres.go   # PostgreSQL storage implementation
├── registry.go           # CLI agent registry (existing)
├── *_test.go             # Test files
├── README.md             # User documentation
└── AGENTS.md             # This file
```

## Adding New Features

### Adding a New Skill Category

1. Add the category constant in `types.go`:
```go
const (
    // ... existing categories
    CategoryNewCategory SkillCategory = "new_category"
)
```

2. Update `IsValidCategory()` function to include the new category

3. Update `AllCategories()` function

4. Add tests in `types_test.go`

### Adding a New Storage Backend

1. Create a new file (e.g., `storage_redis.go`)

2. Implement the `SkillStorage` interface:
```go
type RedisStorage struct {
    client *redis.Client
    config *StorageConfig
}

func NewRedisStorage(addr string) (*RedisStorage, error) {
    // Implementation
}

func (s *RedisStorage) Save(skill *Skill) error {
    // Implementation
}
// ... implement all methods
```

3. Add comprehensive tests in `storage_test.go`

### Adding Execution Hooks

Create a hook that implements the `ExecutionHook` interface:

```go
type MyHook struct {
    // Your fields
}

func (h *MyHook) BeforeExecution(ctx *SkillExecutionContext) error {
    // Pre-execution logic
    return nil
}

func (h *MyHook) AfterExecution(ctx *SkillExecutionContext, result *SkillResult) error {
    // Post-execution logic
    return nil
}
```

Register the hook:

```go
executor := NewSkillExecutor()
executor.AddHook(&MyHook{})
```

## Testing Guidelines

### Unit Tests

- Test each function in isolation
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >90% coverage

Example:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid case", "input", "output", false},
        {"error case", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

### Integration Tests

- Test with real dependencies when possible
- Use `t.TempDir()` for file operations
- Clean up resources after tests

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestFunctionName ./...

# Run with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

### Naming Conventions

- Use `PascalCase` for exported identifiers
- Use `camelCase` for unexported identifiers
- Use `SCREAMING_SNAKE_CASE` for constants
- Acronyms: all caps (`HTTP`, `URL`, `ID`)

### Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Use sentinel errors for common cases
- Check errors immediately after they occur

Example:

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### Struct Tags

Use consistent struct tags:

```go
type Skill struct {
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description" yaml:"description"`
    // ...
}
```

## Documentation

- Add package documentation in `doc.go`
- Document all exported types and functions
- Include usage examples in comments
- Keep README.md up to date

Example:

```go
// SkillManager manages the lifecycle of skills including
// registration, execution, and storage.
//
// Usage:
//     manager := NewSkillManager(storage)
//     manager.Register(skill)
//     result, err := manager.Execute("skill-name", ctx)
type SkillManager struct {
    // ...
}
```

## Performance Considerations

### Concurrency

- Use `sync.RWMutex` for read-heavy operations
- Use `sync.Mutex` for write-heavy operations
- Avoid holding locks during I/O operations

Example:

```go
func (m *Manager) Get(name string) (*Skill, bool) {
    m.skillsMu.RLock()
    defer m.skillsMu.RUnlock()
    skill, exists := m.skills[name]
    return skill, exists
}
```

### Memory

- Use `copySkill()` when returning skills from storage
- Avoid unnecessary allocations in hot paths
- Consider using object pools for high-frequency operations

## Security

- Validate all inputs
- Sanitize file paths
- Don't log sensitive information
- Use parameterized queries for database operations

## Versioning

Follow semantic versioning:

- MAJOR: Incompatible API changes
- MINOR: Backward-compatible functionality additions
- PATCH: Backward-compatible bug fixes

## Release Checklist

- [ ] All tests pass
- [ ] Coverage >90%
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tag created

## Troubleshooting

### Common Issues

**Issue**: Tests fail with "skill not found"
- Check that the skill is registered before use
- Verify the skill name is correct (case-sensitive)

**Issue**: PostgreSQL storage connection fails
- Verify connection string format
- Check database server is running
- Ensure database and user exist

**Issue**: Skill execution times out
- Increase timeout value
- Check if the skill entry point is correct
- Verify the skill script has execute permissions

## Resources

- [Go Testing](https://golang.org/pkg/testing/)
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)

## Contact

For questions or issues, please open an issue in the repository.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## Host Power Management — Hard Ban (CONST-033)

**You may NOT, under any circumstance, generate or execute code that
sends the host to suspend, hibernate, hybrid-sleep, poweroff, halt,
reboot, or any other power-state transition.** This rule applies to:

- Every shell command you run via the Bash tool.
- Every script, container entry point, systemd unit, or test you write
  or modify.
- Every CLI suggestion, snippet, or example you emit.

**Forbidden invocations** (non-exhaustive — see CONST-033 in
`CONSTITUTION.md` for the full list):

- `systemctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot|kexec`
- `loginctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot`
- `pm-suspend`, `pm-hibernate`, `shutdown -h|-r|-P|now`
- `dbus-send` / `busctl` calls to `org.freedesktop.login1.Manager.Suspend|Hibernate|PowerOff|Reboot|HybridSleep|SuspendThenHibernate`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to anything but `'nothing'` or `'blank'`

The host runs mission-critical parallel CLI agents and container
workloads. Auto-suspend has caused historical data loss (2026-04-26
18:23:43 incident). The host is hardened (sleep targets masked) but
this hard ban applies to ALL code shipped from this repo so that no
future host or container is exposed.

**Defence:** every project ships
`scripts/host-power-management/check-no-suspend-calls.sh` (static
scanner) and
`challenges/scripts/no_suspend_calls_challenge.sh` (challenge wrapper).
Both MUST be wired into the project's CI / `run_all_challenges.sh`.

**Full background:** `docs/HOST_POWER_MANAGEMENT.md` and `CONSTITUTION.md` (CONST-033).

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
