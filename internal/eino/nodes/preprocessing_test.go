package nodes

import (
	"context"
	"testing"
)

func TestPreprocessQueryToString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "normalize multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "remove control chars",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "preserve newlines and tabs",
			input:    "hello\nworld\ttab",
			expected: "hello world tab",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PreprocessQueryToString(context.Background(), tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPreprocessQuery(t *testing.T) {
	ctx := context.Background()

	input := &PreprocessInput{
		Query:    "  test query  ",
		UserType: "user1",
	}

	output, err := PreprocessQuery(ctx, input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if output.Query != "test query" {
		t.Errorf("expected trimmed query, got %q", output.Query)
	}

	if output.UserType != "user1" {
		t.Errorf("expected user_type preserved, got %q", output.UserType)
	}
}

