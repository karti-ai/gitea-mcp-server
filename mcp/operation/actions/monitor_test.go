package actions

import (
	"testing"
)

func TestTrimVersionPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.22.5", "1.22.5"},
		{"V1.22.5", "1.22.5"},
		{"1.22.5", "1.22.5"},
		{"v1.23.0", "1.23.0"},
		{"2.0.0", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimVersionPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("trimVersionPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"1.22.5", []string{"1", "22", "5"}},
		{"1.23", []string{"1", "23"}},
		{"2.0.0", []string{"2", "0", "0"}},
		{"1", []string{"1"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitVersion(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitVersion(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitVersion(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseVersionForCheck(t *testing.T) {
	tests := []struct {
		version      string
		wantMajor    int
		wantMinor    int
		wantPatch    int
		wantErr      bool
	}{
		{"1.22.5", 1, 22, 5, false},
		{"1.23.0", 1, 23, 0, false},
		{"v1.22.5", 1, 22, 5, false},
		{"V1.23.0", 1, 23, 0, false},
		{"2.0.0", 2, 0, 0, false},
		{"1.22", 1, 22, 0, false},
		{"invalid", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major, minor, patch, err := parseVersionForCheck(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersionForCheck(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
					t.Errorf("parseVersionForCheck(%q) = (%d, %d, %d), want (%d, %d, %d)",
						tt.version, major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

func TestContainsPath(t *testing.T) {
	tests := []struct {
		path     string
		substr   string
		expected bool
	}{
		{".gitea/workflows/build.yml", "build.yml", true},
		{".github/workflows/test.yml", "test.yml", true},
		{".gitea/workflows/build.yml", "deploy.yml", false},
		{"build.yml", "build.yml", true},
		{"", "test", false},
		{"test", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.substr, func(t *testing.T) {
			result := containsPath(tt.path, tt.substr)
			if result != tt.expected {
				t.Errorf("containsPath(%q, %q) = %v, want %v", tt.path, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{
			name:     "string value",
			m:        map[string]any{"name": "test"},
			key:      "name",
			expected: "test",
		},
		{
			name:     "float64 value",
			m:        map[string]any{"id": float64(123)},
			key:      "id",
			expected: "123",
		},
		{
			name:     "missing key",
			m:        map[string]any{"other": "value"},
			key:      "name",
			expected: "",
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "name",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getString(%v, %q) = %q, want %q", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}

func TestParseJobSummaries(t *testing.T) {
	jobsData := []any{
		map[string]any{
			"id":           float64(1),
			"name":         "build",
			"status":       "completed",
			"conclusion":   "success",
			"started_at":   "2024-01-15T10:00:00Z",
			"completed_at": "2024-01-15T10:05:00Z",
			"steps": []any{
				map[string]any{
					"name":       "Checkout",
					"number":     float64(1),
					"status":     "completed",
					"conclusion": "success",
				},
			},
		},
	}

	summaries := parseJobSummaries(jobsData)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	if summaries[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", summaries[0].ID)
	}
	if summaries[0].Name != "build" {
		t.Errorf("expected name 'build', got %s", summaries[0].Name)
	}
	if len(summaries[0].Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(summaries[0].Steps))
	}
}

func TestParseSteps(t *testing.T) {
	stepsData := []any{
		map[string]any{
			"name":       "Checkout",
			"number":     float64(1),
			"status":     "completed",
			"conclusion": "success",
		},
		map[string]any{
			"name":       "Build",
			"number":     float64(2),
			"status":     "completed",
			"conclusion": "success",
		},
	}

	steps := parseSteps(stepsData)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}

	if steps[0].Name != "Checkout" || steps[0].Number != 1 {
		t.Errorf("first step mismatch: %+v", steps[0])
	}
	if steps[1].Name != "Build" || steps[1].Number != 2 {
		t.Errorf("second step mismatch: %+v", steps[1])
	}
}

func TestMonitorResultTypes(t *testing.T) {
	result := &MonitorResult{
		RunID:       123,
		Status:      "completed",
		Conclusion:  "success",
		WorkflowID:  "build.yml",
		Branch:      "main",
		CommitSHA:   "abc123",
		Duration:    "5m0s",
		DurationSec: 300,
		Jobs: []JobSummary{
			{
				ID:         1,
				Name:       "build",
				Status:     "completed",
				Conclusion: "success",
				Steps: []StepInfo{
					{Name: "Checkout", Number: 1, Status: "completed", Conclusion: "success"},
				},
			},
		},
		Logs: map[string]JobLogs{
			"build": {
				JobID:   1,
				JobName: "build",
				Log:     "Building...",
				Bytes:   100,
			},
		},
	}

	if result.RunID != 123 {
		t.Errorf("RunID mismatch")
	}
	if result.Status != "completed" {
		t.Errorf("Status mismatch")
	}
	if len(result.Jobs) != 1 {
		t.Errorf("Jobs length mismatch")
	}
	if len(result.Logs) != 1 {
		t.Errorf("Logs length mismatch")
	}
}

func TestTrimOnePrefix(t *testing.T) {
	tests := []struct {
		s        string
		prefix   string
		expected string
	}{
		{"v1.22.5", "v", "1.22.5"},
		{"V1.22.5", "V", "1.22.5"},
		{"1.22.5", "v", "1.22.5"},
		{"", "v", ""},
		{"v", "v", ""},
		{"test", "x", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.prefix, func(t *testing.T) {
			result := trimOnePrefix(tt.s, tt.prefix)
			if result != tt.expected {
				t.Errorf("trimOnePrefix(%q, %q) = %q, want %q", tt.s, tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestParseVersionForCheck_EdgeCases(t *testing.T) {
	tests := []struct {
		version   string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{"1.0.0", 1, 0, 0, false},
		{"0.0.1", 0, 0, 1, false},
		{"1.23.0+build", 1, 23, 0, false},
		{"1.23.0-rc1", 1, 23, 0, false},
		{"1.23", 1, 23, 0, false},
		{"1", 1, 0, 0, false},
		{"", 0, 0, 0, true},
		{"abc", 0, 0, 0, true},
		{"1.abc.5", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major, minor, patch, err := parseVersionForCheck(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersionForCheck(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
					t.Errorf("parseVersionForCheck(%q) = (%d, %d, %d), want (%d, %d, %d)",
						tt.version, major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

func TestContainsPath_EdgeCases(t *testing.T) {
	tests := []struct {
		path     string
		substr   string
		expected bool
	}{
		{"", "", false},
		{"build.yml", "", false},
		{"", "build", false},
		{"a/b/c/d.yml", "c/d.yml", true},
		{"a/b/c/d.yml", "b/c", false},
		{"build.yml", "build.yml", true},
		{"/absolute/path", "path", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.substr, func(t *testing.T) {
			result := containsPath(tt.path, tt.substr)
			if result != tt.expected {
				t.Errorf("containsPath(%q, %q) = %v, want %v", tt.path, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestGetString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{
			name:     "int value",
			m:        map[string]any{"count": int(42)},
			key:      "count",
			expected: "",
		},
		{
			name:     "bool value",
			m:        map[string]any{"active": true},
			key:      "active",
			expected: "",
		},
		{
			name:     "empty string value",
			m:        map[string]any{"name": ""},
			key:      "name",
			expected: "",
		},
		{
			name:     "large float64",
			m:        map[string]any{"id": float64(9223372036854775807)},
			key:      "id",
			expected: "9223372036854775807",
		},
		{
			name:     "zero float64",
			m:        map[string]any{"count": float64(0)},
			key:      "count",
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getString(%v, %q) = %q, want %q", tt.m, tt.key, result, tt.expected)
			}
		})
	}
}

func TestParseJobSummaries_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		jobsData []any
		wantLen  int
	}{
		{
			name:     "empty jobs",
			jobsData: []any{},
			wantLen:  0,
		},
		{
			name:     "nil jobs",
			jobsData: nil,
			wantLen:  0,
		},
		{
			name: "job without steps",
			jobsData: []any{
				map[string]any{
					"id":         float64(1),
					"name":       "build",
					"status":     "completed",
					"conclusion": "success",
				},
			},
			wantLen: 1,
		},
		{
			name: "job with empty steps",
			jobsData: []any{
				map[string]any{
					"id":         float64(1),
					"name":       "build",
					"status":     "completed",
					"conclusion": "success",
					"steps":      []any{},
				},
			},
			wantLen: 1,
		},
		{
			name: "invalid job entry",
			jobsData: []any{
				"not a map",
				map[string]any{
					"id":         float64(2),
					"name":       "test",
					"status":     "completed",
					"conclusion": "success",
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summaries := parseJobSummaries(tt.jobsData)
			if len(summaries) != tt.wantLen {
				t.Errorf("parseJobSummaries() returned %d summaries, want %d", len(summaries), tt.wantLen)
			}
		})
	}
}

func TestParseSteps_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		stepsData []any
		wantLen   int
	}{
		{
			name:      "empty steps",
			stepsData: []any{},
			wantLen:   0,
		},
		{
			name:      "nil steps",
			stepsData: nil,
			wantLen:   0,
		},
		{
			name: "invalid step entry",
			stepsData: []any{
				"not a map",
				map[string]any{
					"name":       "Checkout",
					"number":     float64(1),
					"status":     "completed",
					"conclusion": "success",
				},
			},
			wantLen: 1,
		},
		{
			name: "step without conclusion",
			stepsData: []any{
				map[string]any{
					"name":   "Setup",
					"number": float64(1),
					"status": "in_progress",
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := parseSteps(tt.stepsData)
			if len(steps) != tt.wantLen {
				t.Errorf("parseSteps() returned %d steps, want %d", len(steps), tt.wantLen)
			}
		})
	}
}

func TestMonitorResultStruct(t *testing.T) {
	result := &MonitorResult{
		RunID:        123,
		Status:       "completed",
		Conclusion:   "success",
		WorkflowID:   "build.yml",
		WorkflowName: "Build",
		Branch:       "main",
		CommitSHA:    "abc123",
		Duration:     "5m0s",
		DurationSec:  300,
		StartedAt:    "2024-01-15T10:00:00Z",
		CompletedAt:  "2024-01-15T10:05:00Z",
		Jobs:         []JobSummary{},
		Logs:         map[string]JobLogs{},
		Error:        "",
		TimedOut:     false,
	}

	if result.RunID != 123 {
		t.Errorf("RunID = %d, want 123", result.RunID)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %s, want completed", result.Status)
	}
	if result.Conclusion != "success" {
		t.Errorf("Conclusion = %s, want success", result.Conclusion)
	}
	if result.TimedOut {
		t.Error("TimedOut should be false")
	}
}

func TestMonitorResultErrorState(t *testing.T) {
	result := &MonitorResult{
		RunID:       456,
		Status:      "completed",
		Conclusion:  "failure",
		WorkflowID:  "test.yml",
		Branch:      "develop",
		CommitSHA:   "def456",
		Duration:    "2m30s",
		DurationSec: 150,
		Error:       "Test failed",
		TimedOut:    false,
		Jobs: []JobSummary{
			{
				ID:         1,
				Name:       "test",
				Status:     "completed",
				Conclusion: "failure",
			},
		},
	}

	if result.Conclusion != "failure" {
		t.Errorf("Conclusion = %s, want failure", result.Conclusion)
	}
	if result.Error != "Test failed" {
		t.Errorf("Error = %s, want 'Test failed'", result.Error)
	}
	if len(result.Jobs) != 1 {
		t.Errorf("Jobs count = %d, want 1", len(result.Jobs))
	}
}

func TestMonitorResultTimeoutState(t *testing.T) {
	result := &MonitorResult{
		RunID:       789,
		Status:      "in_progress",
		Conclusion:  "",
		WorkflowID:  "deploy.yml",
		Branch:      "main",
		CommitSHA:   "ghi789",
		Duration:    "10m0s",
		DurationSec: 600,
		Error:       "Polling timed out after 10m0s",
		TimedOut:    true,
		Jobs:        []JobSummary{},
	}

	if !result.TimedOut {
		t.Error("TimedOut should be true")
	}
	if result.Error == "" {
		t.Error("Error should not be empty for timeout")
	}
}
