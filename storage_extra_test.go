package agents

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStorage_Update(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	skill := &Skill{
		ID:          "update-skill",
		Name:        "Original Name",
		Description: "Original description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := storage.Save(ctx, skill)
	require.NoError(t, err)

	// Update the skill
	skill.Name = "Updated Name"
	skill.Description = "Updated description"
	err = storage.Save(ctx, skill)
	require.NoError(t, err)

	retrieved, err := storage.Get(ctx, "update-skill")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
}

func TestInMemoryStorage_GetByStatus_NoResults(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "active-skill",
		Name:        "Active Skill",
		Description: "An active skill",
		Status:      SkillStatusActive,
	})

	results := storage.GetByStatus(SkillStatusDisabled)
	assert.Empty(t, results)
}

func TestInMemoryStorage_GetByCategory_NoResults(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "code-skill",
		Name:        "Code Skill",
		Description: "A code skill",
		Category:    SkillCategoryCode,
	})

	results := storage.GetByCategory(SkillCategoryData)
	assert.Empty(t, results)
}

func TestInMemoryStorage_Search_NoResults(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "my-skill",
		Name:        "My Skill",
		Description: "My description",
	})

	results := storage.Search("nonexistent")
	assert.Empty(t, results)
}

func TestInMemoryStorage_Search_EmptyQuery(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	storage.Save(ctx, &Skill{
		ID:          "my-skill",
		Name:        "My Skill",
		Description: "My description",
	})

	results := storage.Search("")
	// Empty string should match everything since contains(s, "") returns true
	assert.Len(t, results, 1)
}

func TestInMemoryStorage_ConcurrentAccess(t *testing.T) {
	storage := NewInMemoryStorage()
	ctx := context.Background()

	// Test concurrent reads and writes
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			storage.Save(ctx, &Skill{
				ID:          "concurrent-skill",
				Name:        "Concurrent Skill",
				Description: "A skill for concurrent testing",
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			storage.Get(ctx, "concurrent-skill")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			storage.Exists(ctx, "concurrent-skill")
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestNewStorage_InvalidType(t *testing.T) {
	// Post-2026-04-11: NewStorage surfaces a warning error for
	// unrecognised types while still returning a usable in-memory
	// fallback. Silent fallback made configuration bugs invisible.
	config := &StorageConfig{
		Type: "invalid",
	}

	storage, err := NewStorage(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown storage type")
	assert.NotNil(t, storage)
	_, ok := storage.(*InMemoryStorage)
	assert.True(t, ok)
}

func TestStorageConfig_FullConfig(t *testing.T) {
	config := &StorageConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "require",
		Options: map[string]interface{}{
			"max_connections": 10,
		},
	}

	assert.Equal(t, "postgres", config.Type)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "user", config.Username)
	assert.Equal(t, "pass", config.Password)
	assert.Equal(t, "require", config.SSLMode)
	assert.Equal(t, 10, config.Options["max_connections"])
}
