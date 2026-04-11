package collaborator

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
	ListCollaboratorsToolName      = "list_collaborators"
	GetCollaboratorToolName        = "get_collaborator"
	AddCollaboratorToolName        = "add_collaborator"
	DeleteCollaboratorToolName     = "delete_collaborator"
	CollaboratorPermissionToolName = "collaborator_permission"
)

var Tool = tool.New()

var (
	ListCollaboratorsTool = mcp.NewTool(
		ListCollaboratorsToolName,
		mcp.WithDescription("List collaborators for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("role", mcp.Description("Filter by role (admin, write, read, none)")),
	)

	GetCollaboratorTool = mcp.NewTool(
		GetCollaboratorToolName,
		mcp.WithDescription("Check if a user is a collaborator"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("collaborator", mcp.Required(), mcp.Description("username to check")),
	)

	AddCollaboratorTool = mcp.NewTool(
		AddCollaboratorToolName,
		mcp.WithDescription("Add a collaborator to a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("collaborator", mcp.Required(), mcp.Description("username to add")),
		mcp.WithString("permission", mcp.Description("Permission level"), mcp.Enum("read", "write", "admin")),
	)

	DeleteCollaboratorTool = mcp.NewTool(
		DeleteCollaboratorToolName,
		mcp.WithDescription("Remove a collaborator from a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("collaborator", mcp.Required(), mcp.Description("username to remove")),
	)

	CollaboratorPermissionTool = mcp.NewTool(
		CollaboratorPermissionToolName,
		mcp.WithDescription("Get collaborator permission details"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("collaborator", mcp.Required(), mcp.Description("username")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListCollaboratorsTool,
		Handler: listCollaboratorsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetCollaboratorTool,
		Handler: getCollaboratorFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    CollaboratorPermissionTool,
		Handler: collaboratorPermissionFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    AddCollaboratorTool,
		Handler: addCollaboratorFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteCollaboratorTool,
		Handler: deleteCollaboratorFn,
	})
}

func listCollaboratorsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Collaborator] Called listCollaboratorsFn")
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

	opt := gitea_sdk.ListCollaboratorsOptions{}

	collabs, _, err := client.ListCollaborators(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list collaborators err: %v", err))
	}

	return to.TextResult(slimUsers(collabs))
}

func getCollaboratorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Collaborator] Called getCollaboratorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	collaborator, err := params.GetString(args, "collaborator")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	isCollab, _, err := client.IsCollaborator(owner, repo, collaborator)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("check collaborator err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"is_collaborator": isCollab,
	})
}

func addCollaboratorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Collaborator] Called addCollaboratorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	collaborator, err := params.GetString(args, "collaborator")
	if err != nil {
		return to.ErrorResult(err)
	}
	permission := params.GetOptionalString(args, "permission", "read")

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	mode := gitea_sdk.AccessMode(permission)
	opt := gitea_sdk.AddCollaboratorOption{
		Permission: &mode,
	}

	_, err = client.AddCollaborator(owner, repo, collaborator, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add collaborator err: %v", err))
	}

	return to.TextResult("Collaborator added successfully")
}

func deleteCollaboratorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Collaborator] Called deleteCollaboratorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	collaborator, err := params.GetString(args, "collaborator")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, err = client.DeleteCollaborator(owner, repo, collaborator)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete collaborator err: %v", err))
	}

	return to.TextResult("Collaborator removed successfully")
}

func collaboratorPermissionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Collaborator] Called collaboratorPermissionFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	collaborator, err := params.GetString(args, "collaborator")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	result, _, err := client.CollaboratorPermission(owner, repo, collaborator)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get collaborator permission err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"permission": result.Permission,
		"role":       result.Role,
		"user":       result.User.UserName,
	})
}

func slimUsers(users []*gitea_sdk.User) []map[string]interface{} {
	result := make([]map[string]interface{}, len(users))
	for i, u := range users {
		result[i] = map[string]interface{}{
			"id":       u.ID,
			"username": u.UserName,
			"name":     u.FullName,
			"email":    u.Email,
		}
	}
	return result
}
