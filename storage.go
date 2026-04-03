package agents

import (
	"context"
	"time"
)

// SkillStorage defines the interface for skill persistence
type SkillStorage interface {
	// Save persists a skill to storage
	Save(ctx context.Context, skill *Skill) error
	
	// Get retrieves a skill by ID (alias for Load)
	Get(ctx context.Context, id string) (*Skill, error)
	
	// Load retrieves a skill by ID
	Load(ctx context.Context, id string) (*Skill, error)
	
	// LoadByName retrieves a skill by name
	LoadByName(ctx context.Context, name string) (*Skill, error)
	
	// Delete removes a skill from storage
	Delete(ctx context.Context, id string) error
	
	// List returns all skills
	List(ctx context.Context) ([]*Skill, error)
	
	// ListByCategory returns skills filtered by category
	ListByCategory(ctx context.Context, category SkillCategory) ([]*Skill, error)
	
	// GetByCategory returns skills filtered by category (synchronous version)
	GetByCategory(category SkillCategory) []*Skill
	
	// GetByStatus returns skills filtered by status (synchronous version)
	GetByStatus(status SkillStatus) []*Skill
	
	// Search searches skills by query string
	Search(query string) []*Skill
	
	// Exists checks if a skill exists
	Exists(ctx context.Context, id string) (bool, error)
	
	// Count returns the number of skills
	Count() int
	
	// Clear removes all skills from storage
	Clear()
	
	// GetAll returns all skills (alias for List without error)
	GetAll() []string
	
	// Update updates an existing skill
	Update(ctx context.Context, skill *Skill) error
	
	// Close closes the storage connection
	Close() error
	
	// HealthCheck verifies storage connectivity
	HealthCheck(ctx context.Context) error
}

// StorageConfig contains common storage configuration
type StorageConfig struct {
	Type     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
	Options  map[string]interface{}
	
	// Timeout for storage operations
	Timeout time.Duration
	
	// MaxRetries for failed operations
	MaxRetries int
	
	// RetryDelay between retries
	RetryDelay time.Duration
}

// DefaultStorageConfig returns default storage configuration
func DefaultStorageConfig() *StorageConfig {
	return &StorageConfig{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// NewStorage creates a new storage instance based on config
func NewStorage(config *StorageConfig) (SkillStorage, error) {
	if config == nil || config.Type == "" || config.Type == "memory" {
		return NewInMemoryStorage(), nil
	}
	
	switch config.Type {
	case "postgres":
		// TODO: Implement PostgreSQL storage initialization
		return NewInMemoryStorage(), nil
	default:
		return NewInMemoryStorage(), nil
	}
}
