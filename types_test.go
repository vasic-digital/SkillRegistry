package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkillExecutionContext(t *testing.T) {
	ctx := NewSkillExecutionContext("test-skill")

	assert.Equal(t, "test-skill", ctx.SkillID)
	assert.NotEmpty(t, ctx.ExecutionID)
	assert.NotZero(t, ctx.StartedAt)
	assert.NotNil(t, ctx.Inputs)
	assert.NotNil(t, ctx.Environment)
	assert.NotNil(t, ctx.Metadata)
}

func TestNewSkillResult(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Equal(t, "skill-456", result.SkillID)
	assert.Equal(t, ExecutionStatusPending, result.Status)
	assert.NotZero(t, result.StartedAt)
	assert.NotNil(t, result.Logs)
	assert.NotNil(t, result.Metadata)
}

func TestSkillResult_Success(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")
	output := map[string]string{"key": "value"}

	successResult := result.Success(output)

	assert.Equal(t, ExecutionStatusSuccess, successResult.Status)
	assert.Equal(t, output, successResult.Output)
	assert.NotZero(t, successResult.CompletedAt)
	assert.NotZero(t, successResult.Duration)
}

func TestSkillResult_Fail(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")
	err := assert.AnError

	failResult := result.Fail(err)

	assert.Equal(t, ExecutionStatusFailed, failResult.Status)
	assert.Equal(t, err.Error(), failResult.Error)
	assert.NotZero(t, failResult.CompletedAt)
	assert.NotZero(t, failResult.Duration)
}

func TestSkillResult_Cancel(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	cancelResult := result.Cancel()

	assert.Equal(t, ExecutionStatusCancelled, cancelResult.Status)
	assert.NotZero(t, cancelResult.CompletedAt)
}

func TestSkillResult_TimedOut(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	timeoutResult := result.TimedOut()

	assert.Equal(t, ExecutionStatusTimeout, timeoutResult.Status)
	assert.Equal(t, "execution timed out", timeoutResult.Error)
}

func TestSkillResult_AddLog(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	result.AddLog("Log message 1")
	result.AddLog("Log message 2")

	assert.Len(t, result.Logs, 2)
	assert.Equal(t, "Log message 1", result.Logs[0])
	assert.Equal(t, "Log message 2", result.Logs[1])
}

func TestSkill_IsActive(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		status  SkillStatus
		want    bool
	}{
		{"active and enabled", true, SkillStatusActive, true},
		{"active but disabled", false, SkillStatusActive, false},
		{"enabled but inactive", true, SkillStatusInactive, false},
		{"disabled and error", false, SkillStatusError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := &Skill{
				Enabled: tt.enabled,
				Status:  tt.status,
			}
			assert.Equal(t, tt.want, skill.IsActive())
		})
	}
}

func TestSkill_HasTrigger(t *testing.T) {
	skill := &Skill{
		Triggers: []string{"trigger1", "trigger2", "trigger3"},
	}

	assert.True(t, skill.HasTrigger("trigger1"))
	assert.True(t, skill.HasTrigger("trigger2"))
	assert.False(t, skill.HasTrigger("trigger4"))
}

func TestSkill_HasTag(t *testing.T) {
	skill := &Skill{
		Tags: []string{"tag1", "tag2"},
	}

	assert.True(t, skill.HasTag("tag1"))
	assert.True(t, skill.HasTag("tag2"))
	assert.False(t, skill.HasTag("tag3"))
}

func TestSkill_GetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		skill    *Skill
		expected time.Duration
	}{
		{
			name:     "default timeout",
			skill:    &Skill{},
			expected: 30 * time.Second,
		},
		{
			name: "custom timeout",
			skill: &Skill{
				Definition: &SkillDefinition{
					Timeout: 60 * time.Second,
				},
			},
			expected: 60 * time.Second,
		},
		{
			name: "no definition",
			skill: &Skill{
				Definition: nil,
			},
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.skill.GetTimeout())
		})
	}
}

func TestSkillFilter_Matches(t *testing.T) {
	skill := &Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill for unit testing",
		Category:    SkillCategoryCode,
		Status:      SkillStatusActive,
		Enabled:     true,
		Tags:        []string{"test", "unit"},
	}

	tests := []struct {
		name    string
		filter  *SkillFilter
		matches bool
	}{
		{
			name:    "empty filter matches all",
			filter:  &SkillFilter{},
			matches: true,
		},
		{
			name: "matches category",
			filter: &SkillFilter{
				Category: SkillCategoryCode,
			},
			matches: true,
		},
		{
			name: "wrong category",
			filter: &SkillFilter{
				Category: SkillCategoryData,
			},
			matches: false,
		},
		{
			name: "matches status",
			filter: &SkillFilter{
				Status: SkillStatusActive,
			},
			matches: true,
		},
		{
			name: "matches enabled",
			filter: &SkillFilter{
				Enabled: boolPtr(true),
			},
			matches: true,
		},
		{
			name: "wrong enabled",
			filter: &SkillFilter{
				Enabled: boolPtr(false),
			},
			matches: false,
		},
		{
			name: "matches tag",
			filter: &SkillFilter{
				Tags: []string{"test"},
			},
			matches: true,
		},
		{
			name: "wrong tag",
			filter: &SkillFilter{
				Tags: []string{"production"},
			},
			matches: false,
		},
		{
			name: "matches search query",
			filter: &SkillFilter{
				SearchQuery: "unit",
			},
			matches: true,
		},
		{
			name: "wrong search query",
			filter: &SkillFilter{
				SearchQuery: "production",
			},
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.matches, tt.filter.Matches(skill))
		})
	}
}

func TestErrors(t *testing.T) {
	assert.NotNil(t, ErrSkillNotFound)
	assert.NotNil(t, ErrSkillAlreadyExists)
	assert.NotNil(t, ErrSkillInvalid)
	assert.NotNil(t, ErrSkillDisabled)
	assert.NotNil(t, ErrSkillExecution)
	assert.NotNil(t, ErrSkillTimeout)
	assert.NotNil(t, ErrDependencyNotFound)
	assert.NotNil(t, ErrCircularDependency)
}

func TestExecutionStatus_Constants(t *testing.T) {
	assert.Equal(t, ExecutionStatus("pending"), ExecutionStatusPending)
	assert.Equal(t, ExecutionStatus("running"), ExecutionStatusRunning)
	assert.Equal(t, ExecutionStatus("success"), ExecutionStatusSuccess)
	assert.Equal(t, ExecutionStatus("failed"), ExecutionStatusFailed)
	assert.Equal(t, ExecutionStatus("cancelled"), ExecutionStatusCancelled)
	assert.Equal(t, ExecutionStatus("timeout"), ExecutionStatusTimeout)
}

func TestSkillStatus_Constants(t *testing.T) {
	assert.Equal(t, SkillStatus("active"), SkillStatusActive)
	assert.Equal(t, SkillStatus("inactive"), SkillStatusInactive)
	assert.Equal(t, SkillStatus("disabled"), SkillStatusDisabled)
	assert.Equal(t, SkillStatus("error"), SkillStatusError)
}

func TestSkillCategory_Constants(t *testing.T) {
	assert.Equal(t, SkillCategory("code"), SkillCategoryCode)
	assert.Equal(t, SkillCategory("data"), SkillCategoryData)
	assert.Equal(t, SkillCategory("devops"), SkillCategoryDevOps)
	assert.Equal(t, SkillCategory("testing"), SkillCategoryTesting)
	assert.Equal(t, SkillCategory("security"), SkillCategorySecurity)
	assert.Equal(t, SkillCategory("monitoring"), SkillCategoryMonitoring)
	assert.Equal(t, SkillCategory("general"), SkillCategoryGeneral)
}

func TestSkillMetrics(t *testing.T) {
	now := time.Now()
	metrics := &SkillMetrics{
		SkillID:          "test-skill",
		TotalExecutions:  100,
		SuccessfulRuns:   95,
		FailedRuns:       5,
		AverageDuration:  500 * time.Millisecond,
		LastExecutedAt:   &now,
		LastError:        "",
		UsageCount30Days: 50,
	}

	assert.Equal(t, "test-skill", metrics.SkillID)
	assert.Equal(t, int64(100), metrics.TotalExecutions)
	assert.Equal(t, int64(95), metrics.SuccessfulRuns)
	assert.Equal(t, int64(5), metrics.FailedRuns)
	assert.Equal(t, 500*time.Millisecond, metrics.AverageDuration)
	assert.Equal(t, &now, metrics.LastExecutedAt)
	assert.Equal(t, int64(50), metrics.UsageCount30Days)
}

func TestSkillDefinition(t *testing.T) {
	def := &SkillDefinition{
		Parameters: []SkillParameter{
			{
				Name:        "param1",
				Type:        "string",
				Description: "A parameter",
				Required:    true,
				Default:     "default_value",
			},
		},
		Returns: SkillReturn{
			Type:        "object",
			Description: "Return value",
		},
		Dependencies: []string{"dep1", "dep2"},
		Permissions:  []string{"read", "write"},
		Timeout:      30 * time.Second,
		Handler:      "default",
		Examples: []SkillExample{
			{
				Name:        "example1",
				Description: "An example",
				Input:       map[string]interface{}{"key": "value"},
				Output:      "result",
			},
		},
		Config: map[string]interface{}{"key": "value"},
	}

	require.Len(t, def.Parameters, 1)
	assert.Equal(t, "param1", def.Parameters[0].Name)
	assert.Equal(t, "object", def.Returns.Type)
	assert.Equal(t, []string{"dep1", "dep2"}, def.Dependencies)
	assert.Equal(t, 30*time.Second, def.Timeout)
}


