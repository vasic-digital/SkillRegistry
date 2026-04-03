package agents

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// PostgresStorage implements SkillStorage using PostgreSQL
type PostgresStorage struct {
	db     *sql.DB
	config *StorageConfig
}

// NewPostgresStorage creates a new PostgreSQL skill storage
func NewPostgresStorage(db *sql.DB, config *StorageConfig) *PostgresStorage {
	if config == nil {
		config = DefaultStorageConfig()
	}
	
	return &PostgresStorage{
		db:     db,
		config: config,
	}
}

// InitSchema creates the necessary database tables
func (s *PostgresStorage) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS skills (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		category VARCHAR(100) NOT NULL,
		description TEXT,
		status VARCHAR(50) DEFAULT 'active',
		version VARCHAR(50),
		triggers JSONB,
		tags JSONB,
		author VARCHAR(255),
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		metadata JSONB,
		content_path VARCHAR(500),
		definition JSONB
	);
	
	CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
	CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name);
	CREATE INDEX IF NOT EXISTS idx_skills_status ON skills(status);
	`
	
	_, err := s.db.ExecContext(ctx, query)
	return err
}

// Save persists a skill to PostgreSQL
func (s *PostgresStorage) Save(ctx context.Context, skill *Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}
	
	if skill.ID == "" {
		skill.ID = generateSkillID()
	}
	
	if skill.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}
	
	query := `
	INSERT INTO skills (id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		category = EXCLUDED.category,
		description = EXCLUDED.description,
		status = EXCLUDED.status,
		version = EXCLUDED.version,
		triggers = EXCLUDED.triggers,
		tags = EXCLUDED.tags,
		author = EXCLUDED.author,
		updated_at = EXCLUDED.updated_at,
		metadata = EXCLUDED.metadata,
		content_path = EXCLUDED.content_path,
		definition = EXCLUDED.definition
	`
	
	_, err := s.db.ExecContext(ctx, query,
		skill.ID, skill.Name, skill.Category, skill.Description,
		skill.Status, skill.Version,
		toJSON(skill.Triggers), toJSON(skill.Tags),
		skill.Author, skill.CreatedAt, skill.UpdatedAt,
		toJSON(skill.Metadata), skill.ContentPath,
		toJSON(skill.Definition),
	)
	
	return err
}

// Get retrieves a skill by ID (alias for Load)
func (s *PostgresStorage) Get(ctx context.Context, id string) (*Skill, error) {
	return s.Load(ctx, id)
}

// Load retrieves a skill by ID
func (s *PostgresStorage) Load(ctx context.Context, id string) (*Skill, error) {
	if id == "" {
		return nil, fmt.Errorf("skill ID is required")
	}
	
	query := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills WHERE id = $1
	`
	
	return s.scanSkill(s.db.QueryRowContext(ctx, query, id))
}

// LoadByName retrieves a skill by name
func (s *PostgresStorage) LoadByName(ctx context.Context, name string) (*Skill, error) {
	if name == "" {
		return nil, fmt.Errorf("skill name is required")
	}
	
	query := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills WHERE name = $1
	`
	
	return s.scanSkill(s.db.QueryRowContext(ctx, query, name))
}

// Delete removes a skill from storage
func (s *PostgresStorage) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("skill ID is required")
	}
	
	query := `DELETE FROM skills WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return ErrSkillNotFound
	}
	
	return nil
}

// List returns all skills
func (s *PostgresStorage) List(ctx context.Context) ([]*Skill, error) {
	query := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return s.scanSkills(rows)
}

// ListByCategory returns skills filtered by category
func (s *PostgresStorage) ListByCategory(ctx context.Context, category SkillCategory) ([]*Skill, error) {
	query := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills WHERE category = $1 ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return s.scanSkills(rows)
}

// GetByCategory returns skills filtered by category (synchronous version)
func (s *PostgresStorage) GetByCategory(category SkillCategory) []*Skill {
	skills, _ := s.ListByCategory(context.Background(), category)
	return skills
}

// GetByStatus returns skills filtered by status (synchronous version)
func (s *PostgresStorage) GetByStatus(status SkillStatus) []*Skill {
	query := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills WHERE status = $1 ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(context.Background(), query, status)
	if err != nil {
		return nil
	}
	defer rows.Close()
	
	skills, _ := s.scanSkills(rows)
	return skills
}

// Search searches skills by query string
func (s *PostgresStorage) Search(query string) []*Skill {
	if query == "" {
		skills, _ := s.List(context.Background())
		return skills
	}
	
	searchQuery := `%` + query + `%`
	sql := `
	SELECT id, name, category, description, status, version, triggers, tags, author, created_at, updated_at, metadata, content_path, definition
	FROM skills 
	WHERE name ILIKE $1 OR description ILIKE $1 OR category ILIKE $1
	ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(context.Background(), sql, searchQuery)
	if err != nil {
		return nil
	}
	defer rows.Close()
	
	skills, _ := s.scanSkills(rows)
	return skills
}

// Exists checks if a skill exists
func (s *PostgresStorage) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM skills WHERE id = $1)`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

// Count returns the number of skills
func (s *PostgresStorage) Count() int {
	query := `SELECT COUNT(*) FROM skills`
	var count int
	err := s.db.QueryRowContext(context.Background(), query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// Clear removes all skills from storage
func (s *PostgresStorage) Clear() {
	s.db.ExecContext(context.Background(), `DELETE FROM skills`)
}
// GetAll returns all skill IDs
func (s *PostgresStorage) GetAll() []string {
	query := `SELECT id FROM skills`
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil
	}
	defer rows.Close()
	
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// Update updates an existing skill
func (s *PostgresStorage) Update(ctx context.Context, skill *Skill) error {
	return s.Save(ctx, skill)
}

// Close closes the storage connection
func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// HealthCheck verifies storage connectivity
func (s *PostgresStorage) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// scanSkill scans a single skill from a row
func (s *PostgresStorage) scanSkill(row *sql.Row) (*Skill, error) {
	var skill Skill
	var triggers, tags, metadata, definition []byte
	
	err := row.Scan(
		&skill.ID, &skill.Name, &skill.Category, &skill.Description,
		&skill.Status, &skill.Version, &triggers, &tags,
		&skill.Author, &skill.CreatedAt, &skill.UpdatedAt,
		&metadata, &skill.ContentPath, &definition,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSkillNotFound
	}
	if err != nil {
		return nil, err
	}
	
	fromJSON(triggers, &skill.Triggers)
	fromJSON(tags, &skill.Tags)
	fromJSON(metadata, &skill.Metadata)
	fromJSON(definition, &skill.Definition)
	
	return &skill, nil
}

// scanSkills scans multiple skills from rows
func (s *PostgresStorage) scanSkills(rows *sql.Rows) ([]*Skill, error) {
	var skills []*Skill
	
	for rows.Next() {
		var skill Skill
		var triggers, tags, metadata, definition []byte
		
		err := rows.Scan(
			&skill.ID, &skill.Name, &skill.Category, &skill.Description,
			&skill.Status, &skill.Version, &triggers, &tags,
			&skill.Author, &skill.CreatedAt, &skill.UpdatedAt,
			&metadata, &skill.ContentPath, &definition,
		)
		if err != nil {
			return nil, err
		}
		
		fromJSON(triggers, &skill.Triggers)
		fromJSON(tags, &skill.Tags)
		fromJSON(metadata, &skill.Metadata)
		fromJSON(definition, &skill.Definition)
		
		skills = append(skills, &skill)
	}
	
	return skills, rows.Err()
}

// toJSON converts value to JSON bytes
func toJSON(v interface{}) []byte {
	if v == nil {
		return []byte("null")
	}
	data, _ := json.Marshal(v)
	return data
}

// fromJSON parses JSON bytes into value
func fromJSON(data []byte, v interface{}) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	return json.Unmarshal(data, v)
}
