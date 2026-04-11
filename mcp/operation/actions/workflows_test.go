package actions

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		want     bool
	}{
		{"build.yml", "*", true},
		{"build.yml", "*.*", true},
		{"build.yml", "*.yml", true},
		{"build.yaml", "*.yml", true},
		{"build.yaml", "*.yaml", true},
		{"build-test.yml", "build-*.yml", true},
		{"build-test.yml", "*-test.yml", true},
		{"build.yml", "deploy*.yml", false},
		{"build.yml", "*.yaml", false},
		{"BUILD.YML", "*.yml", true},
		{"build.yml", "build.yml", true},
		{"deploy.yml", "build.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"_"+tt.pattern, func(t *testing.T) {
			got := matchPattern(tt.filename, tt.pattern)
			if got != tt.want {
				t.Fatalf("matchPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestConvertYamlToInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int value",
			input:    42,
			expected: 42,
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"name": "test",
				"val":  123,
			},
			expected: map[string]interface{}{
				"name": "test",
				"val":  123,
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "value",
				},
			},
			expected: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "value",
				},
			},
		},
		{
			name:     "slice",
			input:    []interface{}{"a", "b", "c"},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name: "map with interface keys",
			input: map[interface{}]interface{}{
				"key": "value",
				123:   "numeric key",
			},
			expected: map[string]interface{}{
				"key": "value",
				"123": "numeric key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertYamlToInterface(tt.input)

			if !deepEqual(got, tt.expected) {
				t.Fatalf("convertYamlToInterface() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func deepEqual(a, b interface{}) bool {
	switch av := a.(type) {
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if !deepEqual(v, bv[k]) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i, v := range av {
			if !deepEqual(v, bv[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

func TestWorkflowFileStruct(t *testing.T) {
	wf := WorkflowFile{
		Name:       "test.yml",
		Path:       ".gitea/workflows/test.yml",
		SHA:        "abc123",
		Size:       1024,
		Content:    map[string]interface{}{"name": "Test Workflow"},
		RawContent: "name: Test Workflow",
		Encoding:   "base64",
	}

	if wf.Name != "test.yml" {
		t.Fatalf("Name = %q, want %q", wf.Name, "test.yml")
	}

	if wf.Path != ".gitea/workflows/test.yml" {
		t.Fatalf("Path = %q, want %q", wf.Path, ".gitea/workflows/test.yml")
	}
}

func TestMatchPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		filename string
		pattern  string
		want     bool
	}{
		// Edge cases for empty strings
		{"", "*", true},
		{"file.yml", "", false},
		{"", "", true},
		// Edge cases for special characters
		{"file-name.yml", "*-name.yml", true},
		{"file_name.yml", "*.yml", true},
		{"file.name.yml", "*.yml", true},
		// Edge case: only wildcard
		{"anything", "*", true},
		{"", "*", true},
		// Edge case: pattern equals filename
		{"exact.yml", "exact.yml", true},
		{"exact.yml", "exact.yaml", false},
		// Edge case: case sensitivity
		{"FILE.YML", "*.yml", true},
		{"File.Yml", "*.yml", true},
		// Edge case: middle wildcards
		{"build-test-deploy.yml", "*test*.yml", true},
		{"build-prod-deploy.yml", "*test*.yml", false},
		// Edge case: multiple extensions
		{"file.tar.gz", "*.gz", true},
		{"file.tar.gz", "*.tar.gz", true},
		// Edge case: dots in filename
		{".github/workflows/build.yml", "*.yml", true},
		// Edge case: numeric patterns
		{"build-123.yml", "build-*.yml", true},
		{"build-abc.yml", "build-*.yml", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename+"_"+tt.pattern, func(t *testing.T) {
			got := matchPattern(tt.filename, tt.pattern)
			if got != tt.want {
				t.Fatalf("matchPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestConvertYamlToInterface_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: []interface{}{},
		},
		{
			name: "nested empty structures",
			input: map[string]interface{}{
				"empty_map":   map[string]interface{}{},
				"empty_slice": []interface{}{},
			},
			expected: map[string]interface{}{
				"empty_map":   map[string]interface{}{},
				"empty_slice": []interface{}{},
			},
		},
		{
			name:     "boolean value",
			input:    true,
			expected: true,
		},
		{
			name:     "float value",
			input:    3.14,
			expected: 3.14,
		},
		{
			name: "deeply nested map",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"value": "deep",
						},
					},
				},
			},
			expected: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"value": "deep",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertYamlToInterface(tt.input)

			if !deepEqual(got, tt.expected) {
				t.Fatalf("convertYamlToInterface() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWorkflowFileStruct_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		file WorkflowFile
		want WorkflowFile
	}{
		{
			name: "empty file",
			file: WorkflowFile{},
			want: WorkflowFile{},
		},
		{
			name: "file with zero size",
			file: WorkflowFile{
				Name:       "empty.yml",
				Path:       ".gitea/workflows/empty.yml",
				SHA:        "abc123",
				Size:       0,
				Content:    nil,
				RawContent: "",
				Encoding:   "",
			},
			want: WorkflowFile{
				Name: "empty.yml",
				Path: ".gitea/workflows/empty.yml",
				SHA:  "abc123",
				Size: 0,
			},
		},
		{
			name: "file with large size",
			file: WorkflowFile{
				Name:       "large.yml",
				Path:       ".gitea/workflows/large.yml",
				SHA:        "def456",
				Size:       1024 * 1024 * 10, // 10MB
				Content:    map[string]interface{}{"name": "Large Workflow"},
				RawContent: "name: Large Workflow\\n# ... lots of content ...",
				Encoding:   "base64",
			},
			want: WorkflowFile{
				Name: "large.yml",
				Path: ".gitea/workflows/large.yml",
				SHA:  "def456",
				Size: 1024 * 1024 * 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.file.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", tt.file.Name, tt.want.Name)
			}
			if tt.file.Path != tt.want.Path {
				t.Errorf("Path = %q, want %q", tt.file.Path, tt.want.Path)
			}
			if tt.file.SHA != tt.want.SHA {
				t.Errorf("SHA = %q, want %q", tt.file.SHA, tt.want.SHA)
			}
			if tt.file.Size != tt.want.Size {
				t.Errorf("Size = %d, want %d", tt.file.Size, tt.want.Size)
			}
		})
	}
}

func TestWorkflowFilesResultStruct_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		result WorkflowFilesResult
	}{
		{
			name: "empty result",
			result: WorkflowFilesResult{
				Directory:  "",
				Files:      []WorkflowFile{},
				TotalCount: 0,
			},
		},
		{
			name: "nil files",
			result: WorkflowFilesResult{
				Directory:  ".gitea/workflows",
				Files:      nil,
				TotalCount: 0,
			},
		},
		{
			name: "single file",
			result: WorkflowFilesResult{
				Directory: ".github/workflows",
				Files: []WorkflowFile{
					{Name: "ci.yml"},
				},
				TotalCount: 1,
			},
		},
		{
			name: "many files",
			result: WorkflowFilesResult{
				Directory: ".gitea/workflows",
				Files: []WorkflowFile{
					{Name: "build.yml"},
					{Name: "test.yml"},
					{Name: "deploy.yml"},
					{Name: "lint.yml"},
					{Name: "security.yml"},
				},
				TotalCount: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.result.Files) != tt.result.TotalCount {
				t.Errorf("Files length (%d) != TotalCount (%d)", len(tt.result.Files), tt.result.TotalCount)
			}
		})
	}
}
