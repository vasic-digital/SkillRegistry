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
go test -v ./...

# With coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Integration with HelixAgent

The SkillRegistry integrates with:
- **MCP Module**: Skills can be exposed as MCP tools
- **Agentic Module**: Skills can be used in agent workflows
- **LLMOrchestrator**: Skills can be invoked by LLMs

## License

Part of the HelixAgent project.
