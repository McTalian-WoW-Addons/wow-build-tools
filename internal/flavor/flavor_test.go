package flavor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToFlavor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Flavor
	}{
		{
			name:     "retail",
			input:    "retail",
			expected: Retail,
		},
		{
			name:     "classic",
			input:    "classic",
			expected: Classic,
		},
		{
			name:     "classicera",
			input:    "classicera",
			expected: ClassicEra,
		},
		{
			name:     "ptr",
			input:    "ptr",
			expected: Ptr,
		},
		{
			name:     "xptr",
			input:    "xptr",
			expected: Xptr,
		},
		{
			name:     "classicptr",
			input:    "classicptr",
			expected: ClassicPtr,
		},
		{
			name:     "classiceraptr",
			input:    "classiceraptr",
			expected: ClassicEraPtr,
		},
		{
			name:     "classicbeta",
			input:    "classicbeta",
			expected: ClassicBeta,
		},
		{
			name:     "unknown defaults to retail",
			input:    "unknown",
			expected: Retail,
		},
		{
			name:     "empty string defaults to retail",
			input:    "",
			expected: Retail,
		},
		{
			name:     "uppercase retail defaults to retail",
			input:    "RETAIL",
			expected: Retail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToFlavor(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlavor_ToDir(t *testing.T) {
	tests := []struct {
		name     string
		flavor   Flavor
		expected string
	}{
		{
			name:     "retail to dir",
			flavor:   Retail,
			expected: "_retail_",
		},
		{
			name:     "classic to dir",
			flavor:   Classic,
			expected: "_classic_",
		},
		{
			name:     "classicera to dir",
			flavor:   ClassicEra,
			expected: "_classic_era_",
		},
		{
			name:     "ptr to dir",
			flavor:   Ptr,
			expected: "_ptr_",
		},
		{
			name:     "xptr to dir",
			flavor:   Xptr,
			expected: "_xptr_",
		},
		{
			name:     "classicptr to dir",
			flavor:   ClassicPtr,
			expected: "_classic_ptr_",
		},
		{
			name:     "classiceraptr to dir",
			flavor:   ClassicEraPtr,
			expected: "_classic_era_ptr_",
		},
		{
			name:     "classicbeta to dir",
			flavor:   ClassicBeta,
			expected: "_classic_beta_",
		},
		{
			name:     "unknown flavor to empty dir",
			flavor:   Flavor("unknown"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.flavor.ToDir()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKnownFlavors(t *testing.T) {
	expected := []Flavor{Retail, Classic, ClassicEra, Ptr, Xptr, ClassicPtr, ClassicEraPtr, ClassicBeta}
	assert.Equal(t, expected, KnownFlavors)
	assert.Len(t, KnownFlavors, 8)
}
