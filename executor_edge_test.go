package agents

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillExecutor_Execute_HandlerReturnsNilResult(t *testing.T) {
	executor := NewSkillExecutor()

	// Register a handler that returns nil result
	executor.RegisterHandler("nil-result", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return nil, nil
	})

	skill := &Skill{
		ID:          "nil-result-skill",
		Name:        "Nil Result",
		Description: "A skill with nil result handler",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "nil-result",
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_Execute_HandlerReturnsError(t *testing.T) {
	executor := NewSkillExecutor()

	// Register a handler that returns an error
	executor.RegisterHandler("error", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return nil, errors.New("handler error")
	})

	skill := &Skill{
		ID:          "error-skill",
		Name:        "Error Skill",
		Description: "A skill with error handler",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "error",
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.Execute(skill, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusFailed, result.Status)
	assert.Contains(t, result.Error, "handler error")
}

func TestSkillExecutor_Execute_Concurrent(t *testing.T) {
	executor := NewSkillExecutorWithConcurrency(2)

	skill := &Skill{
		ID:          "concurrent-test",
		Name:        "Concurrent Test",
		Description: "A skill for concurrent testing",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	// Execute multiple skills concurrently
	done := make(chan *SkillResult, 5)
	for i := 0; i < 5; i++ {
		go func() {
			ctx := NewSkillExecutionContext(skill.ID)
			result, _ := executor.Execute(skill, ctx)
			done <- result
		}()
	}

	for i := 0; i < 5; i++ {
		result := <-done
		assert.Equal(t, ExecutionStatusSuccess, result.Status)
	}
}

func TestSkillExecutor_ExecuteWithTimeout_Success(t *testing.T) {
	executor := NewSkillExecutor()

	// Register a fast handler
	executor.RegisterHandler("fast", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("fast result"), nil
	})

	skill := &Skill{
		ID:          "fast-skill",
		Name:        "Fast Skill",
		Description: "A fast skill",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "fast",
			Timeout: 100 * time.Millisecond,
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.ExecuteWithTimeout(skill, ctx, 200*time.Millisecond)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
	assert.Equal(t, "fast result", result.Output)
}

func TestSkillExecutor_ExecuteWithTimeout_ChannelResult(t *testing.T) {
	executor := NewSkillExecutor()

	skill := &Skill{
		ID:          "timeout-test",
		Name:        "Timeout Test",
		Description: "A skill for timeout testing",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.ExecuteWithTimeout(skill, ctx, 1*time.Second)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillExecutor_ExecuteWithTimeout_ChannelError(t *testing.T) {
	executor := NewSkillExecutor()

	// Register a handler that returns an error
	executor.RegisterHandler("timeout-error", func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		return nil, errors.New("timeout error")
	})

	skill := &Skill{
		ID:          "timeout-error-skill",
		Name:        "Timeout Error",
		Description: "A skill with timeout error",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "timeout-error",
		},
	}

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := executor.ExecuteWithTimeout(skill, ctx, 1*time.Second)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ExecutionStatusFailed, result.Status)
	assert.Contains(t, result.Error, "timeout error")
}
