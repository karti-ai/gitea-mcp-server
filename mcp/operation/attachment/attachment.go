package attachment

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
	ListReleaseAttachmentsToolName      = "list_release_attachments"
	GetReleaseAttachmentToolName        = "get_release_attachment"
	ListIssueCommentAttachmentsToolName = "list_issue_comment_attachments"
	GetIssueCommentAttachmentToolName   = "get_issue_comment_attachment"
)

var Tool = tool.New()

var (
	ListReleaseAttachmentsTool = mcp.NewTool(
		ListReleaseAttachmentsToolName,
		mcp.WithDescription("List attachments for a release"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("release", mcp.Required(), mcp.Description("Release ID")),
	)

	GetReleaseAttachmentTool = mcp.NewTool(
		GetReleaseAttachmentToolName,
		mcp.WithDescription("Get a specific release attachment"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("release", mcp.Required(), mcp.Description("Release ID")),
		mcp.WithNumber("attachment", mcp.Required(), mcp.Description("Attachment ID")),
	)

	ListIssueCommentAttachmentsTool = mcp.NewTool(
		ListIssueCommentAttachmentsToolName,
		mcp.WithDescription("List attachments for an issue comment"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("comment", mcp.Required(), mcp.Description("Comment ID")),
	)

	GetIssueCommentAttachmentTool = mcp.NewTool(
		GetIssueCommentAttachmentToolName,
		mcp.WithDescription("Get a specific issue comment attachment"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithNumber("comment", mcp.Required(), mcp.Description("Comment ID")),
		mcp.WithNumber("attachment", mcp.Required(), mcp.Description("Attachment ID")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListReleaseAttachmentsTool,
		Handler: listReleaseAttachmentsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetReleaseAttachmentTool,
		Handler: getReleaseAttachmentFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListIssueCommentAttachmentsTool,
		Handler: listIssueCommentAttachmentsFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetIssueCommentAttachmentTool,
		Handler: getIssueCommentAttachmentFn,
	})
}

func listReleaseAttachmentsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Attachment] Called listReleaseAttachmentsFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	release, err := params.GetIndex(args, "release")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	attachments, _, err := client.ListReleaseAttachments(owner, repo, release, gitea_sdk.ListReleaseAttachmentsOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list release attachments err: %v", err))
	}

	return to.TextResult(slimAttachments(attachments))
}

func getReleaseAttachmentFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Attachment] Called getReleaseAttachmentFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	release, err := params.GetIndex(args, "release")
	if err != nil {
		return to.ErrorResult(err)
	}
	attachment, err := params.GetIndex(args, "attachment")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	att, _, err := client.GetReleaseAttachment(owner, repo, release, attachment)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get release attachment err: %v", err))
	}

	return to.TextResult(slimAttachment(att))
}

func listIssueCommentAttachmentsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Attachment] Called listIssueCommentAttachmentsFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	comment, err := params.GetIndex(args, "comment")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	attachments, _, err := client.ListIssueCommentAttachments(owner, repo, comment)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list issue comment attachments err: %v", err))
	}

	return to.TextResult(slimAttachments(attachments))
}

func getIssueCommentAttachmentFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Attachment] Called getIssueCommentAttachmentFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	repo, err := params.GetString(args, "repo")
	if err != nil {
		return to.ErrorResult(err)
	}
	comment, err := params.GetIndex(args, "comment")
	if err != nil {
		return to.ErrorResult(err)
	}
	attachment, err := params.GetIndex(args, "attachment")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	att, _, err := client.GetIssueCommentAttachment(owner, repo, comment, attachment)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get issue comment attachment err: %v", err))
	}

	return to.TextResult(slimAttachment(att))
}

func slimAttachments(atts []*gitea_sdk.Attachment) []map[string]interface{} {
	result := make([]map[string]interface{}, len(atts))
	for i, a := range atts {
		result[i] = slimAttachment(a)
	}
	return result
}

func slimAttachment(a *gitea_sdk.Attachment) map[string]interface{} {
	return map[string]interface{}{
		"id":             a.ID,
		"name":           a.Name,
		"size":           a.Size,
		"download_count": a.DownloadCount,
		"download_url":   a.DownloadURL,
		"uuid":           a.UUID,
		"created":        a.Created,
	}
}
