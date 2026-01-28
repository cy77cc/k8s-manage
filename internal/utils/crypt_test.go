package utils

import (
	"testing"
)

func TestEncryptPassword(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"1", "lsolw", "lsosk"},
		{"2", "lsjdflsaa", "lsjflasfd"},
		{"3", "lsjflsdkf", "skjdflasjkdfl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptPassword(tt.in)
			if got != tt.want || err != nil {
				t.Errorf("EncryptPassword(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}
