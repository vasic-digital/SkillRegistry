package agents

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type SkillManager struct {
	storage   SkillStorage
	executor  *SkillExecutor
	validator *SkillValidator
	loader    *Loader
	mu        sync.RWMutex
	skills    map[string]*Skill
	metrics   map[string]*SkillMetrics
}

func NewSkillManager(storage SkillStorage) *SkillManager {
	if storage == nil {
		storage = NewInMemoryStorage()
	}

	return &SkillManager{
		storage:   storage,
		executor:  NewSkillExecutor(),
		validator: NewSkillValidator(),
		loader:    NewLoader(),
		skills:    make(map[string]*Skill),
		metrics:   make(map[string]*SkillMetrics),
	}
}

func (sm *SkillManager) Register(skill *Skill) error {
	if skill == nil {
		return fmt.Errorf("%w: skill is nil", ErrSkillInvalid)
	}

	if err := sm.validator.ValidateSkill(skill); err != nil {
		return err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.skills[skill.ID]; exists {
		return fmt.Errorf("%w: %s", ErrSkillAlreadyExists, skill.ID)
	}

	if err := sm.validator.ValidateSkillDependencies(skill, sm.skills); err != nil {
		return err
	}

	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.Status == "" {
		skill.Status = SkillStatusInactive
	}

	skill.Enabled = false

	if err := sm.storage.Save(context.Background(), skill); err != nil {
		return fmt.Errorf("failed to save skill to storage: %w", err)
	}

	sm.skills[skill.ID] = skill
	sm.metrics[skill.ID] = &SkillMetrics{
		SkillID: skill.ID,
	}

	return nil
}

func (sm *SkillManager) Unregister(skillID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, exists := sm.skills[skillID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	for id, s := range sm.skills {
		if s.Definition != nil {
			for _, dep := range s.Definition.Dependencies {
				if dep == skillID {
					return fmt.Errorf("cannot unregister skill '%s': skill '%s' depends on it", skillID, id)
				}
			}
		}
	}

	if err := sm.storage.Delete(context.Background(), skillID); err != nil {
		return fmt.Errorf("failed to delete skill from storage: %w", err)
	}

	delete(sm.skills, skillID)
	delete(sm.metrics, skillID)

	return nil
}

func (sm *SkillManager) Get(skillID string) (*Skill, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	skill, exists := sm.skills[skillID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	return skill, nil
}

func (sm *SkillManager) List() []*Skill {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	skills := make([]*Skill, 0, len(sm.skills))
	for _, skill := range sm.skills {
		skills = append(skills, skill)
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	return skills
}

func (sm *SkillManager) ListByCategory(category SkillCategory) []*Skill {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Skill
	for _, skill := range sm.skills {
		if skill.Category == category {
			result = append(result, skill)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func (sm *SkillManager) Search(query string) []*Skill {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	query = strings.ToLower(query)
	var result []*Skill

	for _, skill := range sm.skills {
		if strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) ||
			skill.ID == query {
			result = append(result, skill)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func (sm *SkillManager) Filter(filter *SkillFilter) []*Skill {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Skill
	for _, skill := range sm.skills {
		if filter.Matches(skill) {
			result = append(result, skill)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func (sm *SkillManager) Enable(skillID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	skill, exists := sm.skills[skillID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	skill.Enabled = true
	skill.Status = SkillStatusActive
	skill.UpdatedAt = time.Now()

	return sm.storage.Save(context.Background(), skill)
}

func (sm *SkillManager) Disable(skillID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	skill, exists := sm.skills[skillID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	skill.Enabled = false
	skill.Status = SkillStatusInactive
	skill.UpdatedAt = time.Now()

	return sm.storage.Save(context.Background(), skill)
}

func (sm *SkillManager) Execute(skillID string, ctx *SkillExecutionContext) (*SkillResult, error) {
	skill, err := sm.Get(skillID)
	if err != nil {
		return nil, err
	}

	return sm.executor.Execute(skill, ctx)
}

func (sm *SkillManager) ExecuteWithTimeout(skillID string, ctx *SkillExecutionContext, timeout time.Duration) (*SkillResult, error) {
	skill, err := sm.Get(skillID)
	if err != nil {
		return nil, err
	}

	return sm.executor.ExecuteWithTimeout(skill, ctx, timeout)
}

func (sm *SkillManager) LoadFromDirectory(dir string) error {
	skills, err := sm.loader.LoadSkillsFromDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to load skills from directory: %w", err)
	}

	for _, skill := range skills {
		if err := sm.Register(skill); err != nil {
			if !errors.Is(err, ErrSkillAlreadyExists) {
				return fmt.Errorf("failed to register skill '%s': %w", skill.ID, err)
			}
		}
	}

	return nil
}

func (sm *SkillManager) LoadFromFile(path string) error {
	skill, err := sm.loader.LoadSkillFromFile(path)
	if err != nil {
		return fmt.Errorf("failed to load skill from file: %w", err)
	}

	return sm.Register(skill)
}

func (sm *SkillManager) GetMetrics(skillID string) (*SkillMetrics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics, exists := sm.metrics[skillID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	return metrics, nil
}

func (sm *SkillManager) GetAllMetrics() map[string]*SkillMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*SkillMetrics)
	for k, v := range sm.metrics {
		result[k] = v
	}

	return result
}

func (sm *SkillManager) UpdateSkill(skill *Skill) error {
	if skill == nil {
		return fmt.Errorf("%w: skill is nil", ErrSkillInvalid)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, exists := sm.skills[skill.ID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSkillNotFound, skill.ID)
	}

	if err := sm.validator.ValidateSkill(skill); err != nil {
		return err
	}

	skill.UpdatedAt = time.Now()
	skill.CreatedAt = existing.CreatedAt

	if err := sm.storage.Save(context.Background(), skill); err != nil {
		return fmt.Errorf("failed to update skill in storage: %w", err)
	}

	sm.skills[skill.ID] = skill

	return nil
}

func (sm *SkillManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.skills)
}

func (sm *SkillManager) CountActive() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, skill := range sm.skills {
		if skill.IsActive() {
			count++
		}
	}

	return count
}

func (sm *SkillManager) GetCategories() []SkillCategory {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	categories := make(map[SkillCategory]bool)
	for _, skill := range sm.skills {
		if skill.Category != "" {
			categories[skill.Category] = true
		}
	}

	result := make([]SkillCategory, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func (sm *SkillManager) GetTags() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tags := make(map[string]bool)
	for _, skill := range sm.skills {
		for _, tag := range skill.Tags {
			tags[tag] = true
		}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}

	sort.Strings(result)

	return result
}

func (sm *SkillManager) RegisterHandler(handlerType string, handler SkillHandler) {
	sm.executor.RegisterHandler(handlerType, handler)
}

func (sm *SkillManager) AddPreExecutionHook(hook ExecutionHook) {
	sm.executor.AddPreExecutionHook(hook)
}

func (sm *SkillManager) AddPostExecutionHook(hook ExecutionHook) {
	sm.executor.AddPostExecutionHook(hook)
}

func (sm *SkillManager) SetStorage(storage SkillStorage) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.storage = storage
}

func (sm *SkillManager) GetStorage() SkillStorage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.storage
}

func (sm *SkillManager) InitializeFromStorage() error {
	skills, err := sm.storage.List(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list skills from storage: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, skill := range skills {
		sm.skills[skill.ID] = skill
		sm.metrics[skill.ID] = &SkillMetrics{
			SkillID: skill.ID,
		}
	}

	return nil
}
