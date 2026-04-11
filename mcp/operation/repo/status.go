package repo

import (
	"context"
	"fmt"

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
	CreateCommitStatusToolName = "create_commit_status"
)

var (
	CreateCommitStatusTool = mcp.NewTool(
		CreateCommitStatusToolName,
		mcp.WithDescription("Create a commit status check for a repository. Adds a new status context to a commit without overwriting existing statuses."),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("sha", mcp.Required(), mcp.Description("commit SHA (full 40-character SHA or short SHA)")),
		mcp.WithString("state", mcp.Required(), mcp.Description("status state: pending, success, error, or failure")),
		mcp.WithString("target_url", mcp.Description("URL with more details about the status (e.g., review environment link like https://review.lumbridgecorp.com)")),
		mcp.WithString("context", mcp.Description("status context identifier (e.g., 'ci/metal', 'ci/cloud-1', 'continuous-integration/jenkins')"), mcp.DefaultString("default")),
		mcp.WithString("description", mcp.Description("short description of the status")),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateCommitStatusTool,
		Handler: CreateCommitStatusFn,
	})
}

// CreateCommitStatusFn creates a status check for a commit
func CreateCommitStatusFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CreateCommitStatusFn")
	args := req.GetArguments()

	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}

	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}

	sha, err := params.GetString(args, "sha")
	if err != nil {
		return to.ErrorResult(err)
	}

	state, err := params.GetString(args, "state")
	if err != nil {
		return to.ErrorResult(err)
	}

	statusState, err := parseStatusState(state)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("invalid state '%s': must be one of pending, success, failure, error", state))
	}

	targetURL, _ := args["target_url"].(string)
	context, _ := args["context"].(string)
	description, _ := args["description"].(string)

	// Use default context if not provided
	if context == "" {
		context = "default"
	}

	opt := gitea_sdk.CreateStatusOption{
		State:       statusState,
		TargetURL:   targetURL,
		Context:     context,
		Description: description,
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	status, _, err := client.CreateStatus(owner, repo, sha, opt)
	if err != nil {
		translatedErr := errors.TranslateError(err, map[string]string{
			"operation": "CreateCommitStatus",
			"owner":     owner,
			"repo":      repo,
			"sha":       sha,
			"state":     state,
			"context":   context,
		})
		return to.ErrorResult(translatedErr)
	}

	return to.TextResult(slimStatus(status))
}

// parseStatusState converts a string state to gitea_sdk.StatusState
func parseStatusState(state string) (gitea_sdk.StatusState, error) {
	switch state {
	case "pending":
		return gitea_sdk.StatusPending, nil
	case "success":
		return gitea_sdk.StatusSuccess, nil
	case "failure":
		return gitea_sdk.StatusFailure, nil
	case "error":
		return gitea_sdk.StatusError, nil
	default:
		return "", fmt.Errorf("invalid state: %s", state)
	}
}

// slimStatus creates a slimmed down representation of a commit status
func slimStatus(s *gitea_sdk.Status) map[string]any {
	if s == nil {
		return nil
	}
	return map[string]any{
		"id":          s.ID,
		"state":       s.State,
		"target_url":  s.TargetURL,
		"context":     s.Context,
		"description": s.Description,
	}
}
