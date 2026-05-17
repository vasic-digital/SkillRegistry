package agents

import "time"

// installDefaultEchoHandler installs a unit-test-only stub handler under the
// "default" key so tests exercising Execute (without registering a specific
// handler key on the skill definition) get deterministic behaviour.
//
// CONST-050(A) permits mocks/stubs in *_test.go (this file is *_test.go).
// Production code MUST call SkillExecutor.RegisterHandler(handlerType, fn)
// with a real handler; without it, Execute returns ErrNoHandlerRegistered
// (round-26 §11.4 audit fix).
func installDefaultEchoHandler(executor *SkillExecutor) {
	executor.RegisterHandler("default", func(skill *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		result := NewSkillResult(ctx.ExecutionID, skill.ID)
		result.AddLog("test stub: executing skill: " + skill.Name)
		result.AddLog("test stub: description: " + skill.Description)
		if len(ctx.Inputs) > 0 {
			result.Output = map[string]interface{}{
				"skill_id":    skill.ID,
				"skill_name":  skill.Name,
				"executed_at": time.Now(),
				"inputs":      ctx.Inputs,
			}
		} else {
			result.Output = map[string]interface{}{
				"skill_id":    skill.ID,
				"skill_name":  skill.Name,
				"executed_at": time.Now(),
			}
		}
		return result, nil
	})
}
