package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ExecutionHook func(skill *Skill, ctx *SkillExecutionContext) error

type SkillExecutor struct {
	preExecutionHooks  []ExecutionHook
	postExecutionHooks []ExecutionHook
	handlers           map[string]SkillHandler
	mu                 sync.RWMutex
	maxConcurrent      int
	semaphore          chan struct{}
}

type SkillHandler func(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error)

func NewSkillExecutor() *SkillExecutor {
	maxConcurrent := 10
	return &SkillExecutor{
		preExecutionHooks:  make([]ExecutionHook, 0),
		postExecutionHooks: make([]ExecutionHook, 0),
		handlers:           make(map[string]SkillHandler),
		maxConcurrent:      maxConcurrent,
		semaphore:          make(chan struct{}, maxConcurrent),
	}
}

func NewSkillExecutorWithConcurrency(maxConcurrent int) *SkillExecutor {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &SkillExecutor{
		preExecutionHooks:  make([]ExecutionHook, 0),
		postExecutionHooks: make([]ExecutionHook, 0),
		handlers:           make(map[string]SkillHandler),
		maxConcurrent:      maxConcurrent,
		semaphore:          make(chan struct{}, maxConcurrent),
	}
}

func (se *SkillExecutor) Execute(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
	if skill == nil {
		return nil, fmt.Errorf("%w: skill is nil", ErrSkillInvalid)
	}

	if !skill.IsActive() {
		if skill.Status == SkillStatusDisabled || !skill.Enabled {
			return nil, fmt.Errorf("%w: %s", ErrSkillDisabled, skill.ID)
		}
		return nil, fmt.Errorf("skill is not active: %s", skill.ID)
	}

	se.semaphore <- struct{}{}
	defer func() { <-se.semaphore }()

	result := NewSkillResult(ctx.ExecutionID, skill.ID)
	result.Status = ExecutionStatusRunning

	for i, hook := range se.preExecutionHooks {
		if err := hook(skill, ctx); err != nil {
			result.AddLog(fmt.Sprintf("Pre-execution hook %d failed: %v", i, err))
			return result.Fail(fmt.Errorf("pre-execution hook failed: %w", err)), nil
		}
	}

	handler := se.getHandler(skill.Definition)
	if handler == nil {
		result.AddLog("No handler registered for skill, using default handler")
		handler = se.defaultHandler
	}

	execResult, err := handler(skill, ctx)
	if err != nil {
		result.Fail(err)
		result.AddLog(fmt.Sprintf("Execution failed: %v", err))
	} else if execResult != nil {
		result.Success(execResult.Output)
		result.Logs = append(result.Logs, execResult.Logs...)
		result.Metadata = mergeMetadata(result.Metadata, execResult.Metadata)
	} else {
		result.Success(nil)
		result.AddLog("Handler returned nil result")
	}

	for i, hook := range se.postExecutionHooks {
		if hookErr := hook(skill, ctx); hookErr != nil {
			result.AddLog(fmt.Sprintf("Post-execution hook %d failed: %v", i, hookErr))
		}
	}

	return result, nil
}

func (se *SkillExecutor) ExecuteWithTimeout(skill *Skill, ctx *SkillExecutionContext, timeout time.Duration) (*SkillResult, error) {
	if timeout <= 0 {
		timeout = skill.GetTimeout()
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan *SkillResult, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := se.Execute(skill, ctx)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctxWithTimeout.Done():
		result := NewSkillResult(ctx.ExecutionID, skill.ID)
		result.Status = ExecutionStatusRunning
		return result.TimedOut(), ErrSkillTimeout
	}
}

func (se *SkillExecutor) RegisterHandler(handlerType string, handler SkillHandler) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.handlers[handlerType] = handler
}

func (se *SkillExecutor) UnregisterHandler(handlerType string) {
	se.mu.Lock()
	defer se.mu.Unlock()
	delete(se.handlers, handlerType)
}

func (se *SkillExecutor) AddPreExecutionHook(hook ExecutionHook) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.preExecutionHooks = append(se.preExecutionHooks, hook)
}

func (se *SkillExecutor) AddPostExecutionHook(hook ExecutionHook) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.postExecutionHooks = append(se.postExecutionHooks, hook)
}

func (se *SkillExecutor) ClearPreExecutionHooks() {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.preExecutionHooks = make([]ExecutionHook, 0)
}

func (se *SkillExecutor) ClearPostExecutionHooks() {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.postExecutionHooks = make([]ExecutionHook, 0)
}

func (se *SkillExecutor) getHandler(def *SkillDefinition) SkillHandler {
	se.mu.RLock()
	defer se.mu.RUnlock()

	if def == nil || def.Handler == "" {
		return se.handlers["default"]
	}

	if handler, ok := se.handlers[def.Handler]; ok {
		return handler
	}

	return se.handlers["default"]
}

func (se *SkillExecutor) defaultHandler(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
	result := NewSkillResult(ctx.ExecutionID, skill.ID)
	
	result.AddLog(fmt.Sprintf("Executing skill: %s", skill.Name))
	result.AddLog(fmt.Sprintf("Skill description: %s", skill.Description))

	if len(ctx.Inputs) > 0 {
		result.AddLog(fmt.Sprintf("Received %d input parameters", len(ctx.Inputs)))
	}

	output := map[string]interface{}{
		"skill_id":    skill.ID,
		"skill_name":  skill.Name,
		"executed_at": time.Now(),
		"inputs":      ctx.Inputs,
	}

	result.Output = output
	return result, nil
}

func (se *SkillExecutor) ValidateInputs(skill *Skill, inputs map[string]interface{}) error {
	if skill.Definition == nil {
		return nil
	}

	for _, param := range skill.Definition.Parameters {
		if param.Required {
			if _, ok := inputs[param.Name]; !ok {
				if param.Default == nil {
					return fmt.Errorf("missing required parameter: %s", param.Name)
				}
			}
		}
	}

	return nil
}

func (se *SkillExecutor) SetMaxConcurrency(max int) {
	se.mu.Lock()
	defer se.mu.Unlock()
	
	if max <= 0 {
		max = 10
	}
	
	se.maxConcurrent = max
	se.semaphore = make(chan struct{}, max)
}

func (se *SkillExecutor) GetExecutionStats() map[string]interface{} {
	se.mu.RLock()
	defer se.mu.RUnlock()

	return map[string]interface{}{
		"max_concurrent":        se.maxConcurrent,
		"pre_execution_hooks":   len(se.preExecutionHooks),
		"post_execution_hooks":  len(se.postExecutionHooks),
		"registered_handlers":   len(se.handlers),
	}
}

func mergeMetadata(m1, m2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

func CreateLoggingHook(logger func(string)) ExecutionHook {
	return func(skill *Skill, ctx *SkillExecutionContext) error {
		if logger != nil {
			logger(fmt.Sprintf("Executing skill: %s (execution: %s)", skill.Name, ctx.ExecutionID))
		}
		return nil
	}
}

func CreateValidationHook() ExecutionHook {
	return func(skill *Skill, ctx *SkillExecutionContext) error {
		if skill.Definition == nil {
			return nil
		}
		return nil
	}
}
