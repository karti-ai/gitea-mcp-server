package sshkey

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
	ListMySSHKeysToolName   = "list_my_ssh_keys"
	GetSSHKeyToolName       = "get_ssh_key"
	CreateSSHKeyToolName    = "create_ssh_key"
	DeleteSSHKeyToolName    = "delete_ssh_key"
	ListUserSSHKeysToolName = "list_user_ssh_keys"
)

var Tool = tool.New()

var (
	ListMySSHKeysTool = mcp.NewTool(
		ListMySSHKeysToolName,
		mcp.WithDescription("List SSH keys for the authenticated user"),
	)

	GetSSHKeyTool = mcp.NewTool(
		GetSSHKeyToolName,
		mcp.WithDescription("Get a specific SSH key by ID"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("SSH key ID")),
	)

	CreateSSHKeyTool = mcp.NewTool(
		CreateSSHKeyToolName,
		mcp.WithDescription("Create a new SSH key for the authenticated user"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title/description for the SSH key")),
		mcp.WithString("key", mcp.Required(), mcp.Description("The SSH public key content")),
	)

	DeleteSSHKeyTool = mcp.NewTool(
		DeleteSSHKeyToolName,
		mcp.WithDescription("Delete an SSH key"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("SSH key ID to delete")),
	)

	ListUserSSHKeysTool = mcp.NewTool(
		ListUserSSHKeysToolName,
		mcp.WithDescription("List SSH keys for a specific user"),
		mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListMySSHKeysTool,
		Handler: listMySSHKeysFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetSSHKeyTool,
		Handler: getSSHKeyFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListUserSSHKeysTool,
		Handler: listUserSSHKeysFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreateSSHKeyTool,
		Handler: createSSHKeyFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeleteSSHKeyTool,
		Handler: deleteSSHKeyFn,
	})
}

func listMySSHKeysFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[SSHKey] Called listMySSHKeysFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	keys, _, err := client.ListMyPublicKeys(gitea_sdk.ListPublicKeysOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list SSH keys err: %v", err))
	}
	return to.TextResult(slimSSHKeys(keys))
}

func getSSHKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[SSHKey] Called getSSHKeyFn")
	args := req.GetArguments()
	id, err := params.GetIndex(args, "id")
	if err != nil {
		return to.ErrorResult(fmt.Errorf("invalid key id: %v", err))
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	key, _, err := client.GetPublicKey(id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get SSH key err: %v", err))
	}
	return to.TextResult(slimSSHKey(key))
}

func createSSHKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[SSHKey] Called createSSHKeyFn")
	args := req.GetArguments()
	title, err := params.GetString(args, "title")
	if err != nil {
		return to.ErrorResult(err)
	}
	key, err := params.GetString(args, "key")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	createOpt := gitea_sdk.CreateKeyOption{
		Title: title,
		Key:   key,
	}
	respKey, _, err := client.CreatePublicKey(createOpt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create SSH key err: %v", err))
	}
	return to.TextResult(slimSSHKey(respKey))
}

func deleteSSHKeyFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[SSHKey] Called deleteSSHKeyFn")
	args := req.GetArguments()
	id, err := params.GetIndex(args, "id")
	if err != nil {
		return to.ErrorResult(fmt.Errorf("invalid key id: %v", err))
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeletePublicKey(id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete SSH key err: %v", err))
	}
	return to.TextResult("SSH key deleted successfully")
}

func listUserSSHKeysFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[SSHKey] Called listUserSSHKeysFn")
	args := req.GetArguments()
	username, err := params.GetString(args, "username")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	keys, _, err := client.ListPublicKeys(username, gitea_sdk.ListPublicKeysOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list user SSH keys err: %v", err))
	}
	return to.TextResult(slimSSHKeys(keys))
}

func slimSSHKeys(keys []*gitea_sdk.PublicKey) []map[string]interface{} {
	result := make([]map[string]interface{}, len(keys))
	for i, k := range keys {
		result[i] = slimSSHKey(k)
	}
	return result
}

func slimSSHKey(k *gitea_sdk.PublicKey) map[string]interface{} {
	return map[string]interface{}{
		"id":          k.ID,
		"key":         k.Key,
		"title":       k.Title,
		"created":     k.Created,
		"fingerprint": k.Fingerprint,
	}
}
