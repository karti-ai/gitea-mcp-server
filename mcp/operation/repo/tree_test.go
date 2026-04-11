package repo

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{
			name:    "exact match",
			path:    "README.md",
			pattern: "README.md",
			want:    true,
		},
		{
			name:    "wildcard match - all md files",
			path:    "docs/README.md",
			pattern: "*.md",
			want:    true,
		},
		{
			name:    "wildcard match - yaml files",
			path:    ".gitea/workflows/build.yml",
			pattern: "*.yml",
			want:    true,
		},
		{
			name:    "directory prefix match",
			path:    ".gitea/workflows/build.yml",
			pattern: ".gitea/*",
			want:    true,
		},
		{
			name:    "recursive directory match",
			path:    ".github/workflows/test.yml",
			pattern: ".github/**",
			want:    true,
		},
		{
			name:    "double star pattern",
			path:    "src/components/Button.tsx",
			pattern: "**/*.tsx",
			want:    true,
		},
		{
			name:    "double star with prefix",
			path:    "src/internal/utils/helpers.go",
			pattern: "src/**/*.go",
			want:    true,
		},
		{
			name:    "no match - wrong extension",
			path:    "main.go",
			pattern: "*.md",
			want:    false,
		},
		{
			name:    "no match - wrong directory",
			path:    "docs/readme.md",
			pattern: ".gitea/*",
			want:    false,
		},
		{
			name:    "negation pattern - exclude",
			path:    "node_modules/lodash/index.js",
			pattern: "!node_modules/**",
			want:    false,
		},
		{
			name:    "negation pattern - include others",
			path:    "src/main.js",
			pattern: "!node_modules/**",
			want:    true,
		},
		{
			name:    "empty pattern matches all",
			path:    "any/path/file.txt",
			pattern: "",
			want:    true,
		},
		{
			name:    "exact directory match",
			path:    "src/components",
			pattern: "src",
			want:    true,
		},
		{
			name:    "file inside directory",
			path:    "src/components/Button.tsx",
			pattern: "src",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFilterEntries(t *testing.T) {
	entries := []TreeEntry{
		{Path: "README.md", Type: "blob", SHA: "abc123"},
		{Path: "main.go", Type: "blob", SHA: "def456"},
		{Path: "docs", Type: "tree", SHA: "ghi789"},
		{Path: "docs/guide.md", Type: "blob", SHA: "jkl012"},
		{Path: ".gitea/workflows/build.yml", Type: "blob", SHA: "mno345"},
	}

	tests := []struct {
		name     string
		pattern  string
		expected int
	}{
		{
			name:     "no pattern returns all",
			pattern:  "",
			expected: 5,
		},
		{
			name:     "filter markdown files",
			pattern:  "*.md",
			expected: 2,
		},
		{
			name:     "filter yaml files",
			pattern:  "*.yml",
			expected: 1,
		},
		{
			name:     "filter by directory",
			pattern:  ".gitea/*",
			expected: 1,
		},
		{
			name:     "filter go files",
			pattern:  "*.go",
			expected: 1,
		},
		{
			name:     "no match returns empty",
			pattern:  "*.py",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterEntries(entries, tt.pattern)
			if len(filtered) != tt.expected {
				t.Errorf("filterEntries() returned %d entries, want %d", len(filtered), tt.expected)
			}
		})
	}
}

func TestSlimTreeEntries(t *testing.T) {
	entries := []TreeEntry{
		{Path: "README.md", Type: "blob", SHA: "abc123", Size: 1024},
		{Path: "docs", Type: "tree", SHA: "def456", Size: 0},
		{Path: "main.go", Type: "blob", SHA: "ghi789", Size: 2048},
	}

	slimmed := slimTreeEntries(entries)

	if len(slimmed) != len(entries) {
		t.Errorf("slimTreeEntries() returned %d entries, want %d", len(slimmed), len(entries))
	}

	for i, entry := range slimmed {
		if _, ok := entry["path"]; !ok {
			t.Errorf("entry %d missing 'path' field", i)
		}
		if _, ok := entry["type"]; !ok {
			t.Errorf("entry %d missing 'type' field", i)
		}
		if _, ok := entry["sha"]; !ok {
			t.Errorf("entry %d missing 'sha' field", i)
		}

		entryType := entry["type"](string)
		_, hasSize := entry["size"]

		if entryType == "blob" && entries[i].Size > 0 && !hasSize {
			t.Errorf("blob entry %d should have size field", i)
		}
		if entryType == "tree" && hasSize {
			t.Errorf("tree entry %d should not have size field", i)
		}
	}
}

func TestMatchPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{
			name:    "empty path with empty pattern",
			path:    "",
			pattern: "",
			want:    true,
		},
		{
			name:    "empty path with pattern",
			path:    "",
			pattern: "*.go",
			want:    false,
		},
		{
			name:    "path with spaces",
			path:    "path with spaces/file.txt",
			pattern: "*.txt",
			want:    true,
		},
		{
			name:    "special characters in path",
			path:    "path-with-dashes/file_name.txt",
			pattern: "*.txt",
			want:    true,
		},
		{
			name:    "double star at start and end",
			path:    "deep/nested/path/file.go",
			pattern: "**/*.go",
			want:    true,
		},
		{
			name:    "negation with double star",
			path:    "node_modules/deep/package.json",
			pattern: "!node_modules/**",
			want:    false,
		},
		{
			name:    "complex glob pattern",
			path:    "src/components/Button.test.tsx",
			pattern: "**/*.test.tsx",
			want:    true,
		},
		{
			name:    "directory only pattern",
			path:    "src/components",
			pattern: "src/*",
			want:    true,
		},
		{
			name:    "trailing slash in directory",
			path:    "src/components/",
			pattern: "src/*",
			want:    true,
		},
		{
			name:    "single character wildcard",
			path:    "file1.txt",
			pattern: "file?.txt",
			want:    true,
		},
		{
			name:    "range pattern",
			path:    "file5.txt",
			pattern: "file[0-9].txt",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFilterEntries_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		entries  []TreeEntry
		pattern  string
		expected int
	}{
		{
			name:     "nil entries",
			entries:  nil,
			pattern:  "*.go",
			expected: 0,
		},
		{
			name:     "empty entries",
			entries:  []TreeEntry{},
			pattern:  "*.go",
			expected: 0,
		},
		{
			name: "entries with empty paths",
			entries: []TreeEntry{
				{Path: "", Type: "blob", SHA: "abc123"},
				{Path: "main.go", Type: "blob", SHA: "def456"},
			},
			pattern:  "*.go",
			expected: 1,
		},
		{
			name: "negation pattern",
			entries: []TreeEntry{
				{Path: "test.go", Type: "blob", SHA: "abc123"},
				{Path: "vendor/lib.go", Type: "blob", SHA: "def456"},
			},
			pattern:  "!vendor/**",
			expected: 1,
		},
		{
			name: "complex pattern",
			entries: []TreeEntry{
				{Path: "src/main.go", Type: "blob", SHA: "abc123"},
				{Path: "src/test/main_test.go", Type: "blob", SHA: "def456"},
				{Path: "docs/readme.md", Type: "blob", SHA: "ghi789"},
			},
			pattern:  "src/**/*.go",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterEntries(tt.entries, tt.pattern)
			if len(filtered) != tt.expected {
				t.Errorf("filterEntries() returned %d entries, want %d", len(filtered), tt.expected)
			}
		})
	}
}

func TestSlimTreeEntries_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		entries []TreeEntry
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil entries",
			entries: nil,
			wantLen: 0,
		},
		{
			name:    "empty entries",
			entries: []TreeEntry{},
			wantLen: 0,
		},
		{
			name: "entry with zero size blob",
			entries: []TreeEntry{
				{Path: "empty.txt", Type: "blob", SHA: "abc", Size: 0},
			},
			wantLen: 1,
		},
		{
			name: "entry with symlink type",
			entries: []TreeEntry{
				{Path: "link", Type: "symlink", SHA: "def", Size: 0},
			},
			wantLen: 1,
		},
		{
			name: "many entries",
			entries: []TreeEntry{
				{Path: "file1.txt", Type: "blob", SHA: "a", Size: 100},
				{Path: "file2.txt", Type: "blob", SHA: "b", Size: 200},
				{Path: "file3.txt", Type: "blob", SHA: "c", Size: 300},
				{Path: "dir1", Type: "tree", SHA: "d", Size: 0},
				{Path: "dir2", Type: "tree", SHA: "e", Size: 0},
			},
			wantLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slimmed := slimTreeEntries(tt.entries)
			if len(slimmed) != tt.wantLen {
				t.Errorf("slimTreeEntries() returned %d entries, want %d", len(slimmed), tt.wantLen)
			}
		})
	}
}

func TestTreeEntryStruct(t *testing.T) {
	tests := []struct {
		name string
		entry TreeEntry
	}{
		{
			name: "minimal entry",
			entry: TreeEntry{
				Path: "file.txt",
				Type: "blob",
				SHA:  "abc123",
			},
		},
		{
			name: "full entry",
			entry: TreeEntry{
				Path: "file.txt",
				Type: "blob",
				SHA:  "def456789abc",
				Size: 1024,
			},
		},
		{
			name: "tree entry",
			entry: TreeEntry{
				Path: "directory",
				Type: "tree",
				SHA:  "ghi789",
				Size: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.entry.Path == "" {
				t.Error("Path should not be empty")
			}
			if tt.entry.Type == "" {
				t.Error("Type should not be empty")
			}
			if tt.entry.SHA == "" {
				t.Error("SHA should not be empty")
			}
		})
	}
}

func TestTreeResponseStruct(t *testing.T) {
	response := TreeResponse{
		SHA:       "abc123def456",
		URL:       "https://api.example.com/repos/owner/repo/git/trees/abc123",
		Tree:      []TreeEntry{},
		Truncated: false,
	}

	if response.SHA == "" {
		t.Error("SHA should not be empty")
	}
	if response.URL == "" {
		t.Error("URL should not be empty")
	}
	if response.Tree == nil {
		t.Error("Tree should not be nil")
	}
	if response.Truncated {
		t.Error("Truncated should be false")
	}
}
