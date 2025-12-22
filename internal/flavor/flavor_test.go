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
		{"_ptr_", KnownFlavors[4]},
		{"_xptr_", KnownFlavors[5]},
		{"_classic_ptr_", KnownFlavors[6]},
		{"_classic_era_ptr_", KnownFlavors[7]},
		{"_classic_beta_", KnownFlavors[8]},
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
		{"ptr", KnownFlavors[4]},
		{"xptr", KnownFlavors[5]},
		{"classicPtr", KnownFlavors[6]},
		{"classicEraPtr", KnownFlavors[7]},
		{"classicBeta", KnownFlavors[8]},
		{"some_unknown_id", UnknownFlavor},
	}

	for _, test := range tests {
		result := FromId(test.id)
		if result != test.expected {
			t.Errorf("FromId(%s) = %v; want %v", test.id, result, test.expected)
		}
	}
}
