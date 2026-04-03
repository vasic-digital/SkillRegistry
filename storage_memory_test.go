package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.skills)
	assert.Equal(t, 0, storage.Count())
}

func TestInMemoryStorage_Save(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill := &Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(ctx, skill)

	require.NoError(t, err)
	assert.Equal(t, 1, storage.Count())
}

func TestInMemoryStorage_Save_NilSkill(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	err := storage.Save(ctx, nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillInvalid)
}

func TestInMemoryStorage_Save_ContextCancelled(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	skill := &Skill{
		ID:          "test-skill",
		Name:        "Test Skill",
		Description: "A test skill",
	}

	err := storage.Save(ctx, skill)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryStorage_Get(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill := &Skill{
		ID:          "get-skill",
		Name:        "Get Skill",
		Description: "A skill to get",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(ctx, skill)
	require.NoError(t, err)

	retrieved, err := storage.Get(ctx, "get-skill")

	require.NoError(t, err)
	assert.Equal(t, skill.ID, retrieved.ID)
	assert.Equal(t, skill.Name, retrieved.Name)
}

func TestInMemoryStorage_Get_NotFound(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	_, err := storage.Get(ctx, "nonexistent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestInMemoryStorage_Get_ContextCancelled(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.Get(ctx, "any")

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryStorage_Delete(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill := &Skill{
		ID:          "delete-skill",
		Name:        "Delete Skill",
		Description: "A skill to delete",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(ctx, skill)
	require.NoError(t, err)

	err = storage.Delete(ctx, "delete-skill")

	require.NoError(t, err)
	assert.Equal(t, 0, storage.Count())

	_, err = storage.Get(ctx, "delete-skill")
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestInMemoryStorage_Delete_NotFound(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	err := storage.Delete(ctx, "nonexistent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestInMemoryStorage_Delete_ContextCancelled(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := storage.Delete(ctx, "any")

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryStorage_List(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill1 := &Skill{
		ID:          "list-skill-1",
		Name:        "List Skill 1",
		Description: "First skill",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	skill2 := &Skill{
		ID:          "list-skill-2",
		Name:        "List Skill 2",
		Description: "Second skill",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	storage.Save(ctx, skill1)
	storage.Save(ctx, skill2)

	skills, err := storage.List(ctx)

	require.NoError(t, err)
	assert.Len(t, skills, 2)
}

func TestInMemoryStorage_List_Empty(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skills, err := storage.List(ctx)

	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestInMemoryStorage_List_ContextCancelled(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.List(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryStorage_Exists(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill := &Skill{
		ID:          "exists-skill",
		Name:        "Exists Skill",
		Description: "A skill that exists",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	storage.Save(ctx, skill)

	exists, err := storage.Exists(ctx, "exists-skill")

	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = storage.Exists(ctx, "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestInMemoryStorage_Exists_ContextCancelled(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := storage.Exists(ctx, "any")

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInMemoryStorage_Clear(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "skill-1",
		Name:        "Skill 1",
		Description: "First skill",
	})
	storage.Save(ctx, &Skill{
		ID:          "skill-2",
		Name:        "Skill 2",
		Description: "Second skill",
	})

	assert.Equal(t, 2, storage.Count())

	storage.Clear()

	assert.Equal(t, 0, storage.Count())
}

func TestInMemoryStorage_GetByStatus(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "active-skill",
		Name:        "Active Skill",
		Description: "An active skill",
		Status:      SkillStatusActive,
	})
	storage.Save(ctx, &Skill{
		ID:          "inactive-skill",
		Name:        "Inactive Skill",
		Description: "An inactive skill",
		Status:      SkillStatusInactive,
	})

	activeSkills := storage.GetByStatus(SkillStatusActive)

	assert.Len(t, activeSkills, 1)
	assert.Equal(t, "Active Skill", activeSkills[0].Name)
}

func TestInMemoryStorage_GetByCategory(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "code-skill",
		Name:        "Code Skill",
		Description: "A code skill",
		Category:    SkillCategoryCode,
	})
	storage.Save(ctx, &Skill{
		ID:          "data-skill",
		Name:        "Data Skill",
		Description: "A data skill",
		Category:    SkillCategoryData,
	})

	codeSkills := storage.GetByCategory(SkillCategoryCode)

	assert.Len(t, codeSkills, 1)
	assert.Equal(t, "Code Skill", codeSkills[0].Name)
}

func TestInMemoryStorage_Search(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "searchable-skill",
		Name:        "Searchable Skill",
		Description: "A searchable skill with unique content",
	})
	storage.Save(ctx, &Skill{
		ID:          "other-skill",
		Name:        "Other Skill",
		Description: "Another skill",
	})

	results := storage.Search("unique")

	assert.Len(t, results, 1)
	assert.Equal(t, "searchable-skill", results[0].ID)
}

func TestInMemoryStorage_GetAll(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "skill-1",
		Name:        "Skill 1",
		Description: "First skill",
	})
	storage.Save(ctx, &Skill{
		ID:          "skill-2",
		Name:        "Skill 2",
		Description: "Second skill",
	})

	all := storage.GetAll()

	assert.Len(t, all, 2)
	assert.Contains(t, all, "skill-1")
	assert.Contains(t, all, "skill-2")
}
