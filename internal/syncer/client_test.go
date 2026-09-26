package syncer

import "testing"

func TestVersions(t *testing.T) {
	for _, tc := range []struct {
		v    string
		want bool // has the API
	}{
		{"0.8.0", false}, {"0.10.9", false}, {"0.11.0", true}, {"0.11.1", true}, {"1.0.0", true}, {"0.11", true}, {"0.12.0-rc1", true}, {"", false},
	} {
		if got := (Install{Version: tc.v}).API(); got != tc.want {
			t.Errorf("API(%q) = %v, want %v", tc.v, got, tc.want)
		}
	}
}
