package deploykey

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
	ListDeployKeysToolName  = "list_deploy_keys"
	GetDeployKeyToolName    = "get_deploy_key"
	CreateDeployKeyToolName = "create_deploy_key"
	DeleteDeployKeyToolName = "delete_deploy_key"
)

var Tool = tool.New()

var (
	ListDeployKeysTool = mcp.NewTool(
		ListDeployKeysToolName,
		mcp.WithDescription("List deploy keys for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)

	GetDeployKeyTool = mcp.NewTool(
		GetDeployKeyToolName,
		mcp.WithDescription("Get a specific deploy key by ID"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Deploy key ID")),
	)

	CreateDeployKeyTool = mcp.NewTool(
		CreateDeployKeyToolName,
		mcp.WithDescription("Create a new deploy key for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title/description for the deploy key")),
		mcp.WithString("key", mcp.Required(), mcp.Description("The SSH public key content")),
		mcp.WithBoolean("readOnly", mcp.Description("Whether the key is read-only"), mcp.DefaultBool(false)),
	)

	DeleteDeployKeyTool = mcp.NewTool(
		DeleteDeployKeyToolName,
		mcp.WithDescription("Delete a deploy key"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("Deploy key ID to delete")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListDeployKeysTool,
		Handler: listDeployKeysFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetDeployKeyTool,
		Handler: getDeployKeyFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateDeployKeyTool,
		Handler: createDeployKeyFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteDeployKeyTool,
		Handler: deleteDeployKeyFn,
	})
}

func listDeployKeysFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[DeployKey] Called listDeployKeysFn")
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
	keys, _, err := client.ListDeployKeys(owner, repo, gitea_sdk.ListDeployKeysOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list deploy keys err: %v", err))
	}
	return to.TextResult(slimDeployKeys(keys))
}

func getDeployKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[DeployKey] Called getDeployKeyFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	id, err := params.GetIndex(args, "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	key, _, err := client.GetDeployKey(owner, repo, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get deploy key err: %v", err))
	}
	return to.TextResult(slimDeployKey(key))
}

func createDeployKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[DeployKey] Called createDeployKeyFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	title, err := params.GetString(args, "title")
	if err != nil {
		return to.ErrorResult(err)
	}
	key, err := params.GetString(args, "key")
	if err != nil {
		return to.ErrorResult(err)
	}
	readOnly, _ := args["readOnly"].(bool)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	createOpt := gitea_sdk.CreateKeyOption{
		Title:    title,
		Key:      key,
		ReadOnly: readOnly,
	}
	respKey, _, err := client.CreateDeployKey(owner, repo, createOpt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create deploy key err: %v", err))
	}
	return to.TextResult(slimDeployKey(respKey))
}

func deleteDeployKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[DeployKey] Called deleteDeployKeyFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	id, err := params.GetIndex(args, "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteDeployKey(owner, repo, id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete deploy key err: %v", err))
	}
	return to.TextResult("Deploy key deleted successfully")
}

func slimDeployKeys(keys []*gitea_sdk.DeployKey) []map[string]interface{} {
	result := make([]map[string]interface{}, len(keys))
	for i, k := range keys {
		result[i] = slimDeployKey(k)
	}
	return result
}

func slimDeployKey(k *gitea_sdk.DeployKey) map[string]interface{} {
	return map[string]interface{}{
		"id":          k.ID,
		"key":         k.Key,
		"title":       k.Title,
		"created":     k.Created,
		"fingerprint": k.Fingerprint,
		"read_only":   k.ReadOnly,
	}
}
