package settings

import (
	"context"
	"fmt"

	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/params"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	gitea_sdk "code.gitea.io/sdk/gitea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	GetUserSettingsToolName    = "get_user_settings"
	UpdateUserSettingsToolName = "update_user_settings"
	GetRepoSettingsToolName    = "get_repo_settings"
	UpdateRepoSettingsToolName = "update_repo_settings"
)

var Tool = tool.New()

var (
	GetUserSettingsTool = mcp.NewTool(
		GetUserSettingsToolName,
		mcp.WithDescription("Get current user's settings"),
	)

	UpdateUserSettingsTool = mcp.NewTool(
		UpdateUserSettingsToolName,
		mcp.WithDescription("Update current user's settings"),
		mcp.WithString("description", mcp.Description("User description")),
		mcp.WithString("website", mcp.Description("Website URL")),
		mcp.WithString("location", mcp.Description("Location")),
		mcp.WithString("theme", mcp.Description("Theme preference")),
	)

	GetRepoSettingsTool = mcp.NewTool(
		GetRepoSettingsToolName,
		mcp.WithDescription("Get repository settings"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)

	UpdateRepoSettingsTool = mcp.NewTool(
		UpdateRepoSettingsToolName,
		mcp.WithDescription("Update repository settings"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("description", mcp.Description("Repository description")),
		mcp.WithString("website", mcp.Description("Website URL")),
		mcp.WithString("default_branch", mcp.Description("Default branch name")),
		mcp.WithBoolean("private", mcp.Description("Whether repo is private")),
		mcp.WithBoolean("protected", mcp.Description("Whether repo is protected")),
		mcp.WithBoolean("enable_wiki", mcp.Description("Enable wiki")),
		mcp.WithBoolean("enable_issues", mcp.Description("Enable issues")),
		mcp.WithBoolean("enable_pull_requests", mcp.Description("Enable pull requests")),
		mcp.WithString("default_merge_style", mcp.Description("Default merge style"), mcp.Enum("merge", "rebase", "squash")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetUserSettingsTool,
		Handler: getUserSettingsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetRepoSettingsTool,
		Handler: getRepoSettingsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    UpdateUserSettingsTool,
		Handler: updateUserSettingsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    UpdateRepoSettingsTool,
		Handler: updateRepoSettingsFn,
	})
}

func getUserSettingsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Settings] Called getUserSettingsFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	settings, _, err := client.GetUserSettings()
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get user settings err: %v", err))
	}
	return to.TextResult(slimUserSettings(settings))
}

func updateUserSettingsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Settings] Called updateUserSettingsFn")
	args := req.GetArguments()

	opt := gitea_sdk.UserSettingsOptions{}
	if v, ok := args["description"].(string); ok && v != "" {
		opt.Description = &v
	}
	if v, ok := args["website"].(string); ok && v != "" {
		opt.Website = &v
	}
	if v, ok := args["location"].(string); ok && v != "" {
		opt.Location = &v
	}
	if v, ok := args["theme"].(string); ok && v != "" {
		opt.Theme = &v
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	settings, _, err := client.UpdateUserSettings(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update user settings err: %v", err))
	}
	return to.TextResult(slimUserSettings(settings))
}

func getRepoSettingsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Settings] Called getRepoSettingsFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	r, _, err := client.GetRepo(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get repo err: %v", err))
	}
	return to.TextResult(slimRepoSettings(r))
}

func updateRepoSettingsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Settings] Called updateRepoSettingsFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditRepoOption{}
	if v, ok := args["description"].(string); ok && v != "" {
		opt.Description = &v
	}
	if v, ok := args["website"].(string); ok && v != "" {
		opt.Website = &v
	}
	if v, ok := args["default_branch"].(string); ok && v != "" {
		opt.DefaultBranch = &v
	}
	if v, ok := args["private"].(bool); ok {
		opt.Private = &v
	}
	if v, ok := args["protected"].(bool); ok {
		opt.Archived = &v
	}
	if v, ok := args["enable_wiki"].(bool); ok {
		opt.HasWiki = &v
	}
	if v, ok := args["enable_issues"].(bool); ok {
		opt.HasIssues = &v
	}
	if v, ok := args["enable_pull_requests"].(bool); ok {
		opt.HasPullRequests = &v
	}
	if v, ok := args["default_merge_style"].(string); ok && v != "" {
		style := gitea_sdk.MergeStyle(v)
		opt.DefaultMergeStyle = &style
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	r, _, err := client.EditRepo(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update repo settings err: %v", err))
	}
	return to.TextResult(slimRepoSettings(r))
}

func slimUserSettings(s *gitea_sdk.UserSettings) map[string]interface{} {
	return map[string]interface{}{
		"full_name":       s.FullName,
		"description":     s.Description,
		"website":         s.Website,
		"location":        s.Location,
		"theme":           s.Theme,
		"language":        s.Language,
		"diff_view_style": s.DiffViewStyle,
		"hide_email":      s.HideEmail,
		"hide_activity":   s.HideActivity,
	}
}

func slimRepoSettings(r *gitea_sdk.Repository) map[string]interface{} {
	return map[string]interface{}{
		"id":                                r.ID,
		"name":                              r.Name,
		"full_name":                         r.FullName,
		"description":                       r.Description,
		"website":                           r.Website,
		"default_branch":                    r.DefaultBranch,
		"private":                           r.Private,
		"has_wiki":                          r.HasWiki,
		"has_issues":                        r.HasIssues,
		"has_projects":                      r.HasProjects,
		"default_merge_style":               r.DefaultMergeStyle,
		"default_delete_branch_after_merge": r.DefaultDeleteBranchAfterMerge,
	}
}
