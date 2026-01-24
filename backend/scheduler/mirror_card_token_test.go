package scheduler

import (
	"fmt"
	"testing"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		respData string
		want     string
	}{
		{
			name:     "JSON with token field",
			respData: `{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9","user_id":123}`,
			want:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:     "Plain text token",
			respData: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:     "Empty string",
			respData: "",
			want:     "",
		},
		{
			name:     "JSON without token field",
			respData: `{"message":"success","data":"some_value"}`,
			want:     `{"message":"success","data":"some_value"}`,
		},
		{
			name:     "Token with spaces",
			respData: "  token_value_with_spaces  ",
			want:     "token_value_with_spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractToken(tt.respData)
			if got != tt.want {
				t.Errorf("extractToken() = %v, want %v", got, tt.want)
			}
			fmt.Printf("Test %s: respData=%s, extracted=%s\n", tt.name, tt.respData, got)
		})
	}
}

func TestGetNodeURLs(t *testing.T) {
	urls := getNodeURLs()
	if len(urls) == 0 {
		t.Error("getNodeURLs() returned empty list")
	}
	fmt.Printf("Node URLs: %v\n", urls)
}

