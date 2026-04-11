package mirror

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
	ListPushMirrorsToolName  = "list_push_mirrors"
	GetPushMirrorToolName    = "get_push_mirror"
	CreatePushMirrorToolName = "create_push_mirror"
	DeletePushMirrorToolName = "delete_push_mirror"
	SyncMirrorToolName       = "sync_mirror"
)

var Tool = tool.New()

var (
	ListPushMirrorsTool = mcp.NewTool(
		ListPushMirrorsToolName,
		mcp.WithDescription("List push mirrors for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)

	GetPushMirrorTool = mcp.NewTool(
		GetPushMirrorToolName,
		mcp.WithDescription("Get a specific push mirror by remote name"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("remote_name", mcp.Required(), mcp.Description("Remote name")),
	)

	CreatePushMirrorTool = mcp.NewTool(
		CreatePushMirrorToolName,
		mcp.WithDescription("Create a push mirror for a repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("remote_name", mcp.Required(), mcp.Description("Remote name")),
		mcp.WithString("remote_addr", mcp.Required(), mcp.Description("Remote address (git URL)")),
		mcp.WithString("sync_interval", mcp.Description("Sync interval (e.g. 8h30m0s)")),
		mcp.WithBoolean("sync_on_commit", mcp.Description("Sync on commit"), mcp.DefaultBool(false)),
	)

	DeletePushMirrorTool = mcp.NewTool(
		DeletePushMirrorToolName,
		mcp.WithDescription("Delete a push mirror"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("remote_name", mcp.Required(), mcp.Description("Remote name")),
	)

	SyncMirrorTool = mcp.NewTool(
		SyncMirrorToolName,
		mcp.WithDescription("Trigger immediate sync for a mirror repository"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListPushMirrorsTool,
		Handler: listPushMirrorsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetPushMirrorTool,
		Handler: getPushMirrorFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    CreatePushMirrorTool,
		Handler: createPushMirrorFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeletePushMirrorTool,
		Handler: deletePushMirrorFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    SyncMirrorTool,
		Handler: syncMirrorFn,
	})
}

func listPushMirrorsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Mirror] Called listPushMirrorsFn")
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

	mirrors, _, err := client.ListPushMirrors(owner, repo, gitea_sdk.ListOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list push mirrors err: %v", err))
	}

	return to.TextResult(slimMirrors(mirrors))
}

func getPushMirrorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Mirror] Called getPushMirrorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	remoteName, err := params.GetString(args, "remote_name")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	mirror, _, err := client.GetPushMirrorByRemoteName(owner, repo, remoteName)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get push mirror err: %v", err))
	}

	return to.TextResult(slimMirror(mirror))
}

func createPushMirrorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Mirror] Called createPushMirrorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	remoteAddr, err := params.GetString(args, "remote_addr")
	if err != nil {
		return to.ErrorResult(err)
	}
	syncInterval, _ := args["sync_interval"].(string)
	syncOnCommit, _ := args["sync_on_commit"].(bool)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.CreatePushMirrorOption{
		RemoteAddress: remoteAddr,
		Interval:      syncInterval,
		SyncONCommit:  syncOnCommit,
	}

	mirror, _, err := client.PushMirrors(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create push mirror err: %v", err))
	}

	return to.TextResult(slimMirror(mirror))
}

func deletePushMirrorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Mirror] Called deletePushMirrorFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	remoteName, err := params.GetString(args, "remote_name")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, err = client.DeletePushMirror(owner, repo, remoteName)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete push mirror err: %v", err))
	}

	return to.TextResult("Push mirror deleted successfully")
}

func syncMirrorFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Mirror] Called syncMirrorFn")
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

	_, err = client.MirrorSync(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("sync mirror err: %v", err))
	}

	return to.TextResult("Mirror sync triggered successfully")
}

func slimMirrors(mirrors []*gitea_sdk.PushMirrorResponse) []map[string]interface{} {
	result := make([]map[string]interface{}, len(mirrors))
	for i, m := range mirrors {
		result[i] = slimMirror(m)
	}
	return result
}

func slimMirror(m *gitea_sdk.PushMirrorResponse) map[string]interface{} {
	return map[string]interface{}{
		"remote_name":    m.RemoteName,
		"remote_addr":    m.RemoteAddress,
		"interval":       m.Interval,
		"sync_on_commit": m.SyncONCommit,
		"last_update":    m.LastUpdate,
		"last_error":     m.LastError,
	}
}
