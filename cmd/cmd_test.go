package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/McTalian/wow-build-tools/internal/build"
	"github.com/McTalian/wow-build-tools/internal/cmdargs"
	"github.com/McTalian/wow-build-tools/internal/logger"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func resetBuildParams() {
	build.BuildParams = &build.BuildArgs{}
}

func resetCommands() {
	// Reset all command flags to their default values
	rootCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := flag.Value.Set(flag.DefValue)
		if err != nil {
			panic(fmt.Sprintf("Failed to reset flag %s: %v", flag.Name, err))
		}
		flag.Changed = false
	})
	for _, cmd := range rootCmd.Commands() {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			err := flag.Value.Set(flag.DefValue)
			if err != nil {
				panic(fmt.Sprintf("Failed to reset flag %s: %v", flag.Name, err))
			}
			flag.Changed = false
		})
	}
}

func TestBuildHelp(t *testing.T) {
	defer resetBuildParams()
	defer resetCommands()
	logger.InitLogger()
	rootCmd.SetArgs([]string{"build", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected no error when running 'build --help', got: %v", err)
	}
}

func TestBuildMissingToc(t *testing.T) {
	defer resetBuildParams()
	defer resetCommands()
	logger.InitLogger()
	rootCmd.SetArgs([]string{"build", "-t", "."})

	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("Expected error when running 'build' without TOC file, got nil")
	}
}

func TestRootConfigureLogging(t *testing.T) {
	defer resetCommands()
	logger.InitLogger()

	tests := []struct {
		name          string
		setupFlags    func()
		checkEmoji    func() bool
		checkColor    func() bool
		checkLogLevel func() bool
	}{
		{
			name: "default flags",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{}
			},
			checkEmoji: func() bool {
				return logger.Build != ""
			},
			checkColor: func() bool {
				return !color.NoColor
			},
			checkLogLevel: func() bool {
				// Default level should be INFO
				return true
			},
		},
		{
			name: "verbose flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					LevelVerbose: true,
				}
			},
			checkEmoji: func() bool {
				return logger.Build != ""
			},
			checkColor: func() bool {
				return !color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
		{
			name: "debug flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					LevelDebug: true,
				}
			},
			checkEmoji: func() bool {
				return logger.Build != ""
			},
			checkColor: func() bool {
				return !color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
		{
			name: "no-emoji flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					NoEmoji: true,
				}
			},
			checkEmoji: func() bool {
				return logger.Build == ""
			},
			checkColor: func() bool {
				return !color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
		{
			name: "no-color flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					NoColor: true,
				}
			},
			checkEmoji: func() bool {
				// Emoji should still be enabled with no-color
				return logger.Build != ""
			},
			checkColor: func() bool {
				return color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
		{
			name: "boring flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					Boring: true,
				}
			},
			checkEmoji: func() bool {
				return logger.Build == ""
			},
			checkColor: func() bool {
				return color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
		{
			name: "multiple flags - verbose and no-emoji",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					LevelVerbose: true,
					NoEmoji:      true,
				}
			},
			checkEmoji: func() bool {
				return logger.Build == ""
			},
			checkColor: func() bool {
				return !color.NoColor
			},
			checkLogLevel: func() bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			logger.InitLogger()
			color.NoColor = false
			// Reset emoji state by re-enabling them
			logger.EnableEmoji()

			// Setup flags
			tt.setupFlags()

			// Run configureLogger
			configureLogger()

			// Check emoji
			if !tt.checkEmoji() {
				t.Errorf("Emoji state incorrect for test: %s", tt.name)
			}

			// Check color
			if !tt.checkColor() {
				t.Errorf("Color state incorrect for test: %s", tt.name)
			}

			// Check log level
			if !tt.checkLogLevel() {
				t.Errorf("Log level incorrect for test: %s", tt.name)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	defer resetCommands()
	logger.InitLogger()

	tests := []struct {
		name        string
		setupConfig func() (cleanup func(), err error)
		expectError bool
	}{
		{
			name: "no config file - uses default home directory",
			setupConfig: func() (cleanup func(), err error) {
				// No special setup needed - will use the actual config directory
				cleanup = func() {
					viper.Reset()
				}
				return cleanup, nil
			},
			expectError: false,
		},
		{
			name: "invalid yaml config file in home directory",
			setupConfig: func() (cleanup func(), err error) {
				// Create a temporary directory with an invalid YAML config file
				tempDir := t.TempDir()

				// Create the .wow-build-tools subdirectory
				wbtDir := filepath.Join(tempDir, ".wow-build-tools")
				if err := os.MkdirAll(wbtDir, 0755); err != nil {
					return nil, err
				}

				configFile := filepath.Join(wbtDir, ".wbt.yaml")
				err = os.WriteFile(configFile, []byte("invalid:\n  yaml: [\n"), 0644)
				if err != nil {
					return nil, err
				}

				// Set HOME to temp directory so config.GetConfigDir() uses it
				oldHome := os.Getenv("HOME")
				_ = os.Setenv("HOME", tempDir)

				cleanup = func() {
					_ = os.Setenv("HOME", oldHome)
					viper.Reset()
				}
				return cleanup, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setupConfig()
			if err != nil {
				t.Fatalf("Failed to setup config: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			err = loadConfig()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestPreRunE(t *testing.T) {
	defer resetCommands()
	logger.InitLogger()

	tests := []struct {
		name        string
		setupFlags  func()
		setupConfig func() (cleanup func(), err error)
		expectError bool
	}{
		{
			name: "successful pre-run",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{}
			},
			setupConfig: func() (cleanup func(), err error) {
				cleanup = func() {
					viper.Reset()
				}
				return cleanup, nil
			},
			expectError: false,
		},
		{
			name: "pre-run with verbose flag",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{
					LevelVerbose: true,
				}
			},
			setupConfig: func() (cleanup func(), err error) {
				cleanup = func() {
					viper.Reset()
				}
				return cleanup, nil
			},
			expectError: false,
		},
		{
			name: "pre-run with invalid config file",
			setupFlags: func() {
				cmdargs.RootParams = &cmdargs.RootArgs{}
			},
			setupConfig: func() (cleanup func(), err error) {
				tempDir := t.TempDir()

				// Create the .wow-build-tools subdirectory
				wbtDir := filepath.Join(tempDir, ".wow-build-tools")
				if err := os.MkdirAll(wbtDir, 0755); err != nil {
					return nil, err
				}

				configFile := filepath.Join(wbtDir, ".wbt.yaml")
				err = os.WriteFile(configFile, []byte("invalid:\n  yaml: [\n"), 0644)
				if err != nil {
					return nil, err
				}

				// Set HOME to temp directory so config.GetConfigDir() uses it
				oldHome := os.Getenv("HOME")
				_ = os.Setenv("HOME", tempDir)

				cleanup = func() {
					_ = os.Setenv("HOME", oldHome)
					viper.Reset()
				}
				return cleanup, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			logger.InitLogger()
			color.NoColor = false

			tt.setupFlags()
			cleanup, err := tt.setupConfig()
			if err != nil {
				t.Fatalf("Failed to setup config: %v", err)
			}
			if cleanup != nil {
				defer cleanup()
			}

			err = preRunE(nil, nil)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
