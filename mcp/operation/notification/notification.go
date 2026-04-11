package notification

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
	NotificationReadToolName  = "notification_read"
	NotificationWriteToolName = "notification_write"
)

var (
	NotificationReadTool = mcp.NewTool(
		NotificationReadToolName,
		mcp.WithDescription("Read notifications. Use method 'list' to list all notifications, 'list_repo' for repo-specific notifications, 'check' to get unread count."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("list", "list_repo", "check")),
		mcp.WithString("owner", mcp.Description("repository owner (required for 'list_repo')")),
		mcp.WithString("repo", mcp.Description("repository name (required for 'list_repo')")),
		mcp.WithString("status", mcp.Description("status filter"), mcp.Enum("unread", "read", "pinned")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30)),
	)

	NotificationWriteTool = mcp.NewTool(
		NotificationWriteToolName,
		mcp.WithDescription("Mark notifications as read."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("read", "read_repo")),
		mcp.WithString("owner", mcp.Description("repository owner (required for 'read_repo')")),
		mcp.WithString("repo", mcp.Description("repository name (required for 'read_repo')")),
		mcp.WithNumber("id", mcp.Description("notification ID (optional, marks single if provided)")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    NotificationReadTool,
		Handler: notificationReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    NotificationWriteTool,
		Handler: notificationWriteFn,
	})
}

func notificationReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "list":
		return listNotificationsFn(ctx, req)
	case "list_repo":
		return listRepoNotificationsFn(ctx, req)
	case "check":
		return checkNotificationsFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func notificationWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "read":
		return readNotificationsFn(ctx, req)
	case "read_repo":
		return readRepoNotificationsFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func listNotificationsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listNotificationsFn")
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	status, _ := req.GetArguments()["status"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListNotificationOptions{
		Status: []gitea_sdk.NotifyStatus{gitea_sdk.NotifyStatus(status)},
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	notifications, _, err := client.ListNotifications(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list notifications err: %v", err))
	}
	return to.TextResult(slimNotifications(notifications))
}

func listRepoNotificationsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listRepoNotificationsFn")
	owner, err := params.GetString(req.GetArguments(), "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(req.GetArguments(), "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	status, _ := req.GetArguments()["status"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListNotificationOptions{
		Status: []gitea_sdk.NotifyStatus{gitea_sdk.NotifyStatus(status)},
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	notifications, _, err := client.ListRepoNotifications(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list repo notifications err: %v", err))
	}
	return to.TextResult(slimNotifications(notifications))
}

func checkNotificationsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called checkNotificationsFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	count, _, err := client.CheckNotifications()
	if err != nil {
		return to.ErrorResult(fmt.Errorf("check notifications err: %v", err))
	}
	return to.TextResult(map[string]any{"unread_count": count})
}

func readNotificationsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called readNotificationsFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	id, hasID := req.GetArguments()["id"].(float64)
	if hasID && id > 0 {
		_, _, err := client.ReadNotification(int64(id), gitea_sdk.NotifyStatusRead)
		if err != nil {
			return to.ErrorResult(fmt.Errorf("read notification err: %v", err))
		}
		return to.TextResult("Notification marked as read")
	}

	opt := gitea_sdk.MarkNotificationOptions{Status: []gitea_sdk.NotifyStatus{gitea_sdk.NotifyStatusRead}}
	_, _, err = client.ReadNotifications(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("read notifications err: %v", err))
	}
	return to.TextResult("All notifications marked as read")
}

func readRepoNotificationsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called readRepoNotificationsFn")
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
	opt := gitea_sdk.MarkNotificationOptions{Status: []gitea_sdk.NotifyStatus{gitea_sdk.NotifyStatusRead}}
	_, _, err = client.ReadRepoNotifications(owner, repo, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("read repo notifications err: %v", err))
	}
	return to.TextResult("Repository notifications marked as read")
}

func slimNotifications(notifications []*gitea_sdk.NotificationThread) []map[string]any {
	out := make([]map[string]any, 0, len(notifications))
	for _, n := range notifications {
		out = append(out, map[string]any{
			"id":         n.ID,
			"unread":     n.Unread,
			"subject":    n.Subject.Title,
			"type":       n.Subject.Type,
			"url":        n.Subject.URL,
			"repository": n.Repository.FullName,
		})
	}
	return out
}
