package activity

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
	ListUserActivityToolName = "list_user_activity"
	ListOrgActivityToolName  = "list_org_activity"
	ListTeamActivityToolName = "list_team_activity"
)

var Tool = tool.New()

var (
	ListUserActivityTool = mcp.NewTool(
		ListUserActivityToolName,
		mcp.WithDescription("List activity/feeds for a user"),
		mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
		mcp.WithString("date", mcp.Description("Filter by date (YYYY-MM-DD)")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(20)),
	)

	ListOrgActivityTool = mcp.NewTool(
		ListOrgActivityToolName,
		mcp.WithDescription("List activity/feeds for an organization"),
		mcp.WithString("org", mcp.Required(), mcp.Description("Organization name")),
		mcp.WithString("date", mcp.Description("Filter by date (YYYY-MM-DD)")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(20)),
	)

	ListTeamActivityTool = mcp.NewTool(
		ListTeamActivityToolName,
		mcp.WithDescription("List activity/feeds for a team"),
		mcp.WithNumber("team_id", mcp.Required(), mcp.Description("Team ID")),
		mcp.WithString("date", mcp.Description("Filter by date (YYYY-MM-DD)")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(20)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListUserActivityTool,
		Handler: listUserActivityFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListOrgActivityTool,
		Handler: listOrgActivityFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListTeamActivityTool,
		Handler: listTeamActivityFn,
	})
}

func listUserActivityFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Activity] Called listUserActivityFn")
	args := req.GetArguments()
	username, err := params.GetString(args, "username")
	if err != nil {
		return to.ErrorResult(err)
	}
	date := params.GetOptionalString(args, "date", "")
	page, pageSize := params.GetPagination(args, 20)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListUserActivityFeedsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
		Date: date,
	}

	activities, _, err := client.ListUserActivityFeeds(username, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list user activity err: %v", err))
	}

	return to.TextResult(slimActivities(activities))
}

func listOrgActivityFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Activity] Called listOrgActivityFn")
	args := req.GetArguments()
	org, err := params.GetString(args, "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	date := params.GetOptionalString(args, "date", "")
	page, pageSize := params.GetPagination(args, 20)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListOrgActivityFeedsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
		Date: date,
	}

	activities, _, err := client.ListOrgActivityFeeds(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list org activity err: %v", err))
	}

	return to.TextResult(slimActivities(activities))
}

func listTeamActivityFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Activity] Called listTeamActivityFn")
	args := req.GetArguments()
	teamID, err := params.GetIndex(args, "team_id")
	if err != nil {
		return to.ErrorResult(err)
	}
	date := params.GetOptionalString(args, "date", "")
	page, pageSize := params.GetPagination(args, 20)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListTeamActivityFeedsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
		Date: date,
	}

	activities, _, err := client.ListTeamActivityFeeds(teamID, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list team activity err: %v", err))
	}

	return to.TextResult(slimActivities(activities))
}

func slimActivities(activities []*gitea_sdk.Activity) []map[string]interface{} {
	result := make([]map[string]interface{}, len(activities))
	for i, a := range activities {
		result[i] = slimActivity(a)
	}
	return result
}

func slimActivity(a *gitea_sdk.Activity) map[string]interface{} {
	result := map[string]interface{}{
		"id":         a.ID,
		"op_type":    a.OpType,
		"content":    a.Content,
		"repo_id":    a.RepoID,
		"comment_id": a.CommentID,
		"ref_name":   a.RefName,
		"is_private": a.IsPrivate,
		"user_id":    a.UserID,
		"created":    a.Created,
	}
	if a.Repo != nil {
		result["repo_name"] = a.Repo.FullName
	}
	if a.ActUser != nil {
		result["act_user_id"] = a.ActUserID
		result["act_user_name"] = a.ActUser.UserName
	}
	return result
}
