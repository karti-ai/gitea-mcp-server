package pull

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"gitea.com/gitea/gitea-mcp/pkg/flag"
	"github.com/mark3labs/mcp-go/mcp"
)

func Test_editPullRequestFn(t *testing.T) {
	const (
		owner = "octo"
		repo  = "demo"
		index = 7
	)

	indexInputs := []struct {
		name string
		val  any
	}{
		{"float64", float64(index)},
		{"string", "7"},
	}

	for _, ii := range indexInputs {
		t.Run(ii.name, func(t *testing.T) {
			var (
				mu        sync.Mutex
				gotMethod string
				gotPath   string
				gotBody   map[string]any
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/version":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"version":"1.12.0"}`))
				case fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo):
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"private":false}`))
				case fmt.Sprintf("/api/v1/repos/%s/%s/pulls/%d", owner, repo, index):
					mu.Lock()
					gotMethod = r.Method
					gotPath = r.URL.Path
					var body map[string]any
					_ = json.NewDecoder(r.Body).Decode(&body)
					gotBody = body
					mu.Unlock()
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write(fmt.Appendf(nil, `{"number":%d,"title":"%s","state":"open"}`, index, body["title"]))
				default:
					http.NotFound(w, r)
				}
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
					Arguments: map[string]any{
						"owner": owner,
						"repo":  repo,
						"index": ii.val,
						"title": "WIP: my feature",
						"state": "open",
					},
				},
			}

			result, err := editPullRequestFn(context.Background(), req)
			if err != nil {
				t.Fatalf("editPullRequestFn() error = %v", err)
			}

			mu.Lock()
			defer mu.Unlock()

			if gotMethod != http.MethodPatch {
				t.Fatalf("expected PATCH request, got %s", gotMethod)
			}
			if gotPath != fmt.Sprintf("/api/v1/repos/%s/%s/pulls/%d", owner, repo, index) {
				t.Fatalf("unexpected path: %s", gotPath)
			}
			if gotBody["title"] != "WIP: my feature" {
				t.Fatalf("expected title 'WIP: my feature', got %v", gotBody["title"])
			}
			if gotBody["state"] != "open" {
				t.Fatalf("expected state 'open', got %v", gotBody["state"])
			}

			if len(result.Content) == 0 {
				t.Fatalf("expected content in result")
			}
			textContent, ok := mcp.AsTextContent(result.Content[0])
			if !ok {
				t.Fatalf("expected text content, got %T", result.Content[0])
			}

			var parsed map[string]any
			if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
				t.Fatalf("unmarshal result text: %v", err)
			}
			if got := parsed["title"].(string); got != "WIP: my feature" {
				t.Fatalf("result title = %q, want %q", got, "WIP: my feature")
			}
		})
	}
}

func Test_mergePullRequestFn(t *testing.T) {
	const (
		owner = "octo"
		repo  = "demo"
		index = 5
	)

	indexInputs := []struct {
		name string
		val  any
	}{
		{"float64", float64(index)},
		{"string", "5"},
	}

	for _, ii := range indexInputs {
		t.Run(ii.name, func(t *testing.T) {
			var (
				mu        sync.Mutex
				gotMethod string
				gotPath   string
				gotBody   map[string]any
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/version":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"version":"1.12.0"}`))
				case fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo):
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"private":false}`))
				case fmt.Sprintf("/api/v1/repos/%s/%s/pulls/%d/merge", owner, repo, index):
					mu.Lock()
					gotMethod = r.Method
					gotPath = r.URL.Path
					var body map[string]any
					_ = json.NewDecoder(r.Body).Decode(&body)
					gotBody = body
					mu.Unlock()
					w.WriteHeader(http.StatusOK)
				default:
					http.NotFound(w, r)
				}
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
					Arguments: map[string]any{
						"owner":         owner,
						"repo":          repo,
						"index":         ii.val,
						"merge_style":   "squash",
						"title":         "feat: my squashed commit",
						"message":       "Squash merge of PR #5",
						"delete_branch": true,
					},
				},
			}

			result, err := mergePullRequestFn(context.Background(), req)
			if err != nil {
				t.Fatalf("mergePullRequestFn() error = %v", err)
			}

			mu.Lock()
			defer mu.Unlock()

			if gotMethod != http.MethodPost {
				t.Fatalf("expected POST request, got %s", gotMethod)
			}
			if gotPath != fmt.Sprintf("/api/v1/repos/%s/%s/pulls/%d/merge", owner, repo, index) {
				t.Fatalf("unexpected path: %s", gotPath)
			}
			if gotBody["Do"] != "squash" {
				t.Fatalf("expected Do 'squash', got %v", gotBody["Do"])
			}
			if gotBody["MergeTitleField"] != "feat: my squashed commit" {
				t.Fatalf("expected MergeTitleField 'feat: my squashed commit', got %v", gotBody["MergeTitleField"])
			}
			if gotBody["MergeMessageField"] != "Squash merge of PR #5" {
				t.Fatalf("expected MergeMessageField 'Squash merge of PR #5', got %v", gotBody["MergeMessageField"])
			}
			if gotBody["delete_branch_after_merge"] != true {
				t.Fatalf("expected delete_branch_after_merge true, got %v", gotBody["delete_branch_after_merge"])
			}

			if len(result.Content) == 0 {
				t.Fatalf("expected content in result")
			}
			textContent, ok := mcp.AsTextContent(result.Content[0])
			if !ok {
				t.Fatalf("expected text content, got %T", result.Content[0])
			}

			var parsed map[string]any
			if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
				t.Fatalf("unmarshal result text: %v", err)
			}
			if parsed["merged"] != true {
				t.Fatalf("expected merged=true, got %v", parsed["merged"])
			}
			if parsed["merge_style"] != "squash" {
				t.Fatalf("expected merge_style 'squash', got %v", parsed["merge_style"])
			}
			if parsed["branch_deleted"] != true {
				t.Fatalf("expected branch_deleted=true, got %v", parsed["branch_deleted"])
			}
		})
	}
}

func Test_getPullRequestDiffFn(t *testing.T) {
	const (
		owner   = "octo"
		repo    = "demo"
		index   = 12
		diffRaw = "diff --git a/file.txt b/file.txt\n+line\n"
	)

	indexInputs := []struct {
		name string
		val  any
	}{
		{"float64", float64(index)},
		{"string", "12"},
	}

	for _, ii := range indexInputs {
		t.Run(ii.name, func(t *testing.T) {
			var (
				mu            sync.Mutex
				diffRequested bool
				binaryValue   string
			)
			errCh := make(chan error, 1)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/version":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"version":"1.12.0"}`))
				case fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo):
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"private":false}`))
				case fmt.Sprintf("/%s/%s/pulls/%d.diff", owner, repo, index):
					if r.Method != http.MethodGet {
						select {
						case errCh <- fmt.Errorf("unexpected method: %s", r.Method):
						default:
						}
					}
					mu.Lock()
					diffRequested = true
					binaryValue = r.URL.Query().Get("binary")
					mu.Unlock()
					w.Header().Set("Content-Type", "text/plain")
					_, _ = w.Write([]byte(diffRaw))
				default:
					select {
					case errCh <- fmt.Errorf("unexpected request path: %s", r.URL.Path):
					default:
					}
				}
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
					Arguments: map[string]any{
						"owner":  owner,
						"repo":   repo,
						"index":  ii.val,
						"binary": true,
					},
				},
			}

			result, err := getPullRequestDiffFn(context.Background(), req)
			if err != nil {
				t.Fatalf("getPullRequestDiffFn() error = %v", err)
			}

			select {
			case reqErr := <-errCh:
				t.Fatalf("handler error: %v", reqErr)
			default:
			}

			mu.Lock()
			requested := diffRequested
			gotBinary := binaryValue
			mu.Unlock()

			if !requested {
				t.Fatalf("expected diff request to be made")
			}
			if gotBinary != "true" {
				t.Fatalf("expected binary=true query param, got %q", gotBinary)
			}

			if len(result.Content) == 0 {
				t.Fatalf("expected content in result")
			}

			textContent, ok := mcp.AsTextContent(result.Content[0])
			if !ok {
				t.Fatalf("expected text content, got %T", result.Content[0])
			}

			// The diff response is now a plain string
			var parsed string
			if err := json.Unmarshal([]byte(textContent.Text), &parsed); err != nil {
				t.Fatalf("unmarshal result text: %v", err)
			}
			if parsed != diffRaw {
				t.Fatalf("diff = %q, want %q", parsed, diffRaw)
			}
		})
	}
}
