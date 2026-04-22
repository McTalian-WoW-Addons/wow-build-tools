package toc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompatibleInstallFlavorsFromInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []int
		expected   []string
	}{
		{
			name:       "retail interfaces map to retail-family installs",
			interfaces: []int{110007},
			expected:   []string{"beta", "ptr", "retail", "xptr"},
		},
		{
			name:       "classic progression interfaces map to classic-family installs",
			interfaces: []int{40401, 50000},
			expected:   []string{"classic", "classicBeta", "classicPtr"},
		},
		{
			name:       "classic era interfaces map to classic era installs",
			interfaces: []int{11505},
			expected:   []string{"classicEra", "classicEraPtr"},
		},
		{
			name:       "anniversary interfaces map to anniversary installs",
			interfaces: []int{20505},
			expected:   []string{"anniversary"},
		},
		{
			name:       "mixed interfaces produce a deduplicated union",
			interfaces: []int{110007, 11505, 50000},
			expected:   []string{"beta", "classic", "classicBeta", "classicEra", "classicEraPtr", "classicPtr", "ptr", "retail", "xptr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compatibleFlavors := CompatibleInstallFlavorsFromInterfaces(tt.interfaces)
			actual := make([]string, 0, len(compatibleFlavors))
			for _, installFlavor := range compatibleFlavors {
				actual = append(actual, installFlavor.Id)
			}

			assert.ElementsMatch(t, tt.expected, actual)
		})
	}
}

func TestTocCompatibleInstallFlavorsSortsOutput(t *testing.T) {
	toc := &Toc{Interface: []int{50000, 110007}}

	compatibleFlavors := toc.CompatibleInstallFlavors()
	actual := make([]string, 0, len(compatibleFlavors))
	for _, installFlavor := range compatibleFlavors {
		actual = append(actual, installFlavor.Id)
	}

	require.NotEmpty(t, actual)
	assert.Equal(t, []string{"beta", "classic", "classicBeta", "classicPtr", "ptr", "retail", "xptr"}, actual)
}
