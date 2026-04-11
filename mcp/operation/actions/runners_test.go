package actions

import (
	"testing"
)

func TestSlimActionRunner(t *testing.T) {
	tests := []struct {
		name     string
		runner   *ActionRunner
		expected map[string]interface{}
	}{
		{
			name: "complete runner",
			runner: &ActionRunner{
				ID:         1,
				Name:       "metal",
				UUID:       "uuid-123",
				Status:     "online",
				Online:     true,
				Busy:       false,
				Version:    "1.22.5",
				Labels:     []string{"metal", "self-hosted"},
				LastOnline: "2024-01-01T00:00:00Z",
			},
			expected: map[string]interface{}{
				"id":       int64(1),
				"name":     "metal",
				"uuid":     "uuid-123",
				"status":   "online",
				"online":   true,
				"busy":     false,
				"version":  "1.22.5",
				"labels":   []string{"metal", "self-hosted"},
			},
		},
		{
			name: "runner without labels",
			runner: &ActionRunner{
				ID:      2,
				Name:    "cloud-1",
				UUID:    "uuid-456",
				Status:  "offline",
				Online:  false,
				Busy:    false,
				Version: "1.22.5",
			},
			expected: map[string]interface{}{
				"id":      int64(2),
				"name":    "cloud-1",
				"uuid":    "uuid-456",
				"status":  "offline",
				"online":  false,
				"busy":    false,
				"version": "1.22.5",
				"labels":  []string{},
			},
		},
		{
			name:     "nil runner",
			runner:   nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionRunner(tt.runner)
			if !runnerMapsEqual(result, tt.expected) {
				t.Fatalf("slimActionRunner() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSlimActionRunners(t *testing.T) {
	tests := []struct {
		name     string
		runners  []*ActionRunner
		expected map[string]interface{}
	}{
		{
			name: "multiple runners",
			runners: []*ActionRunner{
				{
					ID:      1,
					Name:    "metal",
					Status:  "online",
					Online:  true,
					Labels:  []string{"metal"},
				},
				{
					ID:      2,
					Name:    "cloud-1",
					Status:  "offline",
					Online:  false,
					Labels:  []string{"cloud-1"},
				},
			},
			expected: map[string]interface{}{
				"total_count": 2,
			},
		},
		{
			name:    "empty list",
			runners: []*ActionRunner{},
			expected: map[string]interface{}{
				"total_count": 0,
				"runners":     []interface{}{},
			},
		},
		{
			name:     "nil runners",
			runners:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionRunners(tt.runners)
			if result["total_count"] != tt.expected["total_count"] {
				t.Fatalf("total_count = %v, want %v", result["total_count"], tt.expected["total_count"])
			}
		})
	}
}

func TestSlimActionRunner_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		runner *ActionRunner
		check  func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "runner with zero ID",
			runner: &ActionRunner{
				ID:     0,
				Name:   "zero-runner",
				Status: "offline",
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["id"] != int64(0) {
					t.Errorf("expected ID 0, got %v", result["id"])
				}
			},
		},
		{
			name: "runner with empty name",
			runner: &ActionRunner{
				ID:     1,
				Name:   "",
				Status: "online",
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["name"] != "" {
					t.Errorf("expected empty name, got %v", result["name"])
				}
			},
		},
		{
			name: "runner with empty labels",
			runner: &ActionRunner{
				ID:     2,
				Name:   "no-labels",
				Status: "online",
				Labels: []string{},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				labels, ok := result["labels"].([]string)
				if !ok || len(labels) != 0 {
					t.Errorf("expected empty labels slice, got %v", result["labels"])
				}
			},
		},
		{
			name: "runner with empty LastOnline",
			runner: &ActionRunner{
				ID:         3,
				Name:       "never-online",
				Status:     "offline",
				Online:     false,
				LastOnline: "",
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["last_online"]; exists {
					t.Error("should not have last_online when empty")
				}
			},
		},
		{
			name: "runner with many labels",
			runner: &ActionRunner{
				ID:     4,
				Name:   "many-labels",
				Status: "online",
				Labels: []string{"label1", "label2", "label3", "label4", "label5"},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				labels, ok := result["labels"].([]string)
				if !ok || len(labels) != 5 {
					t.Errorf("expected 5 labels, got %v", result["labels"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionRunner(tt.runner)
			if result == nil {
				t.Fatal("slimActionRunner returned nil")
			}
			tt.check(t, result)
		})
	}
}

func TestSlimActionRunners_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		runners  []*ActionRunner
		check    func(t *testing.T, result map[string]interface{})
	}{
		{
			name:    "nil runners slice",
			runners: nil,
			check: func(t *testing.T, result map[string]interface{}) {
				if result["total_count"] != 0 {
					t.Errorf("expected total_count 0, got %v", result["total_count"])
				}
				runners, ok := result["runners"].([]interface{})
				if !ok || len(runners) != 0 {
					t.Errorf("expected empty runners slice, got %v", result["runners"])
				}
			},
		},
		{
			name:    "single runner",
			runners: []*ActionRunner{{ID: 1, Name: "single"}},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["total_count"] != 1 {
					t.Errorf("expected total_count 1, got %v", result["total_count"])
				}
				runners, ok := result["runners"].([]map[string]interface{})
				if !ok || len(runners) != 1 {
					t.Errorf("expected 1 runner, got %v", result["runners"])
				}
			},
		},
		{
			name: "many runners",
			runners: []*ActionRunner{
				{ID: 1, Name: "runner1"},
				{ID: 2, Name: "runner2"},
				{ID: 3, Name: "runner3"},
				{ID: 4, Name: "runner4"},
				{ID: 5, Name: "runner5"},
			},
			check: func(t *testing.T, result map[string]interface{}) {
				if result["total_count"] != 5 {
					t.Errorf("expected total_count 5, got %v", result["total_count"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionRunners(tt.runners)
			tt.check(t, result)
		})
	}
}

func TestRunnerMapsEqual_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]interface{}
		b        map[string]interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        map[string]interface{}{"key": "value"},
			b:        nil,
			expected: false,
		},
		{
			name:     "different lengths",
			a:        map[string]interface{}{"a": 1},
			b:        map[string]interface{}{"a": 1, "b": 2},
			expected: false,
		},
		{
			name:     "same keys different values",
			a:        map[string]interface{}{"key": "value1"},
			b:        map[string]interface{}{"key": "value2"},
			expected: false,
		},
		{
			name:     "empty maps",
			a:        map[string]interface{}{},
			b:        map[string]interface{}{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runnerMapsEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("runnerMapsEqual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// runnerMapsEqual compares two map[string]interface{} values for equality
func runnerMapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		// Simple comparison for basic types
		switch vv := v.(type) {
		case []string:
			bvv, ok := bv.([]string)
			if !ok || len(vv) != len(bvv) {
				return false
			}
			for i, sv := range vv {
				if sv != bvv[i] {
					return false
				}
			}
		default:
			if v != bv {
				return false
			}
		}
	}
	return true
}
