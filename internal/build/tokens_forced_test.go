package build

import (
	"testing"

	"github.com/McTalian/wow-build-tools/internal/repo"
	"github.com/McTalian/wow-build-tools/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPopulateTokens_ForceBuildType(t *testing.T) {
	tests := []struct {
		name                string
		tag                 string
		forceAlpha          bool
		forceBeta           bool
		forceDev            bool
		expectedReleaseType string
		expectAlphaToken    bool
		expectBetaToken     bool
		expectDebugToken    bool
		expectedAlphaFlag   string
		expectedBetaFlag    string
	}{
		{
			name:                "auto-detect from beta tag",
			tag:                 "v1.2.3-beta.1",
			expectedReleaseType: "beta",
			expectAlphaToken:    false,
			expectBetaToken:     true,
			expectDebugToken:    false,
			expectedAlphaFlag:   "",
			expectedBetaFlag:    "-beta",
		},
		{
			name:                "force alpha overrides beta tag",
			tag:                 "v1.2.3-beta.1",
			forceAlpha:          true,
			expectedReleaseType: "alpha",
			expectAlphaToken:    true,
			expectBetaToken:     false,
			expectDebugToken:    false,
			expectedAlphaFlag:   "-alpha",
			expectedBetaFlag:    "",
		},
		{
			name:                "force beta with no tag",
			forceBeta:           true,
			expectedReleaseType: "beta",
			expectAlphaToken:    false,
			expectBetaToken:     true,
			expectDebugToken:    false,
			expectedAlphaFlag:   "",
			expectedBetaFlag:    "-beta",
		},
		{
			name:                "force dev enables debug token",
			tag:                 "v1.2.3",
			forceDev:            true,
			expectedReleaseType: "alpha",
			expectAlphaToken:    true,
			expectBetaToken:     false,
			expectDebugToken:    true,
			expectedAlphaFlag:   "-alpha",
			expectedBetaFlag:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcsRepo := &repo.MockVcsRepo{
				GetInjectionValuesFunc: func(stm *tokens.SimpleTokenMap) error {
					(*stm)[tokens.ProjectVersion] = "1.0.0"
					return nil
				},
				GetCurrentTagFunc: func() string {
					return tt.tag
				},
			}

			result, err := populateTokens(populateTokensArgs{
				projectName:       "TestAddon",
				buildTimestampStr: "1710000000",
				buildDate:         "2026-04-21",
				buildDateIso:      "2026-04-21T00:00:00Z",
				buildDateInteger:  "20260421000000",
				buildYear:         "2026",
				forceAlpha:        tt.forceAlpha,
				forceBeta:         tt.forceBeta,
				forceDev:          tt.forceDev,
				vR:                vcsRepo,
			})
			require.NoError(t, err)

			assert.Equal(t, tt.expectedReleaseType, result.releaseType)
			assert.Equal(t, tt.expectAlphaToken, result.bTTM[tokens.Alpha])
			assert.Equal(t, tt.expectBetaToken, result.bTTM[tokens.Beta])
			assert.Equal(t, tt.expectDebugToken, result.bTTM[tokens.Debug])
			assert.Equal(t, tt.expectedAlphaFlag, result.flags[tokens.AlphaFlag])
			assert.Equal(t, tt.expectedBetaFlag, result.flags[tokens.BetaFlag])
		})
	}
}
