package migrate

import "testing"

func TestValidateISIN(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "valid danish isin", value: "DK0061538600", want: true},
		{name: "valid finnish isin", value: "FI0008002373", want: true},
		{name: "invalid free text", value: "Udstedt uden", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := validateISIN(tc.value)
			if got != tc.want {
				t.Fatalf("validateISIN(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
