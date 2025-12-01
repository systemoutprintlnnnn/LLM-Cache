package configs

import (
	"testing"
)

func TestDefaultConfigQualityValidation(t *testing.T) {
	// 测试 DefaultConfig 的 Quality 配置可以通过验证
	cfg := DefaultConfig()
	err := cfg.Quality.Validate()
	if err != nil {
		t.Errorf("DefaultConfig Quality validation failed: %v", err)
	}
}

func TestQualityConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  QualityConfig
		wantErr bool
	}{
		{
			name: "disabled quality check passes",
			config: QualityConfig{
				Enabled:   false,
				Threshold: 0.5,
			},
			wantErr: false,
		},
		{
			name: "enabled with no strategies passes",
			config: QualityConfig{
				Enabled:   true,
				Threshold: 0.5,
			},
			wantErr: false,
		},
		{
			name: "enabled with valid strategies passes",
			config: QualityConfig{
				Enabled:   true,
				Threshold: 0.5,
				Strategies: []QualityStrategy{
					{Name: "test", Weight: 1.0, Enabled: true},
				},
			},
			wantErr: false,
		},
		{
			name: "enabled with zero weight strategies fails",
			config: QualityConfig{
				Enabled:   true,
				Threshold: 0.5,
				Strategies: []QualityStrategy{
					{Name: "test", Weight: 0, Enabled: true},
				},
			},
			wantErr: true,
		},
		{
			name: "enabled with all disabled strategies fails",
			config: QualityConfig{
				Enabled:   true,
				Threshold: 0.5,
				Strategies: []QualityStrategy{
					{Name: "test", Weight: 1.0, Enabled: false},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid threshold fails",
			config: QualityConfig{
				Enabled:   true,
				Threshold: 1.5,
			},
			wantErr: true,
		},
		{
			name: "negative threshold fails",
			config: QualityConfig{
				Enabled:   true,
				Threshold: -0.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("QualityConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
