package flavor

import "testing"

func TestFromDir(t *testing.T) {
	tests := []struct {
		dir      string
		expected Flavor
	}{
		{"_retail_", KnownFlavors[0]},
		{"_beta_", KnownFlavors[1]},
		{"_classic_", KnownFlavors[2]},
		{"_classic_era_", KnownFlavors[3]},
		{"_anniversary_", KnownFlavors[4]},
		{"_ptr_", KnownFlavors[5]},
		{"_xptr_", KnownFlavors[6]},
		{"_classic_ptr_", KnownFlavors[7]},
		{"_classic_era_ptr_", KnownFlavors[8]},
		{"_classic_beta_", KnownFlavors[9]},
		{"_some_unknown_dir_", UnknownFlavor},
	}

	for _, test := range tests {
		result := FromDir(test.dir)
		if result != test.expected {
			t.Errorf("FromDir(%s) = %v; want %v", test.dir, result, test.expected)
		}
	}
}

func TestFromId(t *testing.T) {
	tests := []struct {
		id       string
		expected Flavor
	}{
		{"retail", KnownFlavors[0]},
		{"beta", KnownFlavors[1]},
		{"classic", KnownFlavors[2]},
		{"classicEra", KnownFlavors[3]},
		{"anniversary", KnownFlavors[4]},
		{"ptr", KnownFlavors[5]},
		{"xptr", KnownFlavors[6]},
		{"classicPtr", KnownFlavors[7]},
		{"classicEraPtr", KnownFlavors[8]},
		{"classicBeta", KnownFlavors[9]},
		{"some_unknown_id", UnknownFlavor},
	}

	for _, test := range tests {
		result := FromId(test.id)
		if result != test.expected {
			t.Errorf("FromId(%s) = %v; want %v", test.id, result, test.expected)
		}
	}
}
