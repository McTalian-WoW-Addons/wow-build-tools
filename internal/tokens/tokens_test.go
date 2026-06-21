package tokens

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUnknownTokenType_Error(t *testing.T) {
	err := ErrUnknownTokenType{}
	assert.Equal(t, "Unknown token type", err.Error())
}

func TestErrInvalidTokenValue_Error(t *testing.T) {
	err := ErrInvalidTokenValue{}
	assert.Equal(t, "Invalid token value", err.Error())
}

func TestIsValidToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid simple token - file-revision",
			token:    "file-revision",
			expected: true,
		},
		{
			name:     "valid simple token - project-version",
			token:    "project-version",
			expected: true,
		},
		{
			name:     "valid simple token - package-name",
			token:    "package-name",
			expected: true,
		},
		{
			name:     "valid build type token - alpha",
			token:    "alpha",
			expected: true,
		},
		{
			name:     "valid build type token - beta",
			token:    "beta",
			expected: true,
		},
		{
			name:     "valid build type token - debug",
			token:    "debug",
			expected: true,
		},
		{
			name:     "valid build type token - retail",
			token:    "retail",
			expected: true,
		},
		{
			name:     "invalid token",
			token:    "invalid-token",
			expected: false,
		},
		{
			name:     "empty string",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidToken_NormalizeToken(t *testing.T) {
	tests := []struct {
		name     string
		token    ValidToken
		expected string
	}{
		{
			name:     "file-revision",
			token:    FileRevision,
			expected: "@file-revision@",
		},
		{
			name:     "project-version",
			token:    ProjectVersion,
			expected: "@project-version@",
		},
		{
			name:     "package-name",
			token:    PackageName,
			expected: "@package-name@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.NormalizeToken()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTypeToken_NormalizeToken(t *testing.T) {
	tests := []struct {
		name     string
		token    BuildTypeToken
		expected string
	}{
		{
			name:     "alpha",
			token:    Alpha,
			expected: "@alpha@",
		},
		{
			name:     "beta",
			token:    Beta,
			expected: "@beta@",
		},
		{
			name:     "retail",
			token:    Retail,
			expected: "@retail@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.NormalizeToken()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTypeToken_GetVariants(t *testing.T) {
	tests := []struct {
		name     string
		token    BuildTypeToken
		expected BuildTypeTokenVariants
	}{
		{
			name:  "alpha variants",
			token: Alpha,
			expected: BuildTypeTokenVariants{
				Standard:    "alpha",
				StandardEnd: "end-alpha",
				Negative:    "non-alpha",
				NegativeEnd: "end-non-alpha",
			},
		},
		{
			name:  "beta variants",
			token: Beta,
			expected: BuildTypeTokenVariants{
				Standard:    "beta",
				StandardEnd: "end-beta",
				Negative:    "non-beta",
				NegativeEnd: "end-non-beta",
			},
		},
		{
			name:  "retail variants",
			token: Retail,
			expected: BuildTypeTokenVariants{
				Standard:    "retail",
				StandardEnd: "end-retail",
				Negative:    "non-retail",
				NegativeEnd: "end-non-retail",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.token.GetVariants()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimpleTokenMap_Add(t *testing.T) {
	stm := make(SimpleTokenMap)
	stm.Add(ProjectVersion, "1.0.0")
	stm.Add(PackageName, "TestPackage")

	assert.Equal(t, "1.0.0", stm[ProjectVersion])
	assert.Equal(t, "TestPackage", stm[PackageName])
}

func TestSimpleTokenMap_String(t *testing.T) {
	stm := make(SimpleTokenMap)
	stm.Add(ProjectVersion, "1.0.0")

	result := stm.String()
	assert.Contains(t, result, "project-version: 1.0.0")
}

func TestBuildTypeTokenMap_Add(t *testing.T) {
	btm := make(BuildTypeTokenMap)
	btm.Add(Alpha, true)
	btm.Add(Beta, false)

	assert.True(t, btm[Alpha])
	assert.False(t, btm[Beta])
}

func TestNormalizedSimpleTokenMap_Add(t *testing.T) {
	nstm := make(NormalizedSimpleTokenMap)
	nstm.Add(ProjectVersion, "1.0.0")

	result := nstm[ProjectVersion]
	assert.Equal(t, "@project-version@", result.Normalized)
	assert.Equal(t, "1.0.0", result.Value)
}

func TestNormalizedSimpleTokenMap_ExtendSimpleMap(t *testing.T) {
	tests := []struct {
		name           string
		initial        NormalizedSimpleTokenMap
		toExtend       *SimpleTokenMap
		expectedTokens int
	}{
		{
			name:    "extend with valid map",
			initial: make(NormalizedSimpleTokenMap),
			toExtend: &SimpleTokenMap{
				ProjectVersion: "1.0.0",
				PackageName:    "TestPackage",
			},
			expectedTokens: 2,
		},
		{
			name:           "extend with nil map",
			initial:        make(NormalizedSimpleTokenMap),
			toExtend:       nil,
			expectedTokens: 0,
		},
		{
			name: "extend existing map",
			initial: NormalizedSimpleTokenMap{
				ProjectVersion: NormalizedSimpleToken{
					Normalized: "@project-version@",
					Value:      "0.9.0",
				},
			},
			toExtend: &SimpleTokenMap{
				ProjectVersion: "1.0.0",
				PackageName:    "TestPackage",
			},
			expectedTokens: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.ExtendSimpleMap(tt.toExtend)
			assert.Equal(t, tt.expectedTokens, len(tt.initial))

			if tt.toExtend != nil {
				for token, value := range *tt.toExtend {
					result := tt.initial[token]
					assert.Equal(t, value, result.Value)
					assert.Equal(t, token.NormalizeToken(), result.Normalized)
				}
			}
		})
	}
}

func TestUniqueTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    []ValidToken
		expected int
	}{
		{
			name: "no duplicates",
			input: []ValidToken{
				ProjectVersion,
				PackageName,
				FileRevision,
			},
			expected: 3,
		},
		{
			name: "with duplicates",
			input: []ValidToken{
				ProjectVersion,
				PackageName,
				ProjectVersion,
				FileRevision,
				PackageName,
			},
			expected: 3,
		},
		{
			name:     "empty slice",
			input:    []ValidToken{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqueTokens(tt.input)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}
