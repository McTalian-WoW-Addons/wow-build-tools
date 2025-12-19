package external

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLastUpdatedHelper(t *testing.T) {
	log := logger.NewUnbufferedLogGroup("test")
	helper := NewLastUpdatedHelper("/cache", "prefix", true, log)

	assert.Equal(t, "/cache", helper.CacheDir)
	assert.Equal(t, "prefix", helper.FilePrefix)
	assert.True(t, helper.Force)
	assert.Equal(t, log, helper.Log)
}

func TestLastUpdatedHelper_FilePath(t *testing.T) {
	helper := &LastUpdatedHelper{
		CacheDir:   "/test/cache",
		FilePrefix: "external",
	}

	tests := []struct {
		name     string
		tag      string
		expected string
	}{
		{
			name:     "with tag",
			tag:      "v1.0.0",
			expected: filepath.Join("/test/cache", "external_v1.0.0"),
		},
		{
			name:     "without tag",
			tag:      "",
			expected: filepath.Join("/test/cache", "external"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.FilePath(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLastUpdatedHelper_WriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.NewUnbufferedLogGroup("test")

	helper := &LastUpdatedHelper{
		CacheDir:   tmpDir,
		FilePrefix: "test",
		Log:        log,
	}

	filePath := helper.FilePath("")

	// Write a timestamp
	err := helper.Write(filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Read and verify format
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	_, err = time.Parse(time.RFC3339, string(data))
	assert.NoError(t, err, "timestamp should be in RFC3339 format")
}

func TestLastUpdatedHelper_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.NewUnbufferedLogGroup("test")

	helper := &LastUpdatedHelper{
		CacheDir:   tmpDir,
		FilePrefix: "test",
		Log:        log,
	}

	filePath := helper.FilePath("")

	// Create a file
	err := helper.Write(filePath)
	require.NoError(t, err)

	// Delete it
	err = helper.Delete(filePath)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))

	// Delete non-existent file should not error
	err = helper.Delete(filePath)
	assert.NoError(t, err)
}

func TestLastUpdatedHelper_IsStale(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.NewUnbufferedLogGroup("test")

	helper := &LastUpdatedHelper{
		CacheDir:   tmpDir,
		FilePrefix: "test",
		Log:        log,
	}

	t.Run("file does not exist", func(t *testing.T) {
		filePath := helper.FilePath("nonexistent")
		stale, err := helper.IsStale(filePath, 1*time.Hour)
		assert.NoError(t, err)
		assert.True(t, stale, "non-existent file should be stale")
	})

	t.Run("fresh file is not stale", func(t *testing.T) {
		filePath := helper.FilePath("fresh")
		err := helper.Write(filePath)
		require.NoError(t, err)

		stale, err := helper.IsStale(filePath, 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, stale, "recently written file should not be stale")
	})

	t.Run("old file is stale", func(t *testing.T) {
		filePath := helper.FilePath("old")

		// Write an old timestamp
		oldTime := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
		err := os.WriteFile(filePath, []byte(oldTime), 0644)
		require.NoError(t, err)

		stale, err := helper.IsStale(filePath, 1*time.Hour)
		assert.NoError(t, err)
		assert.True(t, stale, "old file should be stale")

		// File should be deleted
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err), "stale file should be deleted")
	})

	t.Run("invalid timestamp format", func(t *testing.T) {
		filePath := helper.FilePath("invalid")

		// Write invalid timestamp
		err := os.WriteFile(filePath, []byte("not-a-timestamp"), 0644)
		require.NoError(t, err)

		stale, err := helper.IsStale(filePath, 1*time.Hour)
		assert.NoError(t, err)
		assert.True(t, stale, "file with invalid timestamp should be stale")

		// File should be deleted
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err), "invalid file should be deleted")
	})

	t.Run("exactly at duration boundary", func(t *testing.T) {
		filePath := helper.FilePath("boundary")

		// Write timestamp exactly at the boundary (should not be stale)
		boundaryTime := time.Now().Add(-59 * time.Minute).Format(time.RFC3339)
		err := os.WriteFile(filePath, []byte(boundaryTime), 0644)
		require.NoError(t, err)

		stale, err := helper.IsStale(filePath, 1*time.Hour)
		assert.NoError(t, err)
		assert.False(t, stale, "file at boundary should not be stale")
	})
}
