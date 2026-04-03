package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSkillManager(t *testing.T) {
	manager := NewSkillManager(nil)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.storage)
	assert.NotNil(t, manager.executor)
	assert.NotNil(t, manager.validator)
	assert.NotNil(t, manager.loader)
	assert.NotNil(t, manager.skills)
	assert.NotNil(t, manager.metrics)
}

func TestNewSkillManager_WithStorage(t *testing.T) {
	storage := NewInMemoryStorage()
	manager := NewSkillManager(storage)

	assert.Equal(t, storage, manager.storage)
}

func TestSkillManager_Register(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill for registration that has a long enough description",
		Category:    SkillCategoryCode,
	}

	err := manager.Register(skill)

	require.NoError(t, err)
	assert.False(t, skill.Enabled)
	assert.Equal(t, SkillStatusInactive, skill.Status)
	assert.NotZero(t, skill.CreatedAt)
	assert.NotZero(t, skill.UpdatedAt)
}

func TestSkillManager_Register_NilSkill(t *testing.T) {
	manager := NewSkillManager(nil)

	err := manager.Register(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillInvalid)
}

func TestSkillManager_Register_InvalidSkill(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:   "invalid",
		Name: "",
	}

	err := manager.Register(skill)

	assert.Error(t, err)
}

func TestSkillManager_Register_Duplicate(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "duplicate-skill",
		Name:        "Duplicate Skill",
		Description: "A test skill for duplicate registration that has a long enough description",
	}

	err := manager.Register(skill)
	require.NoError(t, err)

	err = manager.Register(skill)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillAlreadyExists)
}

func TestSkillManager_Register_WithDependency(t *testing.T) {
	manager := NewSkillManager(nil)

	depSkill := &Skill{
		ID:          "dep-skill",
		Name:        "Dependency Skill",
		Description: "A dependency skill that has a long enough description",
	}
	err := manager.Register(depSkill)
	require.NoError(t, err)

	skill := &Skill{
		ID:          "main-skill",
		Name:        "Main Skill",
		Description: "A main skill that has a long enough description",
		Definition: &SkillDefinition{
			Dependencies: []string{"dep-skill"},
		},
	}

	err = manager.Register(skill)
	require.NoError(t, err)
}

func TestSkillManager_Register_MissingDependency(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "main-skill",
		Name:        "Main Skill",
		Description: "A main skill that has a long enough description",
		Definition: &SkillDefinition{
			Dependencies: []string{"nonexistent"},
		},
	}

	err := manager.Register(skill)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDependencyNotFound)
}

func TestSkillManager_Unregister(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "unregister-skill",
		Name:        "Unregister Skill",
		Description: "A skill to unregister that has a long enough description",
	}

	err := manager.Register(skill)
	require.NoError(t, err)

	err = manager.Unregister(skill.ID)
	require.NoError(t, err)

	_, err = manager.Get(skill.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_Unregister_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	err := manager.Unregister("nonexistent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_Unregister_WithDependent(t *testing.T) {
	manager := NewSkillManager(nil)

	depSkill := &Skill{
		ID:          "dep-skill",
		Name:        "Dependency Skill",
		Description: "A dependency skill that has a long enough description",
	}
	err := manager.Register(depSkill)
	require.NoError(t, err)

	skill := &Skill{
		ID:          "main-skill",
		Name:        "Main Skill",
		Description: "A main skill that has a long enough description",
		Definition: &SkillDefinition{
			Dependencies: []string{"dep-skill"},
		},
	}
	err = manager.Register(skill)
	require.NoError(t, err)

	err = manager.Unregister("dep-skill")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "depends on it")
}

func TestSkillManager_Get(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "get-skill",
		Name:        "Get Skill",
		Description: "A skill to get that has a long enough description",
	}

	err := manager.Register(skill)
	require.NoError(t, err)

	retrieved, err := manager.Get(skill.ID)

	require.NoError(t, err)
	assert.Equal(t, skill.ID, retrieved.ID)
	assert.Equal(t, skill.Name, retrieved.Name)
}

func TestSkillManager_Get_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	_, err := manager.Get("nonexistent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_List(t *testing.T) {
	manager := NewSkillManager(nil)

	skill1 := &Skill{
		ID:          "skill-a",
		Name:        "Skill A",
		Description: "First skill that has a long enough description",
	}
	skill2 := &Skill{
		ID:          "skill-b",
		Name:        "Skill B",
		Description: "Second skill that has a long enough description",
	}

	manager.Register(skill1)
	manager.Register(skill2)

	skills := manager.List()

	assert.Len(t, skills, 2)
	assert.Equal(t, "Skill A", skills[0].Name)
	assert.Equal(t, "Skill B", skills[1].Name)
}

func TestSkillManager_ListByCategory(t *testing.T) {
	manager := NewSkillManager(nil)

	codeSkill := &Skill{
		ID:          "code-skill",
		Name:        "Code Skill",
		Description: "A code skill that has a long enough description",
		Category:    SkillCategoryCode,
	}
	dataSkill := &Skill{
		ID:          "data-skill",
		Name:        "Data Skill",
		Description: "A data skill that has a long enough description",
		Category:    SkillCategoryData,
	}

	manager.Register(codeSkill)
	manager.Register(dataSkill)

	codeSkills := manager.ListByCategory(SkillCategoryCode)

	assert.Len(t, codeSkills, 1)
	assert.Equal(t, "Code Skill", codeSkills[0].Name)
}

func TestSkillManager_Search(t *testing.T) {
	manager := NewSkillManager(nil)

	skill1 := &Skill{
		ID:          "search-test",
		Name:        "Search Test",
		Description: "A searchable skill that has a long enough description",
	}
	skill2 := &Skill{
		ID:          "other-skill",
		Name:        "Other Skill",
		Description: "Another skill that has a long enough description",
	}

	manager.Register(skill1)
	manager.Register(skill2)

	results := manager.Search("search")

	assert.Len(t, results, 1)
	assert.Equal(t, "Search Test", results[0].Name)
}

func TestSkillManager_Filter(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "filter-skill",
		Name:        "Filter Skill",
		Description: "A filterable skill that has a long enough description",
		Category:    SkillCategoryCode,
		Tags:        []string{"test", "filter"},
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	manager.Register(skill)
	manager.Enable(skill.ID)

	filter := &SkillFilter{
		Category: SkillCategoryCode,
		Tags:     []string{"test"},
		Enabled:  boolPtr(true),
	}

	results := manager.Filter(filter)

	assert.Len(t, results, 1)
	assert.Equal(t, "Filter Skill", results[0].Name)
}

func TestSkillManager_Enable(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "enable-skill",
		Name:        "Enable Skill",
		Description: "A skill to enable that has a long enough description",
	}

	manager.Register(skill)
	err := manager.Enable(skill.ID)

	require.NoError(t, err)

	retrieved, _ := manager.Get(skill.ID)
	assert.True(t, retrieved.Enabled)
	assert.Equal(t, SkillStatusActive, retrieved.Status)
}

func TestSkillManager_Enable_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	err := manager.Enable("nonexistent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_Disable(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "disable-skill",
		Name:        "Disable Skill",
		Description: "A skill to disable that has a long enough description",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	manager.Register(skill)
	err := manager.Disable(skill.ID)

	require.NoError(t, err)

	retrieved, _ := manager.Get(skill.ID)
	assert.False(t, retrieved.Enabled)
	assert.Equal(t, SkillStatusInactive, retrieved.Status)
}

func TestSkillManager_Disable_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	err := manager.Disable("nonexistent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_Execute(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "exec-skill",
		Name:        "Execute Skill",
		Description: "A skill to execute that has a long enough description",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	manager.Register(skill)
	manager.Enable(skill.ID)

	ctx := NewSkillExecutionContext(skill.ID)
	ctx.Inputs = map[string]interface{}{"key": "value"}

	result, err := manager.Execute(skill.ID, ctx)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillManager_Execute_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	ctx := NewSkillExecutionContext("nonexistent")

	_, err := manager.Execute("nonexistent", ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_ExecuteWithTimeout(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "timeout-skill",
		Name:        "Timeout Skill",
		Description: "A skill with timeout that has a long enough description",
		Enabled:     true,
		Status:      SkillStatusActive,
	}

	manager.Register(skill)
	manager.Enable(skill.ID)

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := manager.ExecuteWithTimeout(skill.ID, ctx, 1*time.Second)

	require.NoError(t, err)
	assert.Equal(t, ExecutionStatusSuccess, result.Status)
}

func TestSkillManager_Count(t *testing.T) {
	manager := NewSkillManager(nil)

	assert.Equal(t, 0, manager.Count())

	manager.Register(&Skill{
		ID:          "count-1",
		Name:        "Count 1",
		Description: "A counting skill that has a long enough description",
	})
	manager.Register(&Skill{
		ID:          "count-2",
		Name:        "Count 2",
		Description: "Another counting skill that has a long enough description",
	})

	assert.Equal(t, 2, manager.Count())
}

func TestSkillManager_CountActive(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "active-1",
		Name:        "Active 1",
		Description: "An active skill that has a long enough description",
	})
	manager.Enable("active-1")

	manager.Register(&Skill{
		ID:          "inactive-1",
		Name:        "Inactive 1",
		Description: "An inactive skill that has a long enough description",
	})
	// Leave inactive-1 disabled

	assert.Equal(t, 1, manager.CountActive())
}

func TestSkillManager_GetCategories(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "cat-code",
		Name:        "Code Skill",
		Description: "A code skill that has a long enough description",
		Category:    SkillCategoryCode,
	})
	manager.Register(&Skill{
		ID:          "cat-data",
		Name:        "Data Skill",
		Description: "A data skill that has a long enough description",
		Category:    SkillCategoryData,
	})
	manager.Register(&Skill{
		ID:          "cat-code-2",
		Name:        "Code Skill 2",
		Description: "Another code skill that has a long enough description",
		Category:    SkillCategoryCode,
	})

	categories := manager.GetCategories()

	assert.Len(t, categories, 2)
	assert.Contains(t, categories, SkillCategoryCode)
	assert.Contains(t, categories, SkillCategoryData)
}

func TestSkillManager_GetTags(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "tag-skill-1",
		Name:        "Tag Skill 1",
		Description: "A tagged skill that has a long enough description",
		Tags:        []string{"tag1", "tag2"},
	})
	manager.Register(&Skill{
		ID:          "tag-skill-2",
		Name:        "Tag Skill 2",
		Description: "Another tagged skill that has a long enough description",
		Tags:        []string{"tag2", "tag3"},
	})

	tags := manager.GetTags()

	assert.Len(t, tags, 3)
	assert.Contains(t, tags, "tag1")
	assert.Contains(t, tags, "tag2")
	assert.Contains(t, tags, "tag3")
}

func TestSkillManager_RegisterHandler(t *testing.T) {
	manager := NewSkillManager(nil)

	handlerCalled := false
	handler := func(s *Skill, ctx *SkillExecutionContext) (*SkillResult, error) {
		handlerCalled = true
		return NewSkillResult(ctx.ExecutionID, s.ID).Success("custom"), nil
	}

	manager.RegisterHandler("custom", handler)

	skill := &Skill{
		ID:          "custom-handler-skill",
		Name:        "Custom Handler Skill",
		Description: "A skill with custom handler that has a long enough description",
		Enabled:     true,
		Status:      SkillStatusActive,
		Definition: &SkillDefinition{
			Handler: "custom",
		},
	}

	manager.Register(skill)
	manager.Enable(skill.ID)

	ctx := NewSkillExecutionContext(skill.ID)
	result, err := manager.Execute(skill.ID, ctx)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "custom", result.Output)
}

func TestSkillManager_SetStorage(t *testing.T) {
	manager := NewSkillManager(nil)
	newStorage := NewInMemoryStorage()

	manager.SetStorage(newStorage)

	assert.Equal(t, newStorage, manager.GetStorage())
}

func TestSkillManager_InitializeFromStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	skill := &Skill{
		ID:          "stored-skill",
		Name:        "Stored Skill",
		Description: "A stored skill that has a long enough description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	storage.Save(context.Background(), skill)

	manager := NewSkillManager(storage)
	err := manager.InitializeFromStorage()

	require.NoError(t, err)
	assert.Equal(t, 1, manager.Count())

	retrieved, err := manager.Get("stored-skill")
	require.NoError(t, err)
	assert.Equal(t, "Stored Skill", retrieved.Name)
}

func TestSkillManager_UpdateSkill(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "update-skill",
		Name:        "Original Name",
		Description: "Original description that is long enough",
	}

	err := manager.Register(skill)
	require.NoError(t, err)

	updatedSkill := &Skill{
		ID:          "update-skill",
		Name:        "Updated Name",
		Description: "Updated description that is long enough",
	}

	err = manager.UpdateSkill(updatedSkill)
	require.NoError(t, err)

	retrieved, err := manager.Get("update-skill")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
}

func TestSkillManager_UpdateSkill_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "nonexistent",
		Name:        "Name",
		Description: "Description that is long enough",
	}

	err := manager.UpdateSkill(skill)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func boolPtr(b bool) *bool {
	return &b
}
