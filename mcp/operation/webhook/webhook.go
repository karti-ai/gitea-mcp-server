package webhook

import (
	"context"
	"encoding/json"
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
	WebhookReadToolName  = "webhook_read"
	WebhookWriteToolName = "webhook_write"
)

var (
	WebhookReadTool = mcp.NewTool(
		WebhookReadToolName,
		mcp.WithDescription("Read webhooks. Use method 'list_repo' to list repo webhooks, 'list_org' for org webhooks, 'get' to get a specific webhook."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("list_repo", "list_org", "get")),
		mcp.WithString("owner", mcp.Description("repository or organization owner (required for 'list_repo', 'get')")),
		mcp.WithString("repo", mcp.Description("repository name (required for 'list_repo', 'get')")),
		mcp.WithNumber("id", mcp.Description("webhook ID (required for 'get')")),
	)

	WebhookWriteTool = mcp.NewTool(
		WebhookWriteToolName,
		mcp.WithDescription("Create, update, or delete webhooks."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("create_repo", "create_org", "edit_repo", "edit_org", "delete")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository or organization owner")),
		mcp.WithString("repo", mcp.Description("repository name (required for repo operations)")),
		mcp.WithNumber("id", mcp.Description("webhook ID (required for 'edit', 'delete')")),
		mcp.WithString("url", mcp.Required(), mcp.Description("webhook URL")),
		mcp.WithString("secret", mcp.Description("webhook secret")),
		mcp.WithBoolean("active", mcp.Description("whether webhook is active")),
		mcp.WithArray("events", mcp.Description("events to trigger webhook"), mcp.Items(map[string]any{"type": "string"})),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    WebhookReadTool,
		Handler: webhookReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    WebhookWriteTool,
		Handler: webhookWriteFn,
	})
}

func webhookReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "list_repo":
		return listRepoHooksFn(ctx, req)
	case "list_org":
		return listOrgHooksFn(ctx, req)
	case "get":
		return getHookFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func webhookWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "create_repo":
		return createRepoHookFn(ctx, req)
	case "create_org":
		return createOrgHookFn(ctx, req)
	case "edit_repo":
		return editRepoHookFn(ctx, req)
	case "edit_org":
		return editOrgHookFn(ctx, req)
	case "delete":
		return deleteHookFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func listRepoHooksFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listRepoHooksFn")
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
	hooks, _, err := client.ListRepoHooks(owner, repo, gitea_sdk.ListHooksOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo hooks err: %v", err))
	}
	return to.TextResult(slimHooks(hooks))
}

func listOrgHooksFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listOrgHooksFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	hooks, _, err := client.ListOrgHooks(owner, gitea_sdk.ListHooksOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list org hooks err: %v", err))
	}
	return to.TextResult(slimHooks(hooks))
}

func getHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, _ := req.GetArguments()["repo"].(string)
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	var hook *gitea_sdk.Hook
	if repo != "" {
		hook, _, err = client.GetRepoHook(owner, repo, id)
	} else {
		hook, _, err = client.GetOrgHook(owner, id)
	}
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get hook err: %v", err))
	}
	return to.TextResult(slimHook(hook))
}

func createRepoHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called createRepoHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	url, err := params.GetString(req.GetArguments(), "url")
	if err != nil {
		return to.ErrorResult(err)
	}
	return createHook(ctx, owner, repo, url, req.GetArguments(), false)
}

func createOrgHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called createOrgHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	url, err := params.GetString(req.GetArguments(), "url")
	if err != nil {
		return to.ErrorResult(err)
	}
	return createHook(ctx, owner, "", url, req.GetArguments(), true)
}

func createHook(ctx context.Context, owner, repo, url string, args map[string]any, isOrg bool) (*mcp.CallToolResult, error) {
	secret, _ := args["secret"].(string)
	active, _ := args["active"].(bool)
	events, _ := args["events"].([]any)

	eventStr := "push"
	if len(events) > 0 {
		eventsJson, _ := json.Marshal(events)
		eventStr = string(eventsJson)
	}

	opt := gitea_sdk.CreateHookOption{
		Type:   "gitea",
		Active: active,
		Events: []string{eventStr},
		Config: map[string]string{
			"url":          url,
			"content_type": "json",
			"secret":       secret,
		},
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	var hook *gitea_sdk.Hook
	if isOrg {
		hook, _, err = client.CreateOrgHook(owner, opt)
	} else {
		hook, _, err = client.CreateRepoHook(owner, repo, opt)
	}
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create hook err: %v", err))
	}
	return to.TextResult(slimHook(hook))
}

func editRepoHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called editRepoHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	return editHook(ctx, owner, repo, id, req.GetArguments(), false)
}

func editOrgHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called editOrgHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	return editHook(ctx, owner, "", id, req.GetArguments(), true)
}

func editHook(ctx context.Context, owner, repo string, id int64, args map[string]any, isOrg bool) (*mcp.CallToolResult, error) {
	url, _ := args["url"].(string)
	secret, _ := args["secret"].(string)
	active, hasActive := args["active"].(bool)

	opt := gitea_sdk.EditHookOption{}
	if url != "" {
		opt.Config = map[string]string{"url": url}
	}
	if secret != "" {
		if opt.Config == nil {
			opt.Config = map[string]string{}
		}
		opt.Config["secret"] = secret
	}
	if hasActive {
		opt.Active = &active
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	if isOrg {
		_, err = client.EditOrgHook(owner, id, opt)
	} else {
		_, err = client.EditRepoHook(owner, repo, id, opt)
	}
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit hook err: %v", err))
	}
	return to.TextResult("Webhook updated successfully")
}

func deleteHookFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called deleteHookFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, hasRepo := req.GetArguments()["repo"].(string)
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	if hasRepo && repo != "" {
		_, err = client.DeleteRepoHook(owner, repo, id)
	} else {
		_, err = client.DeleteOrgHook(owner, id)
	}
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete hook err: %v", err))
	}
	return to.TextResult("Webhook deleted successfully")
}

func slimHooks(hooks []*gitea_sdk.Hook) []map[string]any {
	out := make([]map[string]any, 0, len(hooks))
	for _, h := range hooks {
		out = append(out, slimHook(h))
	}
	return out
}

func slimHook(h *gitea_sdk.Hook) map[string]any {
	if h == nil {
		return nil
	}
	return map[string]any{
		"id":      h.ID,
		"type":    h.Type,
		"url":     h.URL,
		"active":  h.Active,
		"events":  h.Events,
		"created": h.Created,
	}
}
