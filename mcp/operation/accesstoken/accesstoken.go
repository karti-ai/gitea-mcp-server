package accesstoken

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
	CreateAccessTokenToolName = "create_access_token"
	DeleteAccessTokenToolName = "delete_access_token"
)

var Tool = tool.New()

var (
	CreateAccessTokenTool = mcp.NewTool(
		CreateAccessTokenToolName,
		mcp.WithDescription("Create a new personal access token"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Token name/description")),
		mcp.WithArray("scopes", mcp.Description("Array of permission scopes"), mcp.Items(map[string]any{"type": "string"})),
	)

	DeleteAccessTokenTool = mcp.NewTool(
		DeleteAccessTokenToolName,
		mcp.WithDescription("Delete a personal access token"),
		mcp.WithString("token", mcp.Required(), mcp.Description("The token value or name to delete")),
	)
)

func init() {
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateAccessTokenTool,
		Handler: createAccessTokenFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteAccessTokenTool,
		Handler: deleteAccessTokenFn,
	})
}

func createAccessTokenFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[AccessToken] Called createAccessTokenFn")
	args := req.GetArguments()
	name, err := params.GetString(args, "name")
	if err != nil {
		return to.ErrorResult(err)
	}

	scopesRaw := params.GetStringSlice(args, "scopes")
	var scopes []gitea_sdk.AccessTokenScope
	if scopesRaw == nil {
		scopes = []gitea_sdk.AccessTokenScope{"repo", "user"}
	} else {
		scopes = make([]gitea_sdk.AccessTokenScope, len(scopesRaw))
		for i, s := range scopesRaw {
			scopes[i] = gitea_sdk.AccessTokenScope(s)
		}
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, _, err = client.GetMyUserInfo()
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get current user err: %v", err))
	}

	createOpt := gitea_sdk.CreateAccessTokenOption{
		Name:   name,
		Scopes: scopes,
	}

	token, _, err := client.CreateAccessToken(createOpt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create access token err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"name":             token.Name,
		"token":            token.Token,
		"id":               token.ID,
		"scopes":           token.Scopes,
		"created":          token.Created,
		"token_last_eight": token.TokenLastEight,
	})
}

func deleteAccessTokenFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[AccessToken] Called deleteAccessTokenFn")
	args := req.GetArguments()
	token, err := params.GetString(args, "token")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, err = client.DeleteAccessToken(token)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete access token err: %v", err))
	}

	return to.TextResult("Access token deleted successfully")
}
