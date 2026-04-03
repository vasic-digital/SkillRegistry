package agents

import (
	"errors"
	"time"
)

var (
	ErrSkillNotFound       = errors.New("skill not found")
	ErrSkillAlreadyExists  = errors.New("skill already exists")
	ErrSkillInvalid        = errors.New("invalid skill definition")
	ErrSkillDisabled       = errors.New("skill is disabled")
	ErrSkillExecution      = errors.New("skill execution failed")
	ErrSkillTimeout        = errors.New("skill execution timed out")
	ErrDependencyNotFound  = errors.New("skill dependency not found")
	ErrCircularDependency  = errors.New("circular dependency detected")
	ErrInvalidSkillFormat  = errors.New("invalid skill format")
	ErrStorageUnavailable  = errors.New("storage unavailable")
)

type SkillStatus string

const (
	SkillStatusActive    SkillStatus = "active"
	SkillStatusInactive  SkillStatus = "inactive"
	SkillStatusDisabled  SkillStatus = "disabled"
	SkillStatusError     SkillStatus = "error"
)

type SkillCategory string

const (
	SkillCategoryCode       SkillCategory = "code"
	SkillCategoryData       SkillCategory = "data"
	SkillCategoryDevOps     SkillCategory = "devops"
	SkillCategoryTesting    SkillCategory = "testing"
	SkillCategorySecurity   SkillCategory = "security"
	SkillCategoryMonitoring SkillCategory = "monitoring"
	SkillCategoryGeneral    SkillCategory = "general"
)

type Skill struct {
	ID          string                 `json:"id" yaml:"id"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Version     string                 `json:"version" yaml:"version"`
	Category    SkillCategory          `json:"category" yaml:"category"`
	Status      SkillStatus            `json:"status" yaml:"-"`
	Triggers    []string               `json:"triggers" yaml:"triggers"`
	Tags        []string               `json:"tags" yaml:"tags"`
	Author      string                 `json:"author" yaml:"author"`
	CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
	ContentPath string                 `json:"content_path" yaml:"-"`
	Definition  *SkillDefinition       `json:"definition" yaml:"-"`
	Enabled     bool                   `json:"enabled" yaml:"-"`
}

type SkillDefinition struct {
	Parameters   []SkillParameter        `json:"parameters" yaml:"parameters"`
	Returns      SkillReturn             `json:"returns" yaml:"returns"`
	Dependencies []string                `json:"dependencies" yaml:"dependencies"`
	Permissions  []string                `json:"permissions" yaml:"permissions"`
	Timeout      time.Duration           `json:"timeout" yaml:"timeout"`
	Handler      string                  `json:"handler" yaml:"handler"`
	Examples     []SkillExample          `json:"examples" yaml:"examples"`
	Config       map[string]interface{}  `json:"config" yaml:"config"`
}

type SkillParameter struct {
	Name        string      `json:"name" yaml:"name"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description" yaml:"description"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default" yaml:"default"`
	Validation  string      `json:"validation" yaml:"validation"`
}

type SkillReturn struct {
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`
}

type SkillExample struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Input       map[string]interface{} `json:"input" yaml:"input"`
	Output      interface{}            `json:"output" yaml:"output"`
}

type SkillExecutionContext struct {
	SkillID     string                 `json:"skill_id"`
	ExecutionID string                 `json:"execution_id"`
	Inputs      map[string]interface{} `json:"inputs"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	StartedAt   time.Time              `json:"started_at"`
	Timeout     time.Duration          `json:"timeout"`
	Environment map[string]string      `json:"environment"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type SkillResult struct {
	ExecutionID string                 `json:"execution_id"`
	SkillID     string                 `json:"skill_id"`
	Status      ExecutionStatus        `json:"status"`
	Output      interface{}            `json:"output"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Duration    time.Duration          `json:"duration"`
	Logs        []string               `json:"logs"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ExecutionStatus string

const (
	ExecutionStatusPending    ExecutionStatus = "pending"
	ExecutionStatusRunning    ExecutionStatus = "running"
	ExecutionStatusSuccess    ExecutionStatus = "success"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusCancelled  ExecutionStatus = "cancelled"
	ExecutionStatusTimeout    ExecutionStatus = "timeout"
)

type SkillFilter struct {
	Category    SkillCategory `json:"category,omitempty"`
	Status      SkillStatus   `json:"status,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	SearchQuery string        `json:"search_query,omitempty"`
	Enabled     *bool         `json:"enabled,omitempty"`
}

type SkillMetrics struct {
	SkillID           string        `json:"skill_id"`
	TotalExecutions   int64         `json:"total_executions"`
	SuccessfulRuns    int64         `json:"successful_runs"`
	FailedRuns        int64         `json:"failed_runs"`
	AverageDuration   time.Duration `json:"average_duration"`
	LastExecutedAt    *time.Time    `json:"last_executed_at,omitempty"`
	LastError         string        `json:"last_error,omitempty"`
	UsageCount30Days  int64         `json:"usage_count_30_days"`
}

func NewSkillExecutionContext(skillID string) *SkillExecutionContext {
	return &SkillExecutionContext{
		SkillID:     skillID,
		ExecutionID: generateExecutionID(),
		Inputs:      make(map[string]interface{}),
		StartedAt:   time.Now(),
		Environment: make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}

func NewSkillResult(executionID, skillID string) *SkillResult {
	return &SkillResult{
		ExecutionID: executionID,
		SkillID:     skillID,
		Status:      ExecutionStatusPending,
		StartedAt:   time.Now(),
		Logs:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}
}

func generateExecutionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

func (s *Skill) IsActive() bool {
	return s.Enabled && s.Status == SkillStatusActive
}

func (s *Skill) HasTrigger(trigger string) bool {
	for _, t := range s.Triggers {
		if t == trigger {
			return true
		}
	}
	return false
}

func (s *Skill) HasTag(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (s *Skill) GetTimeout() time.Duration {
	if s.Definition != nil && s.Definition.Timeout > 0 {
		return s.Definition.Timeout
	}
	return 30 * time.Second
}

func (sr *SkillResult) Success(output interface{}) *SkillResult {
	sr.Status = ExecutionStatusSuccess
	sr.Output = output
	sr.CompletedAt = time.Now()
	sr.Duration = sr.CompletedAt.Sub(sr.StartedAt)
	return sr
}

func (sr *SkillResult) Fail(err error) *SkillResult {
	sr.Status = ExecutionStatusFailed
	sr.Error = err.Error()
	sr.CompletedAt = time.Now()
	sr.Duration = sr.CompletedAt.Sub(sr.StartedAt)
	return sr
}

func (sr *SkillResult) Cancel() *SkillResult {
	sr.Status = ExecutionStatusCancelled
	sr.CompletedAt = time.Now()
	sr.Duration = sr.CompletedAt.Sub(sr.StartedAt)
	return sr
}

func (sr *SkillResult) TimedOut() *SkillResult {
	sr.Status = ExecutionStatusTimeout
	sr.Error = "execution timed out"
	sr.CompletedAt = time.Now()
	sr.Duration = sr.CompletedAt.Sub(sr.StartedAt)
	return sr
}

func (sr *SkillResult) AddLog(message string) {
	sr.Logs = append(sr.Logs, message)
}

func (f *SkillFilter) Matches(skill *Skill) bool {
	if f.Category != "" && skill.Category != f.Category {
		return false
	}
	if f.Status != "" && skill.Status != f.Status {
		return false
	}
	if f.Enabled != nil && skill.Enabled != *f.Enabled {
		return false
	}
	if len(f.Tags) > 0 {
		hasTag := false
		for _, tag := range f.Tags {
			if skill.HasTag(tag) {
				hasTag = true
				break
			}
		}
		if !hasTag {
			return false
		}
	}
	if f.SearchQuery != "" {
		query := f.SearchQuery
		if !(contains(skill.Name, query) || contains(skill.Description, query)) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(s[:len(substr)] == substr) || contains(s[1:], substr))
}
