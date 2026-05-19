# SkillRegistry Module

The SkillRegistry module provides a comprehensive skill management system for HelixAgent, enabling registration, execution, and management of AI skills.

## Overview

The SkillRegistry module is designed to:
- Load skills from various formats (YAML, JSON, Markdown with YAML frontmatter)
- Register and manage skills with metadata
- Execute skills with context and timeout support
- Validate skill definitions and dependencies
- Store skills in memory or PostgreSQL

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SkillManager                              │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │  Loader  │  │ Validator│  │ Executor │  │  Storage   │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
    ┌─────────┐       ┌──────────┐      ┌──────────┐
    │  Files  │       │   DB     │      │  Memory  │
    │(YAML/   │       │(PostgreSQL)│     │          │
    │Markdown)│       │          │      │          │
    └─────────┘       └──────────┘      └──────────┘
```

## Components

### Types (`types.go`)
Core data structures:
- `Skill` - Main skill definition with metadata
- `SkillDefinition` - Execution parameters and configuration
- `SkillExecutionContext` - Runtime context for execution
- `SkillResult` - Execution result with logs and metadata
- `SkillFilter` - Filter criteria for skill queries

### Loader (`loader.go`)
Skill loading functionality:
- `LoadSkillFromFile(path)` - Load single skill from file
- `LoadSkillsFromDirectory(dir)` - Load all skills from directory
- `ParseSkillYAML/JSON` - Parse skill from data
- `LoadSkillsRecursive` - Recursively load skills

Supports formats:
- YAML files (.yaml, .yml)
- JSON files (.json)
- Markdown with YAML frontmatter (SKILL.md)

### Validator (`validator.go`)
Skill validation:
- `ValidateSkill(skill)` - Validate skill structure
- `ValidateSkillDependencies` - Check dependency graph
- `ValidateBatch` - Validate multiple skills

Validates:
- Required fields (ID, name, description)
- ID format (lowercase, alphanumeric with hyphens/underscores)
- Version (semantic versioning)
- Category (predefined values)
- Parameters and types
- Circular dependencies

### Executor (`executor.go`)
Skill execution:
- `Execute(skill, ctx)` - Execute skill
- `ExecuteWithTimeout` - Execute with timeout
- `RegisterHandler` - Register custom handlers
- `AddPre/PostExecutionHook` - Add execution hooks

Features:
- Concurrent execution with semaphore
- Pre/post execution hooks
- Custom handlers
- Input validation
- Execution metrics

### Manager (`manager.go`)
High-level skill management:
- `Register/Unregister` - Add/remove skills
- `Get/List/Search/Filter` - Query skills
- `Enable/Disable` - Toggle skill status
- `Execute` - Execute skills
- `LoadFromDirectory/File` - Bulk loading

### Storage

#### In-Memory (`storage_memory.go`)
- `InMemoryStorage` - Thread-safe in-memory storage
- Good for testing and caching

#### PostgreSQL (`storage_postgres.go`)
- `PostgresStorage` - Persistent PostgreSQL storage
- Supports JSONB for flexible metadata
- Automatic table creation

## Usage

### Basic Usage

```go
package main

import (
    skillregistry "dev.helix.agent/SkillRegistry"
)

func main() {
    // Create manager with in-memory storage
    manager := skillregistry.NewSkillManager(nil)

    // Load skills from directory
    err := manager.LoadFromDirectory("./skills")
    if err != nil {
        panic(err)
    }

    // List all skills
    skills := manager.List()
    for _, skill := range skills {
        fmt.Printf("Skill: %s - %s\n", skill.Name, skill.Description)
    }

    // Enable a skill
    err = manager.Enable("my-skill")
    if err != nil {
        panic(err)
    }

    // Execute a skill
    ctx := skillregistry.NewSkillExecutionContext("my-skill")
    ctx.Inputs = map[string]interface{}{
        "param1": "value1",
    }
    
    result, err := manager.Execute("my-skill", ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Result: %v\n", result.Output)
}
```

### Loading Skills

```go
loader := skillregistry.NewLoader()

// Load single skill
skill, err := loader.LoadSkillFromFile("./skills/my-skill.yaml")

// Load from directory
skills, err := loader.LoadSkillsFromDirectory("./skills")

// Load recursively
skills, err := loader.LoadSkillsRecursive("./skills")
```

### Skill Definition (YAML)

```yaml
name: my-skill
description: A sample skill that does something useful
version: 1.0.0
category: code
tags:
  - example
  - utility
triggers:
  - /my-trigger
author: HelixAgent Team

parameters:
  - name: input_text
    type: string
    description: Text to process
    required: true
  
  - name: options
    type: object
    description: Processing options
    required: false
    default: {}

returns:
  type: object
  description: Processing result

dependencies:
  - other-skill

timeout: 30s
```

### Skill Definition (Markdown with Frontmatter)

```markdown
---
name: my-skill
description: A skill defined in Markdown
triggers:
  - /my-trigger
---

# My Skill

Detailed documentation here...
```

### Custom Execution Handler

```go
manager := skillregistry.NewSkillManager(nil)

// Register custom handler
manager.RegisterHandler("custom", func(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
    result := skillregistry.NewSkillResult(ctx.ExecutionID, skill.ID)
    
    // Custom logic here
    output := processInputs(ctx.Inputs)
    
    return result.Success(output), nil
})

// Use handler in skill definition
skill.Definition = &SkillDefinition{
    Handler: "custom",
}
```

### Pre/Post Execution Hooks

```go
// Add logging hook
manager.AddPreExecutionHook(func(skill *Skill, ctx *SkillExecutionContext) error {
    log.Printf("Starting execution of %s", skill.Name)
    return nil
})

// Add validation hook
manager.AddPostExecutionHook(func(skill *Skill, ctx *SkillExecutionContext) error {
    log.Printf("Completed execution of %s", skill.Name)
    return nil
})
```

### PostgreSQL Storage

```go
config := &skillregistry.StorageConfig{
    Type:     "postgres",
    Host:     "localhost",
    Port:     5432,
    Database: "helixagent",
    Username: "user",
    Password: "password",
    SSLMode:  "disable",
}

storage, err := skillregistry.NewPostgresStorage(config)
if err != nil {
    panic(err)
}

manager := skillregistry.NewSkillManager(storage)
```

### Filtering Skills

```go
// Filter by category
codeSkills := manager.ListByCategory(skillregistry.SkillCategoryCode)

// Search by name/description
results := manager.Search("database")

// Advanced filter
filter := &skillregistry.SkillFilter{
    Category:    skillregistry.SkillCategoryDevOps,
    Enabled:     boolPtr(true),
    Tags:        []string{"kubernetes", "docker"},
    SearchQuery: "deploy",
}
filtered := manager.Filter(filter)
```

## Error Handling

Common errors:
- `ErrSkillNotFound` - Skill doesn't exist
- `ErrSkillAlreadyExists` - Duplicate skill ID
- `ErrSkillInvalid` - Invalid skill definition
- `ErrSkillDisabled` - Skill is disabled
- `ErrSkillTimeout` - Execution timed out
- `ErrCircularDependency` - Circular dependency detected

## Testing

Run tests:
```bash
cd SkillRegistry
GOMAXPROCS=2 nice -n 19 go test -race -count=1 -v ./...

# With coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Anti-bluff guarantees (Article XI §11.9 / CONST-035 / CONST-048)

> Verbatim 2026-05-19 operator mandate: *"all existing tests and
> Challenges do work in anti-bluff manner — they MUST confirm that
> all tested codebase really works as expected! We had been in
> position that all tests do execute with success and all Challenges
> as well, but in reality the most of the features does not work and
> can't be used! This MUST NOT be the case and execution of tests
> and Challenges MUST guarantee the quality, the completition and
> full usability by end users of the product!"*

Round-282 enrichment closes the historical "green tests, broken
product" failure mode by enforcing six invariants on every PASS:

1. **Real fixtures on disk** — the runner reads 5 real YAML files
   from `challenges/fixtures/` via `loader.LoadSkillsFromDirectory`.
   No embedded literal corpus, no synthetic in-memory generation.
2. **Real validator dispatch** — `SkillValidator.ValidateSkill` runs
   against every loaded fixture; rejection paths covered by the
   paired-mutation Challenge (empty ID → exit 99).
3. **Real Register → Enable → Execute pipeline** — the runner wires
   a real `SkillManager` over a real `InMemoryStorage` with a real
   `SkillExecutor`, registers a real handler (`challenge.exerciser`),
   then executes every skill and asserts `Status == Success`,
   `Output != nil`, `Duration >= 0`.
4. **Real storage round-trip** — after Register, the runner reads
   the storage interface back (`storage.List`) and asserts the count
   matches. This proves Register actually persisted, not just cached.
5. **Real discovery surfaces** — `Search`, `Filter(Enabled=true)`,
   `ListByCategory(general)` are invoked and asserted against the
   loaded corpus. Each query path is real, not a stub.
6. **5-locale bilingual UX** (CONST-046) — the runner emits
   `en/sr/ja/es/de` summary lines so non-English consumers get
   first-class operator output.

A bluff in any of these layers (silent-skip, swallowed-error, stub
handler) is caught by the paired-mutation Challenge:

```bash
# Default mode — must exit 0 with the captured runtime evidence
bash challenges/skillregistry_describe_challenge.sh

# Mutate mode — plants empty-ID fixture; correct rejection exits 99,
# silent acceptance exits 0 (= Challenge is itself a bluff).
bash challenges/skillregistry_describe_challenge.sh --mutate
```

### Test-coverage ledger

See [`docs/test-coverage.md`](docs/test-coverage.md) for the full
symbol → test / Challenge mapping. Every exported symbol of the
`agents` package is listed alongside the assertions that exercise
it and the anti-bluff dimension each proves. Adding a public symbol
without updating this ledger is a CONST-048 violation.

### Challenge runner — direct invocation

```bash
go run ./challenges/runner -all
```

Expected tail (verbatim, modulo durations):

```
loaded_skills=5 source=challenges/fixtures
skill=challenge-skill-de status=success duration=...
skill=challenge-skill-en status=success duration=...
skill=challenge-skill-es status=success duration=...
skill=challenge-skill-ja status=success duration=...
skill=challenge-skill-sr status=success duration=...
[en] skillregistry: 5 skill(s) registered, 5 execution(s) succeeded
[sr] skillregistry: 5 veština registrovano, 5 izvršavanja uspešno
[ja] skillregistry: 5 個のスキルが登録、5 回の実行に成功
[es] skillregistry: 5 habilidad(es) registrada(s), 5 ejecución(es) exitosa(s)
[de] skillregistry: 5 Fertigkeit(en) registriert, 5 Ausführung(en) erfolgreich
OK skills=5 executions=5 locales=5
```

### Exit-code map

| Code | Meaning                                                                              |
|------|--------------------------------------------------------------------------------------|
| 0    | Every step succeeded; runtime evidence captured on stdout.                           |
| 1    | Usage / flag error.                                                                  |
| 2    | Coverage gap (loader returned 0 skills, registry count drifted, discovery mismatch). |
| 3    | Schema-invariant violation (validator rejected a fixture or post-Register drift).    |
| 4    | Execution invariant violation (Status not Success, missing Output, locale missing).  |
| 99   | (Mutate mode only) Mutation correctly surfaced — Challenge is honest.                |

## Honest known gaps

These are gaps the SkillRegistry has today; the runner does NOT
pretend they are fixed (per CONST-035 — pretending a gap is closed
is the exact bluff this round-282 enrichment guards against):

- **Metrics counters are not incremented by `Execute`.** A
  `SkillMetrics` record is allocated at `Register` time and
  retrievable via `GetMetrics`, but the `TotalExecutions /
  SuccessfulRuns / FailedRuns / AverageDuration` counters stay at
  zero. The runner asserts the metrics surface (record exists,
  SkillID matches, no negative counters) but does NOT assert
  `TotalExecutions >= 1` — that would be a §11.4 PASS-bluff against
  reality. When the implementation gains real counter wiring, the
  runner's assertion in `runAll()` should be tightened in the SAME
  commit.
- **`Permissions []string` on `SkillDefinition`** is parsed but not
  enforced by the validator. Metadata-only today.
- **`PostgresStorage` integration** requires a running PostgreSQL;
  the runner uses `InMemoryStorage`. The Postgres path is covered
  by `TestPostgresStorage_*` at the integration tier, which skips
  with `SKIP-OK:` when no DB is reachable.

## Integration with HelixAgent

The SkillRegistry integrates with:
- **MCP Module**: Skills can be exposed as MCP tools
- **Agentic Module**: Skills can be used in agent workflows
- **LLMOrchestrator**: Skills can be invoked by LLMs

## License

Part of the HelixAgent project.
