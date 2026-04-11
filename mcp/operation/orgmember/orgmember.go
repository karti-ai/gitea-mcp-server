package orgmember

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
	CheckOrgMembershipToolName     = "check_org_membership"
	SetPublicOrgMembershipToolName = "set_public_org_membership"
	ListUserOrgsToolName           = "list_user_orgs_public"
)

var Tool = tool.New()

var (
	CheckOrgMembershipTool = mcp.NewTool(
		CheckOrgMembershipToolName,
		mcp.WithDescription("Check if a user is a member of an organization"),
		mcp.WithString("org", mcp.Required(), mcp.Description("Organization name")),
		mcp.WithString("user", mcp.Required(), mcp.Description("Username to check")),
	)

	SetPublicOrgMembershipTool = mcp.NewTool(
		SetPublicOrgMembershipToolName,
		mcp.WithDescription("Set public organization membership visibility"),
		mcp.WithString("org", mcp.Required(), mcp.Description("Organization name")),
		mcp.WithString("user", mcp.Required(), mcp.Description("Username")),
		mcp.WithBoolean("visible", mcp.Required(), mcp.Description("Make membership public (true) or private (false)")),
	)

	ListUserOrgsTool = mcp.NewTool(
		ListUserOrgsToolName,
		mcp.WithDescription("List organizations a user belongs to"),
		mcp.WithString("user", mcp.Required(), mcp.Description("Username")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30)),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    CheckOrgMembershipTool,
		Handler: checkOrgMembershipFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListUserOrgsTool,
		Handler: listUserOrgsFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    SetPublicOrgMembershipTool,
		Handler: setPublicOrgMembershipFn,
	})
}

func checkOrgMembershipFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[OrgMember] Called checkOrgMembershipFn")
	args := req.GetArguments()
	org, err := params.GetString(args, "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	user, err := params.GetString(args, "user")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	isMember, _, err := client.CheckOrgMembership(org, user)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("check org membership err: %v", err))
	}

	return to.TextResult(map[string]interface{}{
		"is_member": isMember,
	})
}

func setPublicOrgMembershipFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[OrgMember] Called setPublicOrgMembershipFn")
	args := req.GetArguments()
	org, err := params.GetString(args, "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	user, err := params.GetString(args, "user")
	if err != nil {
		return to.ErrorResult(err)
	}
	visible, _ := args["visible"].(bool)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, err = client.SetPublicOrgMembership(org, user, visible)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("set public org membership err: %v", err))
	}

	return to.TextResult(fmt.Sprintf("Public org membership %s for %s", map[bool]string{true: "enabled", false: "disabled"}[visible], user))
}

func listUserOrgsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[OrgMember] Called listUserOrgsFn")
	args := req.GetArguments()
	user, err := params.GetString(args, "user")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(args, 30)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListOrgsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}

	orgs, _, err := client.ListUserOrgs(user, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list user orgs err: %v", err))
	}

	return to.TextResult(slimOrgs(orgs))
}

func slimOrgs(orgs []*gitea_sdk.Organization) []map[string]interface{} {
	result := make([]map[string]interface{}, len(orgs))
	for i, o := range orgs {
		result[i] = map[string]interface{}{
			"id":          o.ID,
			"username":    o.UserName,
			"full_name":   o.FullName,
			"description": o.Description,
			"website":     o.Website,
		}
	}
	return result
}
