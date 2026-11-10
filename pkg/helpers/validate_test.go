package helpers

import "testing"

func TestIsGUID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"12345678-1234-1234-1234-123456789abc", true},
		{"ABCDEF01-ABCD-ABCD-ABCD-ABCDEF012345", true},
		{"", false},
		{"12345678-1234-1234-1234-123456789ab", false},
		{"12345678-1234-1234-1234-123456789abcd", false},
		{"12345678_1234-1234-1234-123456789abc", false},
		{"1234567g-1234-1234-1234-123456789abc", false},
		{"../../../v1.0/users?$top=1#aaaaaaaaaa", false},
		{"{12345678-1234-1234-1234-123456789abc}", false},
	}
	for _, tt := range tests {
		if got := IsGUID(tt.in); got != tt.want {
			t.Errorf("IsGUID(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
