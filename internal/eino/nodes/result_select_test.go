package nodes

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestResultSelector_SelectFirst(t *testing.T) {
	selector := NewResultSelector("first", 0.7)
	ctx := context.Background()

	docs := []*schema.Document{
		{ID: "doc1", Content: "First", MetaData: map[string]any{"score": 0.5}},
		{ID: "doc2", Content: "Second", MetaData: map[string]any{"score": 0.9}},
		{ID: "doc3", Content: "Third", MetaData: map[string]any{"score": 0.7}},
	}

	result, err := selector.Select(ctx, docs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.ID != "doc1" {
		t.Errorf("expected first doc, got %s", result.ID)
	}
}

func TestResultSelector_SelectHighestScore(t *testing.T) {
	selector := NewResultSelector("highest_score", 0.7)
	ctx := context.Background()

	docs := []*schema.Document{
		{ID: "doc1", Content: "First", MetaData: map[string]any{"score": 0.5}},
		{ID: "doc2", Content: "Second", MetaData: map[string]any{"score": 0.9}},
		{ID: "doc3", Content: "Third", MetaData: map[string]any{"score": 0.7}},
	}

	result, err := selector.Select(ctx, docs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.ID != "doc2" {
		t.Errorf("expected highest score doc (doc2), got %s", result.ID)
	}
}

func TestResultSelector_SelectTemperatureSoftmax(t *testing.T) {
	selector := NewResultSelector("temperature_softmax", 0.1) // 低温度 = 更确定性
	ctx := context.Background()

	docs := []*schema.Document{
		{ID: "doc1", Content: "First", MetaData: map[string]any{"score": 0.5}},
		{ID: "doc2", Content: "Second", MetaData: map[string]any{"score": 0.95}},
		{ID: "doc3", Content: "Third", MetaData: map[string]any{"score": 0.5}},
	}

	// 运行多次，低温度应该大部分时候选择最高分
	highScoreCount := 0
	for i := 0; i < 100; i++ {
		result, err := selector.Select(ctx, docs)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.ID == "doc2" {
			highScoreCount++
		}
	}

	// 低温度下，应该大部分时候选择最高分
	if highScoreCount < 80 {
		t.Errorf("expected high score selection to dominate, got %d/100", highScoreCount)
	}
}

func TestResultSelector_EmptyDocs(t *testing.T) {
	selector := NewResultSelector("highest_score", 0.7)
	ctx := context.Background()

	result, err := selector.Select(ctx, []*schema.Document{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for empty docs, got %v", result)
	}
}

func TestResultSelector_SingleDoc(t *testing.T) {
	selector := NewResultSelector("temperature_softmax", 0.7)
	ctx := context.Background()

	docs := []*schema.Document{
		{ID: "doc1", Content: "Only", MetaData: map[string]any{"score": 0.5}},
	}

	result, err := selector.Select(ctx, docs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.ID != "doc1" {
		t.Errorf("expected single doc, got %s", result.ID)
	}
}

func TestGetDocScore(t *testing.T) {
	tests := []struct {
		name     string
		doc      *schema.Document
		expected float64
	}{
		{
			name:     "nil doc",
			doc:      nil,
			expected: 0,
		},
		{
			name: "score in metadata",
			doc: &schema.Document{
				MetaData: map[string]any{"score": 0.75},
			},
			expected: 0.75,
		},
		{
			name: "_score in metadata",
			doc: &schema.Document{
				MetaData: map[string]any{"_score": 0.85},
			},
			expected: 0.85,
		},
		{
			name: "no score",
			doc: &schema.Document{
				MetaData: map[string]any{"other": "value"},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := getDocScore(tt.doc)
			if score != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, score)
			}
		})
	}
}
