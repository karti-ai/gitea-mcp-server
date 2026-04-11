package actions

import (
	"testing"
)

func TestSlimActionArtifact(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "complete artifact",
			input: map[string]any{
				"id":            float64(123),
				"name":          "build-output",
				"size_in_bytes": float64(1024000),
				"download_url":  "https://gitea.example.com/api/v1/repos/owner/repo/actions/artifacts/123/download",
				"run_id":        float64(456),
				"created_at":    "2024-01-15T10:30:00Z",
				"expires_at":    "2024-02-15T10:30:00Z",
				"extra_field":   "should be removed",
			},
			expected: map[string]any{
				"id":            float64(123),
				"name":          "build-output",
				"size_in_bytes": float64(1024000),
				"download_url":  "https://gitea.example.com/api/v1/repos/owner/repo/actions/artifacts/123/download",
				"run_id":        float64(456),
				"created_at":    "2024-01-15T10:30:00Z",
				"expires_at":    "2024-02-15T10:30:00Z",
			},
		},
		{
			name: "minimal artifact",
			input: map[string]any{
				"id":   float64(789),
				"name": "test-results",
			},
			expected: map[string]any{
				"id":   float64(789),
				"name": "test-results",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifact(tt.input)
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("slimActionArtifact() did not return a map")
			}

			for key, expectedValue := range tt.expected {
				if resultMap[key] != expectedValue {
					t.Fatalf("slimActionArtifact()[%q] = %v, want %v", key, resultMap[key], expectedValue)
				}
			}

			if _, exists := resultMap["extra_field"]; exists {
				t.Fatalf("slimActionArtifact() should not include 'extra_field'")
			}
		})
	}
}

func TestSlimActionArtifactNonMap(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name:  "string input",
			input: "not a map",
		},
		{
			name:  "int input",
			input: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifact(tt.input)
			if result != tt.input {
				t.Fatalf("slimActionArtifact() = %v, want %v", result, tt.input)
			}
		})
	}
}

func TestSlimActionArtifacts(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "artifacts with total count",
			input: map[string]any{
				"total_count": float64(2),
				"artifacts": []any{
					map[string]any{
						"id":            float64(1),
						"name":          "artifact-1",
						"size_in_bytes": float64(1000),
						"download_url":  "url1",
					},
					map[string]any{
						"id":            float64(2),
						"name":          "artifact-2",
						"size_in_bytes": float64(2000),
						"download_url":  "url2",
					},
				},
			},
			expected: map[string]any{
				"total_count": float64(2),
			},
		},
		{
			name: "empty artifacts list",
			input: map[string]any{
				"total_count": float64(0),
				"artifacts":   []any{},
			},
			expected: map[string]any{
				"total_count": float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifacts(tt.input)
			resultMap, ok := result.(map[string]any)
			if !ok {
				t.Fatalf("slimActionArtifacts() did not return a map")
			}

			if resultMap["total_count"] != tt.expected["total_count"] {
				t.Fatalf("total_count = %v, want %v", resultMap["total_count"], tt.expected["total_count"])
			}

			if artifacts, ok := resultMap["artifacts"].([]any); ok {
				for _, item := range artifacts {
					if artifact, ok := item.(map[string]any); ok {
						if _, exists := artifact["id"]; !exists {
							t.Fatalf("slimmed artifact should have 'id' field")
						}
					}
				}
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1536 * 1024, "1.50 MB"},
		{100 * 1024 * 1024, "100.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1536 * 1024 * 1024, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Fatalf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{-1, "-1 B"},
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1025, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024*1024 + 1, "1.00 MB"},
		{1536 * 1024 * 1024, "1.50 GB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1024.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Fatalf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestSlimActionArtifact_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input any
		check func(t *testing.T, result any)
	}{
		{
			name:  "nil input",
			input: nil,
			check: func(t *testing.T, result any) {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
			},
		},
		{
			name:  "string input",
			input: "not a map",
			check: func(t *testing.T, result any) {
				if result != "not a map" {
					t.Fatalf("expected 'not a map', got %v", result)
				}
			},
		},
		{
			name:  "int input",
			input: 123,
			check: func(t *testing.T, result any) {
				if result != 123 {
					t.Fatalf("expected 123, got %v", result)
				}
			},
		},
		{
			name: "empty map",
			input: map[string]any{},
			check: func(t *testing.T, result any) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if len(resultMap) != 0 {
					t.Fatalf("expected empty map, got %d fields", len(resultMap))
				}
			},
		},
		{
			name: "map with only extra fields",
			input: map[string]any{
				"extra1": "value1",
				"extra2": "value2",
			},
			check: func(t *testing.T, result any) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if len(resultMap) != 0 {
					t.Fatalf("expected empty map after filtering, got %d fields", len(resultMap))
				}
			},
		},
		{
			name: "map with null values",
			input: map[string]any{
				"id":            float64(123),
				"name":          nil,
				"size_in_bytes": float64(1000),
			},
			check: func(t *testing.T, result any) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if resultMap["name"] != nil {
					t.Fatalf("expected nil name, got %v", resultMap["name"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifact(tt.input)
			tt.check(t, result)
		})
	}
}

func TestSlimActionArtifacts_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input any
		check func(t *testing.T, result any)
	}{
		{
			name:  "nil input",
			input: nil,
			check: func(t *testing.T, result any) {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
			},
		},
		{
			name:  "string input",
			input: "not a map",
			check: func(t *testing.T, result any) {
				if result != "not a map" {
					t.Fatalf("expected 'not a map', got %v", result)
				}
			},
		},
		{
			name: "map without artifacts key",
			input: map[string]any{
				"total_count": float64(0),
			},
			check: func(t *testing.T, result any) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if resultMap["total_count"] != float64(0) {
					t.Fatalf("expected total_count 0, got %v", resultMap["total_count"])
				}
			},
		},
		{
			name: "map with nil artifacts",
			input: map[string]any{
				"total_count": float64(0),
				"artifacts":   nil,
			},
			check: func(t *testing.T, result any) {
				resultMap, ok := result.(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if resultMap["total_count"] != float64(0) {
					t.Fatalf("expected total_count 0, got %v", resultMap["total_count"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifacts(tt.input)
			tt.check(t, result)
		})
	}
}

func TestDefaultMaxArtifactSize_Constant(t *testing.T) {
	expectedSize := int64(100 * 1024 * 1024)
	if DefaultMaxArtifactSize != expectedSize {
		t.Fatalf("DefaultMaxArtifactSize = %d, want %d", DefaultMaxArtifactSize, expectedSize)
	}
}

func TestActionsArtifactToolName_Constant(t *testing.T) {
	expectedName := "list_action_artifacts"
	if ActionsArtifactToolName != expectedName {
		t.Fatalf("ActionsArtifactToolName = %q, want %q", ActionsArtifactToolName, expectedName)
	}
}

func TestSlimActionArtifactsNonMap(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name:  "string input",
			input: "not a map",
		},
		{
			name:  "int input",
			input: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slimActionArtifacts(tt.input)
			if result != tt.input {
				t.Fatalf("slimActionArtifacts() = %v, want %v", result, tt.input)
			}
		})
	}
}
