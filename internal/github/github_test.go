package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGitHubAction(t *testing.T) {
	tests := []struct {
		name     string
		ci       string
		actions  string
		expected bool
	}{
		{
			name:     "both env vars set to true",
			ci:       "true",
			actions:  "true",
			expected: true,
		},
		{
			name:     "CI false",
			ci:       "false",
			actions:  "true",
			expected: false,
		},
		{
			name:     "GITHUB_ACTIONS false",
			ci:       "true",
			actions:  "false",
			expected: false,
		},
		{
			name:     "both false",
			ci:       "false",
			actions:  "false",
			expected: false,
		},
		{
			name:     "CI not set",
			ci:       "",
			actions:  "true",
			expected: false,
		},
		{
			name:     "GITHUB_ACTIONS not set",
			ci:       "true",
			actions:  "",
			expected: false,
		},
		{
			name:     "both not set",
			ci:       "",
			actions:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origCI := os.Getenv("CI")
			origActions := os.Getenv("GITHUB_ACTIONS")

			// Set test values
			if tt.ci != "" {
				_ = os.Setenv("CI", tt.ci)
			} else {
				_ = os.Unsetenv("CI")
			}

			if tt.actions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", tt.actions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}

			// Test
			result := IsGitHubAction()
			assert.Equal(t, tt.expected, result)

			// Restore
			if origCI != "" {
				_ = os.Setenv("CI", origCI)
			} else {
				_ = os.Unsetenv("CI")
			}

			if origActions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", origActions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}
		})
	}
}

func TestGetRunnerTempDir(t *testing.T) {
	tests := []struct {
		name          string
		ci            string
		actions       string
		runnerTemp    string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "in GitHub Actions",
			ci:            "true",
			actions:       "true",
			runnerTemp:    "/tmp/runner",
			expectError:   false,
			expectedValue: "/tmp/runner",
		},
		{
			name:          "not in GitHub Actions",
			ci:            "false",
			actions:       "false",
			runnerTemp:    "/tmp/runner",
			expectError:   true,
			expectedValue: "",
		},
		{
			name:          "in GitHub Actions but RUNNER_TEMP empty",
			ci:            "true",
			actions:       "true",
			runnerTemp:    "",
			expectError:   false,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and set env vars
			origCI := os.Getenv("CI")
			origActions := os.Getenv("GITHUB_ACTIONS")
			origRunnerTemp := os.Getenv("RUNNER_TEMP")

			if tt.ci != "" {
				_ = os.Setenv("CI", tt.ci)
			} else {
				_ = os.Unsetenv("CI")
			}

			if tt.actions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", tt.actions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}

			if tt.runnerTemp != "" {
				_ = os.Setenv("RUNNER_TEMP", tt.runnerTemp)
			} else {
				_ = os.Unsetenv("RUNNER_TEMP")
			}

			// Test
			result, err := GetRunnerTempDir()

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}

			// Restore
			if origCI != "" {
				_ = os.Setenv("CI", origCI)
			} else {
				_ = os.Unsetenv("CI")
			}

			if origActions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", origActions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}

			if origRunnerTemp != "" {
				_ = os.Setenv("RUNNER_TEMP", origRunnerTemp)
			} else {
				_ = os.Unsetenv("RUNNER_TEMP")
			}
		})
	}
}

func TestOutput(t *testing.T) {
	tests := []struct {
		name         string
		ci           string
		actions      string
		outputFile   string
		createFile   bool
		key          string
		value        string
		expectError  bool
		expectOutput bool
	}{
		{
			name:         "not in GitHub Actions",
			ci:           "false",
			actions:      "false",
			key:          "test",
			value:        "value",
			expectError:  false,
			expectOutput: false,
		},
		{
			name:         "in GitHub Actions with valid file",
			ci:           "true",
			actions:      "true",
			createFile:   true,
			key:          "test_key",
			value:        "test_value",
			expectError:  false,
			expectOutput: true,
		},
		{
			name:         "in GitHub Actions without GITHUB_OUTPUT",
			ci:           "true",
			actions:      "true",
			outputFile:   "",
			key:          "test",
			value:        "value",
			expectError:  false,
			expectOutput: false,
		},
		{
			name:         "in GitHub Actions with non-existent file",
			ci:           "true",
			actions:      "true",
			outputFile:   "/nonexistent/path/output",
			createFile:   false,
			key:          "test",
			value:        "value",
			expectError:  true,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save env vars
			origCI := os.Getenv("CI")
			origActions := os.Getenv("GITHUB_ACTIONS")
			origOutput := os.Getenv("GITHUB_OUTPUT")

			// Set CI environment
			if tt.ci != "" {
				_ = os.Setenv("CI", tt.ci)
			} else {
				_ = os.Unsetenv("CI")
			}

			if tt.actions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", tt.actions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}

			var tmpFile string
			if tt.createFile {
				tmpDir := t.TempDir()
				tmpFile = filepath.Join(tmpDir, "github_output")
				err := os.WriteFile(tmpFile, []byte(""), 0644)
				require.NoError(t, err)
				_ = os.Setenv("GITHUB_OUTPUT", tmpFile)
			} else if tt.outputFile != "" {
				_ = os.Setenv("GITHUB_OUTPUT", tt.outputFile)
			} else {
				_ = os.Unsetenv("GITHUB_OUTPUT")
			}

			// Test
			err := Output(tt.key, tt.value)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify output was written
			if tt.expectOutput {
				content, err := os.ReadFile(tmpFile)
				require.NoError(t, err)
				assert.Contains(t, string(content), tt.key+"="+tt.value)
			}

			// Restore
			if origCI != "" {
				_ = os.Setenv("CI", origCI)
			} else {
				_ = os.Unsetenv("CI")
			}

			if origActions != "" {
				_ = os.Setenv("GITHUB_ACTIONS", origActions)
			} else {
				_ = os.Unsetenv("GITHUB_ACTIONS")
			}

			if origOutput != "" {
				_ = os.Setenv("GITHUB_OUTPUT", origOutput)
			} else {
				_ = os.Unsetenv("GITHUB_OUTPUT")
			}
		})
	}
}
