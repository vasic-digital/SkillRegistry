package agents

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkillExecutor(t *testing.T) {
	executor := NewSkillExecutor()

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.preExecutionHooks)
	assert.NotNil(t, executor.postExecutionHooks)
	assert.NotNil(t, executor.handlers)
	assert.NotNil(t, executor.semaphore)
	assert.Equal(t, 10, executor.maxConcurrent)
}

func TestNewSkillExecutorWithConcurrency(t *testing.T) {
	executor := NewSkillExecutorWithConcurrency(5)
	assert.Equal(t, 5, executor.maxConcurrent)

	executorZero := NewSkillExecutorWithConcurrency(0)
	assert.Equal(t, 10, executorZero.maxConcurrent)
}

func TestSkillExecutor_Execute(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Timeout: 30 * time.Second,
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	ctx.Inputs = map[string]interface{}{"key": "value"}

	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
	assert.Equal(t, skill.ID, result.SkillID)
	assert.NotNil(t, result.Output)
	assert.NotZero(t, result.Duration)
}

func TestSkillExecutor_Execute_NilSkill(t *testing.T) {
	executor := NewSkillExecutor()
	ctx := NewSkillExecutionContext("test")

	_, err := executor.Execute(nil, ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillInvalid)
}

func TestSkillExecutor_Execute_DisabledSkill(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "disabled-skill",
		Name:        "Disabled Skill",
		Description: "A disabled skill",
		Enabled:     false,
		Status:      SkillStatusInactive,
	}

	ctx := NewSkillExecutionContext(skill.ID)

	_, err := executor.Execute(skill, ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillDisabled)
}

func TestSkillExecutor_Execute_InactiveSkill(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "inactive-skill",
		Name:        "Inactive Skill",
		Description: "An inactive skill",
		Enabled:     true,
		Status:      SkillStatusInactive,
	}

	ctx := NewSkillExecutionContext(skill.ID)

	_, err := executor.Execute(skill, ctx)

	assert.Error(t, err)
}

func TestSkillExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "timeout-skill",
		Name:        "Timeout Skill",
		Description: "A skill that tests timeout",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)

	result, err := executor.ExecuteWithTimeout(skill, ctx, 1*time.Second)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_ExecuteWithTimeout_ActualTimeout(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "slow-skill",
		Name:        "Slow Skill",
		Description: "A slow skill",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "slow",
		},
	}

	executor.RegisterHandler("slow", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		time.Sleep(100 * time.Millisecond)
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("done"), nil
	})

	ctx := NewSkillExecutionContext(skill.ID)

	result, err := executor.ExecuteWithTimeout(skill, ctx, 50*time.Millisecond)

	assert.ErrorIs(t, err, ErrSkillTimeout)
	assert.Equal(t, ExecutionStatusTimeout, result.Status)
}

func TestSkillExecutor_RegisterHandler(t *testing.T) {
	executor := NewSkillExecutor()

	handlerCalled := false
	handler := func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		handlerCalled = true
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("custom"), nil
	}

	executor.RegisterHandler("custom", handler)

	skill := &Skill{
		ID:          "custom-skill",
		Name:        "Custom Skill",
		Description: "A custom skill",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "custom",
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "custom", result.Output)
}

func TestSkillExecutor_UnregisterHandler(t *testing.T) {
	executor := NewSkillExecutor()

	handler := func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("test"), nil
	}

	executor.RegisterHandler("temp", handler)
	executor.UnregisterHandler("temp")

	_, exists := executor.handlers["temp"]
	assert.False(t, exists)
}

func TestSkillExecutor_AddPreExecutionHook(t *testing.T) {
	executor := NewSkillExecutor()

	hookCalled := false
	hook := func(s *Skill, ctx *SkillExecutionContext) error {
		hookCalled = true
		return nil
	}

	executor.AddPreExecutionHook(hook)

	skill := &Skill{
		ID:          "hook-skill",
		Name:        "Hook Skill",
		Description: "A skill with hooks",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	_, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.True(t, hookCalled)
}

func TestSkillExecutor_AddPostExecutionHook(t *testing.T) {
	executor := NewSkillExecutor()

	hookCalled := false
	hook := func(s *Skill, ctx *SkillExecutionContext) error {
		hookCalled = true
		return nil
	}

	executor.AddPostExecutionHook(hook)

	skill := &Skill{
		ID:          "hook-skill",
		Name:        "Hook Skill",
		Description: "A skill with hooks",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	_, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.True(t, hookCalled)
}

func TestSkillExecutor_PreExecutionHook_Error(t *testing.T) {
	executor := NewSkillExecutor()

	hookError := errors.New("hook failed")
	hook := func(s *Skill, ctx *SkillExecutionContext) error {
		return hookError
	}

	executor.AddPreExecutionHook(hook)

	skill := &Skill{
		ID:          "hook-error-skill",
		Name:        "Hook Error Skill",
		Description: "A skill with failing hook",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusFailed, result.Status)
	assert.Contains(t, result.Error, "pre-execution hook failed")
}

func TestSkillExecutor_ClearPreExecutionHooks(t *testing.T) {
	executor := NewSkillExecutor()

	executor.AddPreExecutionHook(func(s *Skill, ctx *SkillExecutionContext) error {
		return nil
	})

	executor.ClearPreExecutionHooks()

	assert.Empty(t, executor.preExecutionHooks)
}

func TestSkillExecutor_ClearPostExecutionHooks(t *testing.T) {
	executor := NewSkillExecutor()

	executor.AddPostExecutionHook(func(s *Skill, ctx *SkillExecutionContext) error {
		return nil
	})

	executor.ClearPostExecutionHooks()

	assert.Empty(t, executor.postExecutionHooks)
}

func TestSkillExecutor_ValidateInputs(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		Definition: &SkillDefinition{
			Parameters: []SkillParameter{
				{Name: "required_param", Type: "string", Required: true},
				{Name: "optional_param", Type: "string", Required: false, Default: "default"},
			},
		},
	}

	tests := []struct {
		name    string
		inputs  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "all required provided",
			inputs:  map[string]interface{}{"required_param": "value"},
			wantErr: false,
		},
		{
			name:    "missing required",
			inputs:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "extra params allowed",
			inputs:  map[string]interface{}{"required_param": "value", "extra": "extra"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateInputs(skill, tt.inputs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkillExecutor_SetMaxConcurrency(t *testing.T) {
	executor := NewSkillExecutor()

	executor.SetMaxConcurrency(20)
	assert.Equal(t, 20, executor.maxConcurrent)

	executor.SetMaxConcurrency(0)
	assert.Equal(t, 10, executor.maxConcurrent)

	executor.SetMaxConcurrency(-5)
	assert.Equal(t, 10, executor.maxConcurrent)
}

func TestSkillExecutor_GetExecutionStats(t *testing.T) {
	executor := NewSkillExecutor()

	executor.RegisterHandler("test", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return nil, nil
	})
	executor.AddPreExecutionHook(func(s *Skill, ctx *SkillExecutionContext) error {
		return nil
	})
	executor.AddPostExecutionHook(func(s *Skill, ctx *SkillExecutionContext) error {
		return nil
	})

	stats := executor.GetExecutionStats()

	assert.Equal(t, 10, stats["max_concurrent"])
	assert.Equal(t, 1, stats["pre_execution_hooks"])
	assert.Equal(t, 1, stats["post_execution_hooks"])
	assert.Equal(t, 1, stats["registered_handlers"])
}

func TestCreateLoggingHook(t *testing.T) {
	logCalled := false
	var loggedMsg string

	logger := func(msg string) {
		logCalled = true
		loggedMsg = msg
	}

	hook := CreateLoggingHook(logger)
	skill := &Skill{Name: "Test Skill"}
	ctx := NewSkillExecutionContext("test")

	err := hook(skill, ctx)

	require.NoError(t, err)
	assert.True(t, logCalled)
	assert.Contains(t, loggedMsg, "Test Skill")
}

func TestCreateValidationHook(t *testing.T) {
	hook := CreateValidationHook()
	skill := &Skill{
		Name:       "Test",
		Definition: &SkillDefinition{},
	}
	ctx := NewSkillExecutionContext("test")

	err := hook(skill, ctx)
	assert.NoError(t, err)
}

func TestMergeMetadata(t *testing.T) {
	m1 := map[string]interface{}{"key1": "value1", "key2": "value2"}
	m2 := map[string]interface{}{"key2": "overridden", "key3": "value3"}

	result := mergeMetadata(m1, m2)

	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "overridden", result["key2"])
	assert.Equal(t, "value3", result["key3"])
}
