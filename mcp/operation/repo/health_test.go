package repo

import (
	"encoding/json"
	"testing"
)

func TestCalculateHealthScore(t *testing.T) {
	tests := []struct {
		name     string
		result   *HealthResult
		expected int
	}{
		{
			name: "perfect health - active repo",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   5,
				},
				Issues: &IssuesInfo{
					Available: true,
					OpenCount: 5,
				},
				PullRequests: &PullRequestsInfo{
					Available: true,
					OpenCount: 2,
				},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "success",
					HasRecentRuns:     true,
				},
				BranchProtection: &BranchProtectionInfo{
					Available:              true,
					ProtectedBranchesCount: 1,
				},
				RepositoryInfo: &RepositoryInfo{
					IsArchived: false,
				},
			},
			expected: 100,
		},
		{
			name: "stale commits - 35 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   35,
				},
				Issues: &IssuesInfo{
					Available: true,
					OpenCount: 5,
				},
			},
			expected: 85,
		},
		{
			name: "very stale commits - 100 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   100,
				},
				Issues: &IssuesInfo{
					Available: true,
					OpenCount: 5,
				},
			},
			expected: 70,
		},
		{
			name: "too many open issues",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   5,
				},
				Issues: &IssuesInfo{
					Available: true,
					OpenCount: 60,
				},
			},
			expected: 90,
		},
		{
			name: "workflow failure",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   5,
				},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "failure",
					HasRecentRuns:     true,
				},
			},
			expected: 85,
		},
		{
			name: "archived repository",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   5,
				},
				RepositoryInfo: &RepositoryInfo{
					IsArchived: true,
				},
			},
			expected: 60,
		},
		{
			name:     "empty result - no data",
			result:   &HealthResult{},
			expected: 100,
		},
		{
			name: "boundary - minimum score",
			result: &HealthResult{
				LastCommit: &CommitInfo{
					Available: true,
					DaysAgo:   1000,
				},
				RepositoryInfo: &RepositoryInfo{
					IsArchived: true,
				},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "failure",
					HasRecentRuns:     false,
				},
				Issues: &IssuesInfo{
					Available: true,
					OpenCount: 100,
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateHealthScore(tt.result)
			if score != tt.expected {
				t.Errorf("calculateHealthScore() = %d, want %d", score, tt.expected)
			}
		})
	}
}

func TestGetHealthStatus(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{95, "excellent"},
		{90, "excellent"},
		{85, "good"},
		{70, "good"},
		{60, "fair"},
		{50, "fair"},
		{40, "poor"},
		{30, "poor"},
		{20, "critical"},
		{0, "critical"},
		{100, "excellent"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			status := getHealthStatus(tt.score)
			if status != tt.expected {
				t.Errorf("getHealthStatus(%d) = %s, want %s", tt.score, status, tt.expected)
			}
		})
	}
}

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{
			name:     "string value",
			m:        map[string]any{"status": "success"},
			key:      "status",
			expected: "success",
		},
		{
			name:     "missing key",
			m:        map[string]any{"other": "value"},
			key:      "status",
			expected: "",
		},
		{
			name:     "non-string value",
			m:        map[string]any{"count": 42},
			key:      "count",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]any{},
			key:      "status",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringFromMap(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getStringFromMap() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHealthResultJSONMarshaling(t *testing.T) {
	result := &HealthResult{
		Repository:   "owner/repo",
		HealthScore:  85,
		HealthStatus: "good",
		LastCommit: &CommitInfo{
			SHA:       "abc123",
			Message:   "Initial commit",
			Author:    "user",
			Date:      "2024-01-15T10:30:00Z",
			DaysAgo:   5,
			Available: true,
		},
		Issues: &IssuesInfo{
			OpenCount:  10,
			TotalCount: 50,
			Available:  true,
		},
		PullRequests: &PullRequestsInfo{
			OpenCount:  3,
			TotalCount: 15,
			Available:  true,
		},
		WorkflowStatus: &WorkflowStatusInfo{
			LastRunStatus:     "completed",
			LastRunConclusion: "success",
			HasRecentRuns:     true,
			Available:         true,
		},
		BranchProtection: &BranchProtectionInfo{
			ProtectedBranchesCount: 2,
			ProtectedBranches:      []string{"main", "develop"},
			Available:              true,
		},
		RepositoryInfo: &RepositoryInfo{
			Stars:      100,
			Forks:      20,
			Language:   "Go",
			IsPrivate:  false,
			IsArchived: false,
			Available:  true,
		},
		CheckedAt:     "2024-01-20T10:00:00Z",
		PartialResult: false,
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal HealthResult: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	var unmarshaled HealthResult
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal HealthResult: %v", err)
	}

	if unmarshaled.HealthScore != result.HealthScore {
		t.Errorf("HealthScore mismatch: got %d, want %d", unmarshaled.HealthScore, result.HealthScore)
	}

	if unmarshaled.HealthStatus != result.HealthStatus {
		t.Errorf("HealthStatus mismatch: got %s, want %s", unmarshaled.HealthStatus, result.HealthStatus)
	}

	if unmarshaled.LastCommit == nil || unmarshaled.LastCommit.SHA != result.LastCommit.SHA {
		t.Error("LastCommit mismatch")
	}
}

func TestCalculateHealthScore_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		result   *HealthResult
		expected int
	}{
		{
			name: "all nil fields",
			result: &HealthResult{
				LastCommit:       nil,
				Issues:           nil,
				PullRequests:     nil,
				WorkflowStatus:   nil,
				BranchProtection: nil,
				RepositoryInfo:   nil,
			},
			expected: 100,
		},
		{
			name: "unavailable fields",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: false},
				Issues:     &IssuesInfo{Available: false},
			},
			expected: 100,
		},
		{
			name: "stale commits boundary - exactly 30 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 30},
			},
			expected: 100,
		},
		{
			name: "stale commits boundary - exactly 31 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 31},
			},
			expected: 85,
		},
		{
			name: "stale commits boundary - exactly 90 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 90},
			},
			expected: 85,
		},
		{
			name: "stale commits boundary - exactly 91 days",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 91},
			},
			expected: 70,
		},
		{
			name: "issues boundary - exactly 20",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				Issues:     &IssuesInfo{Available: true, OpenCount: 20},
			},
			expected: 100,
		},
		{
			name: "issues boundary - exactly 21",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				Issues:     &IssuesInfo{Available: true, OpenCount: 21},
			},
			expected: 95,
		},
		{
			name: "issues boundary - exactly 50",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				Issues:     &IssuesInfo{Available: true, OpenCount: 50},
			},
			expected: 95,
		},
		{
			name: "issues boundary - exactly 51",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				Issues:     &IssuesInfo{Available: true, OpenCount: 51},
			},
			expected: 90,
		},
		{
			name: "PRs boundary - exactly 10",
			result: &HealthResult{
				LastCommit:   &CommitInfo{Available: true, DaysAgo: 5},
				PullRequests: &PullRequestsInfo{Available: true, OpenCount: 10},
			},
			expected: 100,
		},
		{
			name: "PRs boundary - exactly 11",
			result: &HealthResult{
				LastCommit:   &CommitInfo{Available: true, DaysAgo: 5},
				PullRequests: &PullRequestsInfo{Available: true, OpenCount: 11},
			},
			expected: 95,
		},
		{
			name: "workflow cancelled",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "cancelled",
					HasRecentRuns:     true,
				},
			},
			expected: 95,
		},
		{
			name: "workflow no recent runs",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 5},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "success",
					HasRecentRuns:     false,
				},
			},
			expected: 95,
		},
		{
			name: "archived with negative score",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 1000},
				RepositoryInfo: &RepositoryInfo{
					IsArchived: true,
				},
				WorkflowStatus: &WorkflowStatusInfo{
					Available:         true,
					LastRunConclusion: "failure",
					HasRecentRuns:     false,
				},
			},
			expected: 0,
		},
		{
			name: "maximum score cap",
			result: &HealthResult{
				LastCommit: &CommitInfo{Available: true, DaysAgo: 0},
				Issues:     &IssuesInfo{Available: true, OpenCount: 0},
				BranchProtection: &BranchProtectionInfo{
					Available:              true,
					ProtectedBranchesCount: 10,
				},
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateHealthScore(tt.result)
			if score != tt.expected {
				t.Errorf("calculateHealthScore() = %d, want %d", score, tt.expected)
			}
		})
	}
}

func TestGetHealthStatus_Boundaries(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{100, "excellent"},
		{91, "excellent"},
		{89, "good"},
		{71, "good"},
		{69, "fair"},
		{51, "fair"},
		{49, "poor"},
		{31, "poor"},
		{29, "critical"},
		{1, "critical"},
		{-10, "critical"},
		{110, "excellent"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%d", tt.score), func(t *testing.T) {
			status := getHealthStatus(tt.score)
			if status != tt.expected {
				t.Errorf("getHealthStatus(%d) = %s, want %s", tt.score, status, tt.expected)
			}
		})
	}
}

func TestGetStringFromMap_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected string
	}{
		{
			name:     "nil map",
			m:        nil,
			key:      "status",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]any{},
			key:      "status",
			expected: "",
		},
		{
			name:     "int value",
			m:        map[string]any{"count": int(42)},
			key:      "count",
			expected: "",
		},
		{
			name:     "float64 value",
			m:        map[string]any{"count": float64(42)},
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
			name:     "nested map value",
			m:        map[string]any{"data": map[string]any{"key": "value"}},
			key:      "data",
			expected: "",
		},
		{
			name:     "slice value",
			m:        map[string]any{"items": []string{"a", "b"}},
			key:      "items",
			expected: "",
		},
		{
			name:     "empty string value",
			m:        map[string]any{"name": ""},
			key:      "name",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringFromMap(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("getStringFromMap() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestHealthCheckStructs(t *testing.T) {
	t.Run("CommitInfo", func(t *testing.T) {
		commit := &CommitInfo{
			SHA:       "abc123",
			Message:   "Initial commit",
			Author:    "user@example.com",
			Date:      "2024-01-15T10:30:00Z",
			DaysAgo:   5,
			Available: true,
		}
		if commit.SHA != "abc123" {
			t.Error("SHA mismatch")
		}
		if commit.DaysAgo != 5 {
			t.Error("DaysAgo mismatch")
		}
	})

	t.Run("IssuesInfo", func(t *testing.T) {
		issues := &IssuesInfo{
			OpenCount:  10,
			TotalCount: 50,
			Available:  true,
		}
		if issues.OpenCount != 10 {
			t.Error("OpenCount mismatch")
		}
		if !issues.Available {
			t.Error("Available should be true")
		}
	})

	t.Run("PullRequestsInfo", func(t *testing.T) {
		prs := &PullRequestsInfo{
			OpenCount:  3,
			TotalCount: 15,
			Available:  true,
		}
		if prs.OpenCount != 3 {
			t.Error("OpenCount mismatch")
		}
	})

	t.Run("WorkflowStatusInfo", func(t *testing.T) {
		wf := &WorkflowStatusInfo{
			LastRunStatus:     "completed",
			LastRunConclusion: "success",
			HasRecentRuns:     true,
			Available:         true,
			Error:             "",
		}
		if wf.LastRunConclusion != "success" {
			t.Error("LastRunConclusion mismatch")
		}
	})

	t.Run("BranchProtectionInfo", func(t *testing.T) {
		bp := &BranchProtectionInfo{
			ProtectedBranchesCount: 2,
			ProtectedBranches:      []string{"main", "develop"},
			Available:              true,
		}
		if bp.ProtectedBranchesCount != 2 {
			t.Error("ProtectedBranchesCount mismatch")
		}
		if len(bp.ProtectedBranches) != 2 {
			t.Error("ProtectedBranches length mismatch")
		}
	})

	t.Run("RepositoryInfo", func(t *testing.T) {
		repo := &RepositoryInfo{
			Stars:      100,
			Forks:      20,
			Language:   "Go",
			IsPrivate:  false,
			IsArchived: false,
			Available:  true,
		}
		if repo.Stars != 100 {
			t.Error("Stars mismatch")
		}
		if repo.IsArchived {
			t.Error("IsArchived should be false")
		}
	})

	t.Run("HealthCheckError", func(t *testing.T) {
		err := HealthCheckError{
			Check: "workflow_status",
			Error: "API not available",
		}
		if err.Check != "workflow_status" {
			t.Error("Check mismatch")
		}
	})
}

func TestHealthResultErrors(t *testing.T) {
	result := &HealthResult{
		Repository:   "owner/repo",
		HealthScore:  75,
		HealthStatus: "good",
		Errors: []HealthCheckError{
			{Check: "workflow_status", Error: "API not available"},
			{Check: "branch_protection", Error: "No permissions"},
		},
		PartialResult: true,
		CheckedAt:     "2024-01-20T10:00:00Z",
	}

	if !result.PartialResult {
		t.Error("PartialResult should be true when errors exist")
	}
	if len(result.Errors) != 2 {
		t.Errorf("Errors count = %d, want 2", len(result.Errors))
	}
}

func TestHealthResultWithNilFields(t *testing.T) {
	result := &HealthResult{
		Repository:    "owner/repo",
		HealthScore:   100,
		HealthStatus:  "excellent",
		LastCommit:    nil,
		Issues:        nil,
		PullRequests:  nil,
		WorkflowStatus: nil,
		BranchProtection: nil,
		RepositoryInfo: nil,
		Errors:        []HealthCheckError{},
		CheckedAt:     "2024-01-20T10:00:00Z",
		PartialResult: false,
	}

	score := calculateHealthScore(result)
	if score != 100 {
		t.Errorf("calculateHealthScore() with nil fields = %d, want 100", score)
	}
}
