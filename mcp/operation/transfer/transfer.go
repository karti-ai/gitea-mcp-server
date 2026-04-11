package transfer

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
	TransferRepoToolName   = "transfer_repo"
	AcceptTransferToolName = "accept_transfer"
	RejectTransferToolName = "reject_transfer"
)

var Tool = tool.New()

var (
	TransferRepoTool = mcp.NewTool(
		TransferRepoToolName,
		mcp.WithDescription("Transfer repository ownership to another user or organization"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Current repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("new_owner", mcp.Required(), mcp.Description("New owner (user or org)")),
		mcp.WithArray("teams", mcp.Description("Teams to transfer (for org-to-org)"), mcp.Items(map[string]any{"type": "string"})),
	)

	AcceptTransferTool = mcp.NewTool(
		AcceptTransferToolName,
		mcp.WithDescription("Accept a repository transfer"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
	)

	RejectTransferTool = mcp.NewTool(
		RejectTransferToolName,
		mcp.WithDescription("Reject a repository transfer"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("Repository name")),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    TransferRepoTool,
		Handler: transferRepoFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    AcceptTransferTool,
		Handler: acceptTransferFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    RejectTransferTool,
		Handler: rejectTransferFn,
	})
}

func transferRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Transfer] Called transferRepoFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	newOwner, err := params.GetString(args, "new_owner")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.TransferRepoOption{
		NewOwner: newOwner,
	}

	r, _, err := client.TransferRepo(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("transfer repo err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"id":        r.ID,
		"name":      r.Name,
		"full_name": r.FullName,
		"owner":     r.Owner.UserName,
	})
}

func acceptTransferFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Transfer] Called acceptTransferFn")
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

	r, _, err := client.AcceptRepoTransfer(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("accept transfer err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"id":        r.ID,
		"name":      r.Name,
		"full_name": r.FullName,
		"owner":     r.Owner.UserName,
	})
}

func rejectTransferFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Transfer] Called rejectTransferFn")
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

	_, _, err = client.RejectRepoTransfer(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("reject transfer err: %v", err))
	}

	return to.TextResult("Repository transfer rejected")
}
