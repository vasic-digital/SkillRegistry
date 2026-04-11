package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// An unrecognised Type falls back to in-memory but also returns a
	// diagnostic error so the caller can log it — the storage value
	// itself is always usable. This keeps boot resilient while making
	// configuration bugs visible.
	config := &StorageConfig{
		Type: "unknown",
	}

	storage, err := NewStorage(config)

	require.Error(t, err, "unknown type must surface a warning error")
	assert.Contains(t, err.Error(), "unknown storage type")
	assert.NotNil(t, storage, "storage must still be returned (in-memory fallback)")
	_, ok := storage.(*InMemoryStorage)
	assert.True(t, ok)
}

func TestNewStorage_PostgresMissingConfig(t *testing.T) {
	// Postgres without Host must fail fast — we do not silently fall
	// back to in-memory for an explicit postgres request.
	config := &StorageConfig{
		Type:     "postgres",
		Database: "skills",
	}

	storage, err := NewStorage(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Host")
	assert.Nil(t, storage, "postgres misconfiguration must not silently fall back")
}

func TestNewStorage_PostgresUnreachableFailsFast(t *testing.T) {
	// Point at a guaranteed-dead host:port. The ping context should
	// fire within the configured Timeout and surface a wrapped error
	// rather than hanging the caller.
	config := &StorageConfig{
		Type:     "postgres",
		Host:     "127.0.0.1",
		Port:     1, // privileged port we cannot open
		Database: "skills",
		Timeout:  500 * 1_000_000, // 500ms in nanoseconds
	}

	storage, err := NewStorage(config)
	require.Error(t, err)
	assert.Nil(t, storage)
	assert.Contains(t, err.Error(), "postgres storage")
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
