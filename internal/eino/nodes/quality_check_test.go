package nodes

import (
	"context"
	"testing"

	"llm-cache/internal/eino/config"
)

func TestQualityChecker_Check(t *testing.T) {
	cfg := &config.QualityConfig{
		Enabled:           true,
		MinQuestionLength: 5,
		MinAnswerLength:   10,
		MaxQuestionLength: 1000,
		MaxAnswerLength:   10000,
		ScoreThreshold:    0.3,
		BlacklistKeywords: []string{"spam", "forbidden"},
	}

	checker := NewQualityChecker(cfg)
	ctx := context.Background()

	tests := []struct {
		name       string
		input      *QualityCheckInput
		wantPassed bool
		wantReason string
	}{
		{
			name: "valid input",
			input: &QualityCheckInput{
				Question: "What is the weather today?",
				Answer:   "The weather is sunny and warm with clear skies.",
				UserType: "user1",
			},
			wantPassed: true,
		},
		{
			name: "question too short",
			input: &QualityCheckInput{
				Question: "Hi",
				Answer:   "This is a valid answer with enough length.",
				UserType: "user1",
			},
			wantPassed: false,
			wantReason: "question too short",
		},
		{
			name: "answer too short",
			input: &QualityCheckInput{
				Question: "What is the weather?",
				Answer:   "Sunny",
				UserType: "user1",
			},
			wantPassed: false,
			wantReason: "answer too short",
		},
		{
			name: "contains blacklisted word in question",
			input: &QualityCheckInput{
				Question: "Is this spam content?",
				Answer:   "No, this is valid content with enough length.",
				UserType: "user1",
			},
			wantPassed: false,
			wantReason: "contains blacklisted content",
		},
		{
			name: "contains blacklisted word in answer",
			input: &QualityCheckInput{
				Question: "What is the answer?",
				Answer:   "This answer contains forbidden content that should be rejected.",
				UserType: "user1",
			},
			wantPassed: false,
			wantReason: "contains blacklisted content",
		},
		{
			name: "force write bypasses check",
			input: &QualityCheckInput{
				Question:   "Hi",
				Answer:     "Sunny",
				UserType:   "user1",
				ForceWrite: true,
			},
			wantPassed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := checker.Check(ctx, tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("expected passed=%v, got passed=%v, reason=%s", tt.wantPassed, result.Passed, result.Reason)
			}

			if !tt.wantPassed && tt.wantReason != "" && result.Reason != tt.wantReason {
				t.Errorf("expected reason=%q, got reason=%q", tt.wantReason, result.Reason)
			}
		})
	}
}

func TestQualityChecker_Disabled(t *testing.T) {
	cfg := &config.QualityConfig{
		Enabled: false,
	}

	checker := NewQualityChecker(cfg)
	ctx := context.Background()

	input := &QualityCheckInput{
		Question: "X",
		Answer:   "Y",
		UserType: "user1",
	}

	result, err := checker.Check(ctx, input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected disabled checker to pass, got passed=%v", result.Passed)
	}
}

func TestCalculateQualityScore(t *testing.T) {
	tests := []struct {
		name     string
		question string
		answer   string
		minScore float64
		maxScore float64
	}{
		{
			name:     "good question and answer",
			question: "What is the weather forecast for tomorrow?",
			answer:   "Tomorrow will be sunny with temperatures around 25 degrees. It will be a great day for outdoor activities.",
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name:     "short question",
			question: "Hi?",
			answer:   "Hello! How can I help you today? I am here to assist with any questions.",
			minScore: 0.6,
			maxScore: 1.0,
		},
		{
			name:     "no question mark",
			question: "Tell me about weather",
			answer:   "The weather is the state of the atmosphere at a given time and place.",
			minScore: 0.6,
			maxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateQualityScore(tt.question, tt.answer)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score in [%v, %v], got %v", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

