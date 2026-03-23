# SkillRegistry API Reference

## Overview

`digital.vasic.skillregistry` provides CLI agent skill registration and management. It handles skill definitions, capabilities, configuration templates, and multi-agent coordination for HelixAgent's 48 CLI agents.

## Core Types

### SkillDefinition

```go
type SkillDefinition struct {
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Version      string            `json:"version"`
    Agent        string            `json:"agent"`
    Capabilities []string          `json:"capabilities"`
    Config       map[string]any    `json:"config"`
    Enabled      bool              `json:"enabled"`
}
```

### Registry

```go
type Registry struct {
    // Thread-safe skill registration and lookup
}

func NewRegistry() *Registry
func (r *Registry) Register(skill SkillDefinition) error
func (r *Registry) Get(name string) (*SkillDefinition, error)
func (r *Registry) List() []SkillDefinition
func (r *Registry) Deregister(name string) error
```

## Usage

```go
registry := skillregistry.NewRegistry()

// Register a skill
err := registry.Register(SkillDefinition{
    Name:         "code-review",
    Description:  "Reviews code for quality and security",
    Agent:        "opencode",
    Capabilities: []string{"review", "suggest", "format"},
    Enabled:      true,
})

// Lookup
skill, err := registry.Get("code-review")
```

## SQL Definitions

See `docs/sql-definitions.md` for the database schema backing persistent skill storage.
