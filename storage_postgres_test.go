package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage_Memory(t *testing.T) {
	config := &StorageConfig{
		Type: "memory",
	}

	storage, err := NewStorage(config)

	assert.NoError(t, err)
	assert.NotNil(t, storage)
	_, ok := storage.(*InMemoryStorage)
	assert.True(t, ok)
}

func TestNewStorage_Default(t *testing.T) {
	config := &StorageConfig{
		Type: "unknown",
	}

	storage, err := NewStorage(config)

	assert.NoError(t, err)
	assert.NotNil(t, storage)
	_, ok := storage.(*InMemoryStorage)
	assert.True(t, ok)
}

func TestStorageConfig_Defaults(t *testing.T) {
	config := &StorageConfig{
		Type:     "postgres",
		Username: "user",
		Password: "pass",
	}

	assert.Equal(t, "postgres", config.Type)
	assert.Equal(t, "", config.Host)
	assert.Equal(t, 0, config.Port)
	assert.Equal(t, "", config.Database)
	assert.Equal(t, "", config.SSLMode)
}
