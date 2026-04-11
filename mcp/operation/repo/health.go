package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitea.com/gitea/gitea-mcp/pkg/errors"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	RepoHealthCheckToolName = "repo_health_check"
)

var (
	RepoHealthCheckTool = mcp.NewTool(
		RepoHealthCheckToolName,
		mcp.WithDescription("Check repository health by aggregating multiple status metrics including last commit date, open issues/PRs count, workflow status, and branch protection. Returns a comprehensive health score (0-100)."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithBoolean("include_workflows", mcp.Description("include workflow run status (may require additional API calls)"), mcp.DefaultBool(true)),
		mcp.WithBoolean("include_protection", mcp.Description("include branch protection status"), mcp.DefaultBool(true)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    RepoHealthCheckTool,
		Handler: repoHealthCheckFn,
	})
}

// HealthResult represents the complete health check result
type HealthResult struct {
	Repository       string                `json:"repository"`
	HealthScore      int                   `json:"health_score"`
	HealthStatus     string                `json:"health_status"`
	LastCommit       *CommitInfo           `json:"last_commit,omitempty"`
	Issues           *IssuesInfo           `json:"issues,omitempty"`
	PullRequests     *PullRequestsInfo     `json:"pull_requests,omitempty"`
	WorkflowStatus   *WorkflowStatusInfo   `json:"workflow_status,omitempty"`
	BranchProtection *BranchProtectionInfo `json:"branch_protection,omitempty"`
	RepositoryInfo   *RepositoryInfo       `json:"repository_info,omitempty"`
	Errors           []HealthCheckError    `json:"errors,omitempty"`
	CheckedAt        string                `json:"checked_at"`
	PartialResult    bool                  `json:"partial_result"`
}

// CommitInfo contains last commit information
type CommitInfo struct {
	SHA       string `json:"sha"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Date      string `json:"date"`
	DaysAgo   int    `json:"days_ago"`
	Available bool   `json:"available"`
}

// IssuesInfo contains issue metrics
type IssuesInfo struct {
	OpenCount  int  `json:"open_count"`
	TotalCount int  `json:"total_count"`
	Available  bool `json:"available"`
}

// PullRequestsInfo contains PR metrics
type PullRequestsInfo struct {
	OpenCount  int  `json:"open_count"`
	TotalCount int  `json:"total_count"`
	Available  bool `json:"available"`
}

// WorkflowStatusInfo contains workflow information
type WorkflowStatusInfo struct {
	LastRunStatus     string `json:"last_run_status,omitempty"`
	LastRunConclusion string `json:"last_run_conclusion,omitempty"`
	HasRecentRuns     bool   `json:"has_recent_runs"`
	Available         bool   `json:"available"`
	Error             string `json:"error,omitempty"`
}

// BranchProtectionInfo contains protection metrics
type BranchProtectionInfo struct {
	ProtectedBranchesCount int      `json:"protected_branches_count"`
	ProtectedBranches      []string `json:"protected_branches,omitempty"`
	Available              bool     `json:"available"`
}

// RepositoryInfo contains basic repo metrics
type RepositoryInfo struct {
	Stars      int    `json:"stars"`
	Forks      int    `json:"forks"`
	Language   string `json:"language,omitempty"`
	IsPrivate  bool   `json:"is_private"`
	IsArchived bool   `json:"is_archived"`
	Available  bool   `json:"available"`
}

// HealthCheckError represents an error from a specific check
type HealthCheckError struct {
	Check string `json:"check"`
	Error string `json:"error"`
}

func repoHealthCheckFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called repoHealthCheckFn")

	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "RepoHealthCheck",
			"param":     "owner",
		}))
	}

	repoName, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "RepoHealthCheck",
			"param":     "repo",
		}))
	}

	includeWorkflows := params.GetOptionalBool(req.GetArguments(), "include_workflows", true)
	includeProtection := params.GetOptionalBool(req.GetArguments(), "include_protection", true)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "RepoHealthCheck",
			"owner":     owner,
			"repo":      repoName,
		}))
	}

	result := &HealthResult{
		Repository:    fmt.Sprintf("%s/%s", owner, repoName),
		CheckedAt:     time.Now().UTC().Format(time.RFC3339),
		PartialResult: false,
		Errors:        []HealthCheckError{},
	}

	// Check 1: Repository Info (always try first)
	repoInfo, err := checkRepositoryInfo(ctx, client, owner, repoName)
	if err != nil {
		result.Errors = append(result.Errors, HealthCheckError{
			Check: "repository_info",
			Error: err.Error(),
		})
		result.PartialResult = true
	} else {
		result.RepositoryInfo = repoInfo
	}

	// Check 2: Last Commit
	commitInfo, err := checkLastCommit(ctx, client, owner, repoName)
	if err != nil {
		result.Errors = append(result.Errors, HealthCheckError{
			Check: "last_commit",
			Error: err.Error(),
		})
		result.PartialResult = true
	} else {
		result.LastCommit = commitInfo
	}

	// Check 3: Issues
	issuesInfo, err := checkIssues(ctx, client, owner, repoName)
	if err != nil {
		result.Errors = append(result.Errors, HealthCheckError{
			Check: "issues",
			Error: err.Error(),
		})
		result.PartialResult = true
	} else {
		result.Issues = issuesInfo
	}

	// Check 4: Pull Requests
	prsInfo, err := checkPullRequests(ctx, client, owner, repoName)
	if err != nil {
		result.Errors = append(result.Errors, HealthCheckError{
			Check: "pull_requests",
			Error: err.Error(),
		})
		result.PartialResult = true
	} else {
		result.PullRequests = prsInfo
	}

	// Check 5: Workflow Status (optional, may fail on older Gitea versions)
	if includeWorkflows {
		workflowInfo, err := checkWorkflowStatus(ctx, owner, repoName)
		if err != nil {
			// Don't mark as partial for workflow errors on older Gitea versions
			if !errors.IsActionsAPIUnavailable(err) {
				result.Errors = append(result.Errors, HealthCheckError{
					Check: "workflow_status",
					Error: err.Error(),
				})
			}
			result.WorkflowStatus = &WorkflowStatusInfo{
				Available: false,
				Error:     err.Error(),
			}
		} else {
			result.WorkflowStatus = workflowInfo
		}
	}

	// Check 6: Branch Protection (optional)
	if includeProtection {
		protectionInfo, err := checkBranchProtection(ctx, client, owner, repoName)
		if err != nil {
			result.Errors = append(result.Errors, HealthCheckError{
				Check: "branch_protection",
				Error: err.Error(),
			})
			result.PartialResult = true
			result.BranchProtection = &BranchProtectionInfo{
				Available: false,
			}
		} else {
			result.BranchProtection = protectionInfo
		}
	}

	// Calculate health score
	result.HealthScore = calculateHealthScore(result)
	result.HealthStatus = getHealthStatus(result.HealthScore)

	// Return result as JSON
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return to.ErrorResult(errors.TranslateError(err, map[string]string{
			"operation": "RepoHealthCheck",
			"step":      "marshal_result",
		}))
	}

	return to.TextResult(string(jsonBytes))
}

func checkRepositoryInfo(ctx context.Context, client *gitea_sdk.Client, owner, repo string) (*RepositoryInfo, error) {
	r, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "GetRepo",
			"owner":     owner,
			"repo":      repo,
		})
	}

	return &RepositoryInfo{
		Stars:      r.Stars,
		Forks:      r.Forks,
		Language:   r.Language,
		IsPrivate:  r.Private,
		IsArchived: r.Archived,
		Available:  true,
	}, nil
}

func checkLastCommit(ctx context.Context, client *gitea_sdk.Client, owner, repo string) (*CommitInfo, error) {
	opt := gitea_sdk.ListCommitOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	}

	commits, _, err := client.ListRepoCommits(owner, repo, opt)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListRepoCommits",
			"owner":     owner,
			"repo":      repo,
		})
	}

	if len(commits) == 0 {
		return &CommitInfo{
			Available: false,
		}, nil
	}

	c := commits[0]
	info := &CommitInfo{
		SHA:       c.SHA,
		Available: true,
	}

	if c.RepoCommit != nil {
		info.Message = c.RepoCommit.Message
		if c.RepoCommit.Author != nil {
			info.Author = c.RepoCommit.Author.Name
			info.Date = c.RepoCommit.Author.Date
			// Calculate days ago
			if commitTime, err := time.Parse(time.RFC3339, c.RepoCommit.Author.Date); err == nil {
				info.DaysAgo = int(time.Since(commitTime).Hours() / 24)
			}
		}
	}

	return info, nil
}

func checkIssues(ctx context.Context, client *gitea_sdk.Client, owner, repo string) (*IssuesInfo, error) {
	// Get open issues count
	openOpt := gitea_sdk.ListIssueOption{
		State: gitea_sdk.StateOpen,
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	}

	openIssues, _, err := client.ListRepoIssues(owner, repo, openOpt)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListRepoIssues",
			"owner":     owner,
			"repo":      repo,
			"state":     "open",
		})
	}

	// Get total issues count (we can use the repo info for this to save API calls)
	// For simplicity, we'll just use what we can get from list
	totalOpt := gitea_sdk.ListIssueOption{
		State: gitea_sdk.StateAll,
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	}

	totalIssues, _, err := client.ListRepoIssues(owner, repo, totalOpt)
	if err != nil {
		// If we got open count, we can still return partial info
		return &IssuesInfo{
			OpenCount: len(openIssues),
			Available: true,
		}, nil
	}

	return &IssuesInfo{
		OpenCount:  len(openIssues),
		TotalCount: len(totalIssues),
		Available:  true,
	}, nil
}

func checkPullRequests(ctx context.Context, client *gitea_sdk.Client, owner, repo string) (*PullRequestsInfo, error) {
	// Get open PRs
	openOpt := gitea_sdk.ListPullRequestsOptions{
		State: gitea_sdk.StateOpen,
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	}

	openPRs, _, err := client.ListRepoPullRequests(owner, repo, openOpt)
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListRepoPullRequests",
			"owner":     owner,
			"repo":      repo,
			"state":     "open",
		})
	}

	// Get total PRs
	totalOpt := gitea_sdk.ListPullRequestsOptions{
		State: gitea_sdk.StateAll,
		ListOptions: gitea_sdk.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	}

	totalPRs, _, err := client.ListRepoPullRequests(owner, repo, totalOpt)
	if err != nil {
		return &PullRequestsInfo{
			OpenCount: len(openPRs),
			Available: true,
		}, nil
	}

	return &PullRequestsInfo{
		OpenCount:  len(openPRs),
		TotalCount: len(totalPRs),
		Available:  true,
	}, nil
}

func checkWorkflowStatus(ctx context.Context, owner, repo string) (*WorkflowStatusInfo, error) {
	// Use the REST API directly to get recent workflow runs
	var result struct {
		WorkflowRuns []map[string]any `json:"workflow_runs"`
	}

	status, err := gitea.DoJSON(ctx, "GET", fmt.Sprintf("repos/%s/%s/actions/runs", owner, repo), nil, nil, &result)
	if err != nil {
		// Check if this is an Actions API unavailability error
		if status == 404 || status == 405 {
			return nil, errors.NewEnhancedError(
				err,
				"Actions API not available on this Gitea version",
				errors.CategoryActions,
			).WithOperation("CheckWorkflowStatus")
		}
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListWorkflowRuns",
			"owner":     owner,
			"repo":      repo,
		})
	}

	info := &WorkflowStatusInfo{
		Available: len(result.WorkflowRuns) > 0,
	}

	if len(result.WorkflowRuns) > 0 {
		// Get the most recent run
		run := result.WorkflowRuns[0]
		info.LastRunStatus = getStringFromMap(run, "status")
		info.LastRunConclusion = getStringFromMap(run, "conclusion")
		info.HasRecentRuns = true

		// Check if run is recent (within 7 days)
		if createdAt := getStringFromMap(run, "created_at"); createdAt != "" {
			if runTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
				info.HasRecentRuns = time.Since(runTime).Hours() < 24*7
			}
		}
	}

	return info, nil
}

func checkBranchProtection(ctx context.Context, client *gitea_sdk.Client, owner, repo string) (*BranchProtectionInfo, error) {
	protections, _, err := client.ListBranchProtections(owner, repo, gitea_sdk.ListBranchProtectionsOptions{})
	if err != nil {
		return nil, errors.TranslateError(err, map[string]string{
			"operation": "ListBranchProtections",
			"owner":     owner,
			"repo":      repo,
		})
	}

	branches := make([]string, 0, len(protections))
	for _, p := range protections {
		branches = append(branches, p.BranchName)
	}

	return &BranchProtectionInfo{
		ProtectedBranchesCount: len(protections),
		ProtectedBranches:      branches,
		Available:              true,
	}, nil
}

func calculateHealthScore(result *HealthResult) int {
	score := 100

	// Deduct for stale commits (more than 30 days)
	if result.LastCommit != nil && result.LastCommit.Available {
		if result.LastCommit.DaysAgo > 90 {
			score -= 30
		} else if result.LastCommit.DaysAgo > 30 {
			score -= 15
		}
	}

	// Deduct for too many open issues (relative scoring)
	if result.Issues != nil && result.Issues.Available {
		if result.Issues.OpenCount > 50 {
			score -= 10
		} else if result.Issues.OpenCount > 20 {
			score -= 5
		}
	}

	// Deduct for old/stale PRs
	if result.PullRequests != nil && result.PullRequests.Available {
		if result.PullRequests.OpenCount > 10 {
			score -= 5
		}
	}

	// Deduct for workflow failures
	if result.WorkflowStatus != nil && result.WorkflowStatus.Available {
		if result.WorkflowStatus.LastRunConclusion == "failure" {
			score -= 15
		} else if result.WorkflowStatus.LastRunConclusion == "cancelled" {
			score -= 5
		}
		if !result.WorkflowStatus.HasRecentRuns {
			score -= 5
		}
	}

	// Bonus for good practices
	if result.BranchProtection != nil && result.BranchProtection.Available {
		if result.BranchProtection.ProtectedBranchesCount > 0 {
			score += 5 // Bonus for having protected branches
		}
	}

	// Penalty for archived repos
	if result.RepositoryInfo != nil && result.RepositoryInfo.IsArchived {
		score -= 40
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func getHealthStatus(score int) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 70:
		return "good"
	case score >= 50:
		return "fair"
	case score >= 30:
		return "poor"
	default:
		return "critical"
	}
}

func getStringFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
