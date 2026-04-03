package agents

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillExecutor_Execute_PostHookError(t *testing.T) {
	executor := NewSkillExecutor()

	// Add a post-execution hook that returns an error
	hookCalled := false
	executor.AddPostExecutionHook(func(s *Skill, ctx *SkillExecutionContext) error {
		hookCalled = true
		return errors.New("post hook error")
	})

	skill := &Skill{
		ID:          "post-hook-skill",
		Name:        "Post Hook Skill",
		Description: "A skill with failing post hook",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.True(t, hookCalled)
	// Execution should still succeed even if post hook fails
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_getHandler_WithDefinition(t *testing.T) {
	executor := NewSkillExecutor()

	handlerCalled := false
	executor.RegisterHandler("custom", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		handlerCalled = true
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("custom"), nil
	})

	skill := &Skill{
		ID:          "handler-test",
		Name:        "Handler Test",
		Description: "A skill for handler testing",
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

func TestSkillExecutor_getHandler_Unregistered(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "unregistered-handler",
		Name:        "Unregistered Handler",
		Description: "A skill with unregistered handler",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "nonexistent",
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	// Should use default handler
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_ValidateInputs_WithDefault(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		Definition: &SkillDefinition{
			Parameters: []SkillParameter{
				{Name: "required_param", Type: "string", Required: true},
				{Name: "optional_with_default", Type: "string", Required: true, Default: "default_value"},
			},
		},
	}

	// Should pass because optional_with_default has a default
	inputs := map[string]interface{}{
		"required_param": "value",
	}

	err := executor.ValidateInputs(skill, inputs)
	assert.NoError(t, err)
}

func TestSkillExecutor_ExecuteWithTimeout_ZeroTimeout(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "zero-timeout",
		Name:        "Zero Timeout",
		Description: "A skill with zero timeout",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Timeout: 100 * time.Millisecond,
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.ExecuteWithTimeout(skill, ctx, 0)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_Execute_DefaultHandlerLogs(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "log-test",
		Name:        "Log Test",
		Description: "A skill for testing logs",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	ctx.Inputs = map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
	// Check that logs were added
	assert.GreaterOrEqual(t, len(result.Logs), 2)
}

func TestCreateLoggingHook_NilLogger(t *testing.T) {
	hook := CreateLoggingHook(nil)
	skill := &Skill{Name: "Test"}
	ctx := NewSkillExecutionContext("test")

	// Should not panic with nil logger
	err := hook(skill, ctx)
	assert.NoError(t, err)
}

func TestCreateValidationHook_WithNilDefinition(t *testing.T) {
	hook := CreateValidationHook()
	skill := &Skill{
		Name:       "Test",
		Definition: nil,
	}
	ctx := NewSkillExecutionContext("test")

	// Should handle nil definition gracefully
	err := hook(skill, ctx)
	assert.NoError(t, err)
}

func TestSkillResult_Success_MultipleCalls(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	output1 := map[string]string{"status": "first"}
	output2 := map[string]string{"status": "second"}

	result.Success(output1)
	result.Success(output2)

	// Last call should win
	assert.Equal(t, output2, result.Output)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_ConcurrentExecution(t *testing.T) {
	executor := NewSkillExecutorWithConcurrency(5)

	skill := &Skill{
		ID:          "concurrent",
		Name:        "Concurrent",
		Description: "A skill for concurrent testing",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			ctx := NewSkillExecutionContext(skill.ID)
			_, err := executor.Execute(skill, ctx)
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSkillExecutor_ExecuteWithTimeout_ContextTimeout(t *testing.T) {
	executor := NewSkillExecutor()

	// Register a slow handler
	executor.RegisterHandler("slow", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		time.Sleep(100 * time.Millisecond)
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("done"), nil
	})

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

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.ExecuteWithTimeout(skill, ctx, 50*time.Millisecond)

	assert.ErrorIs(t, err, ErrSkillTimeout)
	assert.Equal(t, ExecutionStatusTimeout, result.Status)
}
