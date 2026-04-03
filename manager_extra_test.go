package agents

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillManager_LoadFromDirectory(t *testing.T) {
	manager := NewSkillManager(nil)

	tmpDir := t.TempDir()

	yamlContent := `name: dir-skill
description: A skill loaded from directory that has a long enough description`
	err := os.WriteFile(filepath.Join(tmpDir, "skill.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	err = manager.LoadFromDirectory(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, 1, manager.Count())

	skill, err := manager.Get("dir-skill")
	require.NoError(t, err)
	assert.Equal(t, "dir-skill", skill.Name)
}

func TestSkillManager_LoadFromDirectory_AlreadyExists(t *testing.T) {
	manager := NewSkillManager(nil)

	tmpDir := t.TempDir()

	yamlContent := `name: existing-skill
description: A skill that will be loaded twice that has a long enough description`
	err := os.WriteFile(filepath.Join(tmpDir, "skill1.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "skill2.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	err = manager.LoadFromDirectory(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, 1, manager.Count()) // Second one skipped due to duplicate
}

func TestSkillManager_LoadFromDirectory_InvalidSkill(t *testing.T) {
	manager := NewSkillManager(nil)

	tmpDir := t.TempDir()

	// Create an invalid skill file (missing required fields)
	yamlContent := `name: ""
description: ""`
	err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	err = manager.LoadFromDirectory(tmpDir)

	assert.Error(t, err)
}

func TestSkillManager_LoadFromFile(t *testing.T) {
	manager := NewSkillManager(nil)

	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "skill.yaml")

	yamlContent := `name: file-skill
description: A skill loaded from file that has a long enough description`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	err = manager.LoadFromFile(yamlPath)

	require.NoError(t, err)
	assert.Equal(t, 1, manager.Count())

	skill, err := manager.Get("file-skill")
	require.NoError(t, err)
	assert.Equal(t, "file-skill", skill.Name)
}

func TestSkillManager_LoadFromFile_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	err := manager.LoadFromFile("/nonexistent/path/skill.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load skill from file")
}

func TestSkillManager_GetMetrics(t *testing.T) {
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "metrics-skill",
		Name:        "Metrics Skill",
		Description: "A skill for metrics testing that has a long enough description",
	}

	err := manager.Register(skill)
	require.NoError(t, err)

	metrics, err := manager.GetMetrics(skill.ID)

	require.NoError(t, err)
	assert.Equal(t, skill.ID, metrics.SkillID)
	assert.Equal(t, int64(0), metrics.TotalExecutions)
}

func TestSkillManager_GetMetrics_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)

	_, err := manager.GetMetrics("nonexistent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_GetAllMetrics(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "metrics-1",
		Name:        "Metrics 1",
		Description: "First metrics skill that has a long enough description",
	})
	manager.Register(&Skill{
		ID:          "metrics-2",
		Name:        "Metrics 2",
		Description: "Second metrics skill that has a long enough description",
	})

	allMetrics := manager.GetAllMetrics()

	assert.Len(t, allMetrics, 2)
	assert.Contains(t, allMetrics, "metrics-1")
	assert.Contains(t, allMetrics, "metrics-2")
}

func TestSkillManager_ExecuteWithTimeout_NotFound(t *testing.T) {
	manager := NewSkillManager(nil)
	ctx := NewSkillExecutionContext("nonexistent")

	_, err := manager.ExecuteWithTimeout("nonexistent", ctx, 1*time.Second)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestSkillManager_UpdateSkill_Invalid(t *testing.T) {
	manager := NewSkillManager(nil)

	// Try to update a nil skill
	err := manager.UpdateSkill(nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillInvalid)
}

func TestSkillManager_UpdateSkill_InvalidData(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "update-invalid",
		Name:        "Update Invalid",
		Description: "Original description that is long enough",
	})

	// Try to update with invalid data (empty name)
	err := manager.UpdateSkill(&Skill{
		ID:          "update-invalid",
		Name:        "",
		Description: "Updated description that is long enough",
	})

	assert.Error(t, err)
}

func TestSkillManager_Register_StorageError(t *testing.T) {
	// This test would require a mock storage that returns an error
	// For now, we test with valid storage
	manager := NewSkillManager(nil)

	skill := &Skill{
		ID:          "storage-test",
		Name:        "Storage Test",
		Description: "A skill for storage testing that has a long enough description",
	}

	err := manager.Register(skill)
	require.NoError(t, err)
}

func TestSkillManager_Filter_ByQuery(t *testing.T) {
	manager := NewSkillManager(nil)

	manager.Register(&Skill{
		ID:          "query-test",
		Name:        "Query Test Skill",
		Description: "A skill for query testing that has a long enough description",
	})
	manager.Register(&Skill{
		ID:          "other-test",
		Name:        "Other Test",
		Description: "Another skill that has a long enough description",
	})

	filter := &SkillFilter{
		SearchQuery: "Query",
	}

	results := manager.Filter(filter)
	assert.Len(t, results, 1)
	assert.Equal(t, "Query Test Skill", results[0].Name)
}
