package protection

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

var Tool = tool.New()

const (
	ProtectionReadToolName  = "protection_read"
	ProtectionWriteToolName = "protection_write"
)

var (
	ProtectionReadTool = mcp.NewTool(
		ProtectionReadToolName,
		mcp.WithDescription("Read branch protection. Use method 'list' to list all protections, 'get' to get specific branch protection."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("list", "get")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("branch", mcp.Description("branch name (required for 'get')")),
	)

	ProtectionWriteTool = mcp.NewTool(
		ProtectionWriteToolName,
		mcp.WithDescription("Create, update, or delete branch protection rules."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("create", "edit", "delete")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("branch", mcp.Required(), mcp.Description("branch name")),
		mcp.WithBoolean("require_signed_commits", mcp.Description("require signed commits")),
		mcp.WithBoolean("enable_status_check", mcp.Description("enable status checks")),
		mcp.WithNumber("required_approvals", mcp.Description("required approval count")),
		mcp.WithBoolean("dismiss_stale_approvals", mcp.Description("dismiss stale approvals")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ProtectionReadTool,
		Handler: protectionReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    ProtectionWriteTool,
		Handler: protectionWriteFn,
	})
}

func protectionReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "list":
		return listProtectionsFn(ctx, req)
	case "get":
		return getProtectionFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func protectionWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "create":
		return createProtectionFn(ctx, req)
	case "edit":
		return editProtectionFn(ctx, req)
	case "delete":
		return deleteProtectionFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func listProtectionsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listProtectionsFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	protections, _, err := client.ListBranchProtections(owner, repo, gitea_sdk.ListBranchProtectionsOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list branch protections err: %v", err))
	}
	return to.TextResult(slimProtections(protections))
}

func getProtectionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getProtectionFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	branch, err := params.GetString(req.GetArguments(), "branch")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	protection, _, err := client.GetBranchProtection(owner, repo, branch)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get branch protection err: %v", err))
	}
	return to.TextResult(slimProtection(protection))
}

func createProtectionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called createProtectionFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	branch, err := params.GetString(req.GetArguments(), "branch")
	if err != nil {
		return to.ErrorResult(err)
	}

	args := req.GetArguments()
	opt := gitea_sdk.CreateBranchProtectionOption{
		BranchName: branch,
	}
	if v, ok := args["require_signed_commits"].(bool); ok {
		opt.RequireSignedCommits = v
	}
	if v, ok := args["enable_status_check"].(bool); ok {
		opt.EnableStatusCheck = v
	}
	if v, ok := args["required_approvals"].(float64); ok {
		opt.RequiredApprovals = int64(v)
	}
	if v, ok := args["dismiss_stale_approvals"].(bool); ok {
		opt.DismissStaleApprovals = v
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	protection, _, err := client.CreateBranchProtection(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create branch protection err: %v", err))
	}
	return to.TextResult(slimProtection(protection))
}

func editProtectionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called editProtectionFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	branch, err := params.GetString(req.GetArguments(), "branch")
	if err != nil {
		return to.ErrorResult(err)
	}

	args := req.GetArguments()
	opt := gitea_sdk.EditBranchProtectionOption{}
	if v, ok := args["require_signed_commits"].(bool); ok {
		opt.RequireSignedCommits = &v
	}
	if v, ok := args["enable_status_check"].(bool); ok {
		opt.EnableStatusCheck = &v
	}
	if v, ok := args["required_approvals"].(float64); ok {
		vv := int64(v)
		opt.RequiredApprovals = &vv
	}
	if v, ok := args["dismiss_stale_approvals"].(bool); ok {
		opt.DismissStaleApprovals = &v
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	protection, _, err := client.EditBranchProtection(owner, repo, branch, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit branch protection err: %v", err))
	}
	return to.TextResult(slimProtection(protection))
}

func deleteProtectionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called deleteProtectionFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	branch, err := params.GetString(req.GetArguments(), "branch")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteBranchProtection(owner, repo, branch)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete branch protection err: %v", err))
	}
	return to.TextResult("Branch protection deleted successfully")
}

func slimProtections(bps []*gitea_sdk.BranchProtection) []map[string]any {
	out := make([]map[string]any, 0, len(bps))
	for _, bp := range bps {
		out = append(out, slimProtection(bp))
	}
	return out
}

func slimProtection(bp *gitea_sdk.BranchProtection) map[string]any {
	if bp == nil {
		return nil
	}
	return map[string]any{
		"branch_name":             bp.BranchName,
		"rule_name":               bp.RuleName,
		"require_signed_commits":  bp.RequireSignedCommits,
		"enable_status_check":     bp.EnableStatusCheck,
		"required_approvals":      bp.RequiredApprovals,
		"dismiss_stale_approvals": bp.DismissStaleApprovals,
	}
}
