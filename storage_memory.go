package agents

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// generateSkillID generates a unique skill ID
func generateSkillID() string {
	return fmt.Sprintf("skill_%d", time.Now().UnixNano())
}

// MemoryStorage implements SkillStorage using in-memory storage
type MemoryStorage struct {
	skills   map[string]*Skill
	byName   map[string]string // name -> id mapping
	mu       sync.RWMutex
	config   *StorageConfig
}

// NewMemoryStorage creates a new in-memory skill storage
func NewMemoryStorage(config *StorageConfig) *MemoryStorage {
	if config == nil {
		config = DefaultStorageConfig()
	}
	
	return &MemoryStorage{
		skills: make(map[string]*Skill),
		byName: make(map[string]string),
		config: config,
	}
}

// NewInMemoryStorage creates a new in-memory skill storage (no arguments version)
func NewInMemoryStorage() *MemoryStorage {
	return NewMemoryStorage(nil)
}

// Save persists a skill to memory storage
func (s *MemoryStorage) Save(ctx context.Context, skill *Skill) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	
	if skill == nil {
		return fmt.Errorf("%w: skill is nil", ErrSkillInvalid)
	}
	
	if skill.ID == "" {
		skill.ID = generateSkillID()
	}
	
	if skill.Name == "" {
		return fmt.Errorf("%w: skill name is required", ErrSkillInvalid)
	}
	
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.skills[skill.ID] = skill
	s.byName[skill.Name] = skill.ID
	
	return nil
}

// Get retrieves a skill by ID (alias for Load)
func (s *MemoryStorage) Get(ctx context.Context, id string) (*Skill, error) {
	return s.Load(ctx, id)
}

// Load retrieves a skill by ID
func (s *MemoryStorage) Load(ctx context.Context, id string) (*Skill, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	
	if id == "" {
		return nil, fmt.Errorf("skill ID is required")
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	skill, exists := s.skills[id]
	if !exists {
		return nil, ErrSkillNotFound
	}
	
	return skill, nil
}

// LoadByName retrieves a skill by name
func (s *MemoryStorage) LoadByName(ctx context.Context, name string) (*Skill, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	
	if name == "" {
		return nil, fmt.Errorf("skill name is required")
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	id, exists := s.byName[name]
	if !exists {
		return nil, ErrSkillNotFound
	}
	
	skill, exists := s.skills[id]
	if !exists {
		return nil, ErrSkillNotFound
	}
	
	return skill, nil
}

// Delete removes a skill from storage
func (s *MemoryStorage) Delete(ctx context.Context, id string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	
	if id == "" {
		return fmt.Errorf("skill ID is required")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	skill, exists := s.skills[id]
	if !exists {
		return ErrSkillNotFound
	}
	
	delete(s.skills, id)
	delete(s.byName, skill.Name)
	
	return nil
}

// List returns all skills
func (s *MemoryStorage) List(ctx context.Context) ([]*Skill, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	skills := make([]*Skill, 0, len(s.skills))
	for _, skill := range s.skills {
		skills = append(skills, skill)
	}
	
	return skills, nil
}

// ListByCategory returns skills filtered by category
func (s *MemoryStorage) ListByCategory(ctx context.Context, category SkillCategory) ([]*Skill, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var skills []*Skill
	for _, skill := range s.skills {
		if skill.Category == category {
			skills = append(skills, skill)
		}
	}
	
	return skills, nil
}

// GetByCategory returns skills filtered by category (synchronous version)
func (s *MemoryStorage) GetByCategory(category SkillCategory) []*Skill {
	skills, _ := s.ListByCategory(context.Background(), category)
	return skills
}

// GetByStatus returns skills filtered by status (synchronous version)
func (s *MemoryStorage) GetByStatus(status SkillStatus) []*Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var skills []*Skill
	for _, skill := range s.skills {
		if skill.Status == status {
			skills = append(skills, skill)
		}
	}
	
	return skills
}

// Search searches skills by query string
func (s *MemoryStorage) Search(query string) []*Skill {
	if query == "" {
		skills, _ := s.List(context.Background())
		return skills
	}
	
	query = strings.ToLower(query)
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var skills []*Skill
	for _, skill := range s.skills {
		if strings.Contains(strings.ToLower(skill.Name), query) ||
		   strings.Contains(strings.ToLower(skill.Description), query) ||
		   strings.Contains(strings.ToLower(string(skill.Category)), query) {
			skills = append(skills, skill)
		}
	}
	
	return skills
}

// Exists checks if a skill exists
func (s *MemoryStorage) Exists(ctx context.Context, id string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	_, exists := s.skills[id]
	return exists, nil
}

// Count returns the number of skills
func (s *MemoryStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return len(s.skills)
}

// Clear removes all skills from storage
func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.skills = make(map[string]*Skill)
	s.byName = make(map[string]string)
}

// GetAll returns all skills (alias for List without error)
// GetAll returns all skill IDs
func (s *MemoryStorage) GetAll() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ids := make([]string, 0, len(s.skills))
	for id := range s.skills {
		ids = append(ids, id)
	}
	return ids
}

// Update updates an existing skill
func (s *MemoryStorage) Update(ctx context.Context, skill *Skill) error {
	if skill == nil {
		return fmt.Errorf("skill cannot be nil")
	}
	
	if skill.ID == "" {
		return fmt.Errorf("skill ID is required for update")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	existing, exists := s.skills[skill.ID]
	if !exists {
		return ErrSkillNotFound
	}
	
	// Update name mapping if name changed
	if skill.Name != existing.Name {
		delete(s.byName, existing.Name)
		s.byName[skill.Name] = skill.ID
	}
	
	skill.UpdatedAt = time.Now()
	skill.CreatedAt = existing.CreatedAt
	s.skills[skill.ID] = skill
	
	return nil
}

// Close closes the storage connection
func (s *MemoryStorage) Close() error {
	return nil
}

// HealthCheck verifies storage connectivity
func (s *MemoryStorage) HealthCheck(ctx context.Context) error {
	return nil
}

// InMemoryStorage is an alias for MemoryStorage for backward compatibility
type InMemoryStorage = MemoryStorage
