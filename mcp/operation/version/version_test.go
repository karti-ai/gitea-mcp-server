package version

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/flag"
	"github.com/mark3labs/mcp-go/mcp"
)

func Test_checkGiteaVersionFn_success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"1.22.5"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	origHost := flag.Host
	origToken := flag.Token
	origVersion := flag.Version
	flag.Host = server.URL
	flag.Token = ""
	flag.Version = "test"
	defer func() {
		flag.Host = origHost
		flag.Token = origToken
		flag.Version = origVersion
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := CheckGiteaVersionFn(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckGiteaVersionFn() error = %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatalf("expected content in result")
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	var parsed VersionInfo
	if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if parsed.Version != "1.22.5" {
		t.Errorf("version = %q, want %q", parsed.Version, "1.22.5")
	}
	if parsed.Major != 1 {
		t.Errorf("major = %d, want 1", parsed.Major)
	}
	if parsed.Minor != 22 {
		t.Errorf("minor = %d, want 22", parsed.Minor)
	}
	if parsed.Patch != 5 {
		t.Errorf("patch = %d, want 5", parsed.Patch)
	}

	// 1.22.5 should have secrets, variables, runners, commit_status but not actions/artifacts
	if !parsed.Capabilities.SecretsAPI {
		t.Error("expected secrets_api=true for 1.22.5")
	}
	if !parsed.Capabilities.VariablesAPI {
		t.Error("expected variables_api=true for 1.22.5")
	}
	if !parsed.Capabilities.RunnersAPI {
		t.Error("expected runners_api=true for 1.22.5")
	}
	if !parsed.Capabilities.CommitStatusAPI {
		t.Error("expected commit_status_api=true for 1.22.5")
	}
	if parsed.Capabilities.ActionsAPI {
		t.Error("expected actions_api=false for 1.22.5")
	}
	if parsed.Capabilities.ArtifactsAPI {
		t.Error("expected artifacts_api=false for 1.22.5")
	}
}

func Test_checkGiteaVersionFn_version123(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"1.23.0"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	origHost := flag.Host
	origToken := flag.Token
	origVersion := flag.Version
	flag.Host = server.URL
	flag.Token = ""
	flag.Version = "test"
	defer func() {
		flag.Host = origHost
		flag.Token = origToken
		flag.Version = origVersion
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := CheckGiteaVersionFn(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckGiteaVersionFn() error = %v", err)
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	var parsed VersionInfo
	if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if parsed.Version != "1.23.0" {
		t.Errorf("version = %q, want %q", parsed.Version, "1.23.0")
	}

	// 1.23.0 should have all capabilities
	if !parsed.Capabilities.ActionsAPI {
		t.Error("expected actions_api=true for 1.23.0")
	}
	if !parsed.Capabilities.ArtifactsAPI {
		t.Error("expected artifacts_api=true for 1.23.0")
	}
	if !parsed.Capabilities.SecretsAPI {
		t.Error("expected secrets_api=true for 1.23.0")
	}
	if !parsed.Capabilities.VariablesAPI {
		t.Error("expected variables_api=true for 1.23.0")
	}
	if !parsed.Capabilities.RunnersAPI {
		t.Error("expected runners_api=true for 1.23.0")
	}
	if !parsed.Capabilities.CommitStatusAPI {
		t.Error("expected commit_status_api=true for 1.23.0")
	}
}

func Test_checkGiteaVersionFn_version111(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"1.11.0"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	origHost := flag.Host
	origToken := flag.Token
	origVersion := flag.Version
	flag.Host = server.URL
	flag.Token = ""
	flag.Version = "test"
	defer func() {
		flag.Host = origHost
		flag.Token = origToken
		flag.Version = origVersion
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := CheckGiteaVersionFn(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckGiteaVersionFn() error = %v", err)
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	var parsed VersionInfo
	if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	// 1.11.0 should only have commit_status
	if !parsed.Capabilities.CommitStatusAPI {
		t.Error("expected commit_status_api=true for 1.11.0")
	}
	if parsed.Capabilities.SecretsAPI {
		t.Error("expected secrets_api=false for 1.11.0")
	}
	if parsed.Capabilities.VariablesAPI {
		t.Error("expected variables_api=false for 1.11.0")
	}
	if parsed.Capabilities.RunnersAPI {
		t.Error("expected runners_api=false for 1.11.0")
	}
	if parsed.Capabilities.ActionsAPI {
		t.Error("expected actions_api=false for 1.11.0")
	}
	if parsed.Capabilities.ArtifactsAPI {
		t.Error("expected artifacts_api=false for 1.11.0")
	}
}

func Test_checkGiteaVersionFn_withVPrefix(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"v1.24.0"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	origHost := flag.Host
	origToken := flag.Token
	origVersion := flag.Version
	flag.Host = server.URL
	flag.Token = ""
	flag.Version = "test"
	defer func() {
		flag.Host = origHost
		flag.Token = origToken
		flag.Version = origVersion
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := CheckGiteaVersionFn(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckGiteaVersionFn() error = %v", err)
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	var parsed VersionInfo
	if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if parsed.Version != "v1.24.0" {
		t.Errorf("version = %q, want %q", parsed.Version, "v1.24.0")
	}
	if parsed.Major != 1 {
		t.Errorf("major = %d, want 1", parsed.Major)
	}
	if parsed.Minor != 24 {
		t.Errorf("minor = %d, want 24", parsed.Minor)
	}
	if parsed.Patch != 0 {
		t.Errorf("patch = %d, want 0", parsed.Patch)
	}

	// 1.24 should have all capabilities
	if !parsed.Capabilities.ActionsAPI {
		t.Error("expected actions_api=true for 1.24.0")
	}
}

func Test_checkGiteaVersionFn_error(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Internal Server Error"}`, http.StatusInternalServerError)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	origHost := flag.Host
	origToken := flag.Token
	origVersion := flag.Version
	flag.Host = server.URL
	flag.Token = ""
	flag.Version = "test"
	defer func() {
		flag.Host = origHost
		flag.Token = origToken
		flag.Version = origVersion
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{},
		},
	}

	result, err := CheckGiteaVersionFn(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckGiteaVersionFn() error = %v", err)
	}

	textContent, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	// Should return error message in text result
	if textContent.Text == "" {
		t.Error("expected error message in result")
	}
}

func Test_parseVersion(t *testing.T) {
	tests := []struct {
		version      string
		wantMajor    int
		wantMinor    int
		wantPatch    int
		wantErr      bool
	}{
		{"1.22.5", 1, 22, 5, false},
		{"1.23.0", 1, 23, 0, false},
		{"1.24.1", 1, 24, 1, false},
		{"v1.22.5", 1, 22, 5, false},
		{"V1.22.5", 1, 22, 5, false},
		{"1.22", 1, 22, 0, false},
		{"1", 0, 0, 0, true},
		{"", 0, 0, 0, true},
		{"abc", 0, 0, 0, true},
		{"1.x.5", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major, minor, patch, err := parseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor {
					t.Errorf("parseVersion(%q) major = %d, want %d", tt.version, major, tt.wantMajor)
				}
				if minor != tt.wantMinor {
					t.Errorf("parseVersion(%q) minor = %d, want %d", tt.version, minor, tt.wantMinor)
				}
				if patch != tt.wantPatch {
					t.Errorf("parseVersion(%q) patch = %d, want %d", tt.version, patch, tt.wantPatch)
				}
			}
		})
	}
}

func Test_determineCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		major    int
		minor    int
		patch    int
		expected CapabilityMatrix
	}{
		{
			name:  "1.11.0",
			major: 1, minor: 11, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      false,
				SecretsAPI:     false,
				VariablesAPI:   false,
				RunnersAPI:     false,
				ArtifactsAPI:   false,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "1.12.0",
			major: 1, minor: 12, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      false,
				SecretsAPI:     false,
				VariablesAPI:   false,
				RunnersAPI:     false,
				ArtifactsAPI:   false,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "1.22.0",
			major: 1, minor: 22, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      false,
				SecretsAPI:     true,
				VariablesAPI:   true,
				RunnersAPI:     true,
				ArtifactsAPI:   false,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "1.22.5",
			major: 1, minor: 22, patch: 5,
			expected: CapabilityMatrix{
				ActionsAPI:      false,
				SecretsAPI:     true,
				VariablesAPI:   true,
				RunnersAPI:     true,
				ArtifactsAPI:   false,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "1.23.0",
			major: 1, minor: 23, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      true,
				SecretsAPI:     true,
				VariablesAPI:   true,
				RunnersAPI:     true,
				ArtifactsAPI:   true,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "1.24.0",
			major: 1, minor: 24, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      true,
				SecretsAPI:     true,
				VariablesAPI:   true,
				RunnersAPI:     true,
				ArtifactsAPI:   true,
				CommitStatusAPI: true,
			},
		},
		{
			name:  "2.0.0",
			major: 2, minor: 0, patch: 0,
			expected: CapabilityMatrix{
				ActionsAPI:      true,
				SecretsAPI:     true,
				VariablesAPI:   true,
				RunnersAPI:     true,
				ArtifactsAPI:   true,
				CommitStatusAPI: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineCapabilities(tt.major, tt.minor, tt.patch)
			if got != tt.expected {
				t.Errorf("determineCapabilities(%d, %d, %d) = %+v, want %+v",
					tt.major, tt.minor, tt.patch, got, tt.expected)
			}
		})
	}
}
