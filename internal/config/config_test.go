package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/flavor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetWoWPath(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Create mock WoW directory structure
	retailDir := filepath.Join(tempDir, flavor.FromId("retail").Dir)
	classicDir := filepath.Join(tempDir, flavor.FromId("classic").Dir)
	require.NoError(t, os.MkdirAll(retailDir, 0755))
	require.NoError(t, os.MkdirAll(classicDir, 0755))

	// Reset viper
	viper.Reset()

	tests := []struct {
		name      string
		value     []string
		wantError bool
	}{
		{
			name:      "set path with value",
			value:     []string{tempDir},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			err := setWoWPath(nil, tt.value...)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tempDir, viper.GetString("wowPath.base"))
			}
		})
	}
}

func TestSetFlavorPath(t *testing.T) {
	// Reset viper
	viper.Reset()

	tests := []struct {
		name      string
		flavor    Flavor
		value     []string
		wantError bool
	}{
		{
			name:      "set retail path",
			flavor:    flavor.FromId("retail"),
			value:     []string{"/path/to/retail"},
			wantError: false,
		},
		{
			name:      "set classic path",
			flavor:    flavor.FromId("classic"),
			value:     []string{"/path/to/classic"},
			wantError: false,
		},
		{
			name:      "set ptr path",
			flavor:    flavor.FromId("ptr"),
			value:     []string{"/path/to/ptr"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			err := setFlavorPath(nil, tt.flavor, tt.value...)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.value[0], viper.GetString("wowPath."+tt.flavor.Id))
			}
		})
	}
}

func TestRunConfig(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() {
		err := os.Chdir(originalWd)
		if err != nil {
			t.Fatalf("Failed to change directory back to original: %v", err)
		}
	}()

	require.NoError(t, os.Chdir(tempDir))

	tests := []struct {
		name      string
		args      []string
		setup     func()
		wantError bool
	}{
		{
			name:      "invalid primary argument",
			args:      []string{"invalid"},
			setup:     func() { viper.Reset() },
			wantError: true,
		},
		{
			name:      "wowPath without subcommand",
			args:      []string{"wowPath"},
			setup:     func() { viper.Reset() },
			wantError: true,
		},
		{
			name:      "wowPath with invalid subcommand",
			args:      []string{"wowPath", "invalid"},
			setup:     func() { viper.Reset() },
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			// Create a temporary config file to avoid prompts
			viper.SetConfigFile(filepath.Join(tempDir, ".wbt.yaml"))
			err := viper.WriteConfig()
			assert.NoError(t, err)

			localConfigDisabled = true
			err = RunConfig(tt.args)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
