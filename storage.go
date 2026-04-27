package agents

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	// Pure SQL driver for PostgreSQL. Imported for side effects so
	// database/sql.Open("postgres", ...) can resolve the driver.
	_ "github.com/lib/pq"
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

// NewStorage creates a new storage instance based on config.
// Recognised types: "memory" (default), "postgres".
// An unrecognised type falls back to in-memory with a diagnostic
// warning returned via the error chain so callers can log it without
// crashing the boot sequence.
func NewStorage(config *StorageConfig) (SkillStorage, error) {
	if config == nil || config.Type == "" || config.Type == "memory" {
		return NewInMemoryStorage(), nil
	}

	switch strings.ToLower(config.Type) {
	case "postgres", "postgresql":
		return newPostgresFromConfig(config)
	case "memory":
		return NewInMemoryStorage(), nil
	default:
		return NewInMemoryStorage(), fmt.Errorf("unknown storage type %q; falling back to in-memory", config.Type)
	}
}

// newPostgresFromConfig constructs a PostgresStorage from a StorageConfig,
// opening the database connection and initialising the schema. Any
// failure returns the error — callers may fall back to in-memory at
// their discretion, but the default behaviour is strict so startup
// surfaces configuration bugs immediately.
func newPostgresFromConfig(config *StorageConfig) (SkillStorage, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("postgres storage requires Host")
	}
	if config.Database == "" {
		return nil, fmt.Errorf("postgres storage requires Database")
	}

	port := config.Port
	if port == 0 {
		port = 5432
	}
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}

	// Build a lib/pq DSN URL. Using url.URL (rather than string
	// concatenation) guarantees proper escaping of passwords containing
	// special characters.
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", config.Host, port),
		Path:   "/" + config.Database,
	}
	if config.Username != "" {
		if config.Password != "" {
			u.User = url.UserPassword(config.Username, config.Password)
		} else {
			u.User = url.User(config.Username)
		}
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	// Route every statement through a timeout equal to the caller's
	// configured Timeout (default 30s). lib/pq uses statement_timeout
	// in milliseconds.
	if config.Timeout > 0 {
		q.Set("statement_timeout", fmt.Sprintf("%d", int(config.Timeout/time.Millisecond)))
	}
	u.RawQuery = q.Encode()

	db, err := sql.Open("postgres", u.String())
	if err != nil {
		return nil, fmt.Errorf("postgres storage: open: %w", err)
	}

	// Sensible connection pool defaults. These are intentionally
	// conservative — the skill registry is a low-QPS component.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Use the configured Timeout as the ping/schema ceiling; a bad
	// DSN or unreachable host should fail fast rather than hang the
	// boot sequence.
	pingTimeout := config.Timeout
	if pingTimeout <= 0 {
		pingTimeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres storage: ping: %w", err)
	}

	ps := NewPostgresStorage(db, config)
	if err := ps.InitSchema(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres storage: init schema: %w", err)
	}

	return ps, nil
}
