package org

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
	OrgReadToolName  = "org_read"
	OrgWriteToolName = "org_write"
)

var (
	OrgReadTool = mcp.NewTool(
		OrgReadToolName,
		mcp.WithDescription("Read organization information. Use method 'get' to get org details, 'list' to list orgs, 'list_members' to list org members, 'list_teams' to list org teams."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("get", "list", "list_members", "list_teams")),
		mcp.WithString("org", mcp.Description("organization name (required for 'get', 'list_members', 'list_teams')")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30)),
	)

	OrgWriteTool = mcp.NewTool(
		OrgWriteToolName,
		mcp.WithDescription("Create or update organizations."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("create", "update", "delete")),
		mcp.WithString("org", mcp.Description("organization name (required for 'update', 'delete')")),
		mcp.WithString("description", mcp.Description("organization description")),
		mcp.WithString("full_name", mcp.Description("full name")),
		mcp.WithString("location", mcp.Description("location")),
		mcp.WithString("website", mcp.Description("website")),
		mcp.WithString("visibility", mcp.Description("visibility"), mcp.Enum("public", "private", "limited")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    OrgReadTool,
		Handler: orgReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    OrgWriteTool,
		Handler: orgWriteFn,
	})
}

func orgReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "get":
		return getOrgFn(ctx, req)
	case "list":
		return listOrgsFn(ctx, req)
	case "list_members":
		return listOrgMembersFn(ctx, req)
	case "list_teams":
		return listOrgTeamsFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func orgWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "create":
		return createOrgFn(ctx, req)
	case "update":
		return updateOrgFn(ctx, req)
	case "delete":
		return deleteOrgFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func getOrgFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getOrgFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	orgData, _, err := client.GetOrg(org)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get org err: %v", err))
	}
	return to.TextResult(slimOrg(orgData))
}

func listOrgsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listOrgsFn")
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	opt := gitea_sdk.ListOrgsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	orgs, _, err := client.ListMyOrgs(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list orgs err: %v", err))
	}
	return to.TextResult(slimOrgs(orgs))
}

func listOrgMembersFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listOrgMembersFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	opt := gitea_sdk.ListOrgMembershipOption{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	members, _, err := client.ListOrgMembership(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list org members err: %v", err))
	}
	return to.TextResult(slimUsers(members))
}

func listOrgTeamsFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listOrgTeamsFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	opt := gitea_sdk.ListTeamsOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	teams, _, err := client.ListOrgTeams(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list org teams err: %v", err))
	}
	return to.TextResult(slimTeams(teams))
}

func createOrgFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called createOrgFn")
	name, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	description, _ := req.GetArguments()["description"].(string)
	fullName, _ := req.GetArguments()["full_name"].(string)
	location, _ := req.GetArguments()["location"].(string)
	website, _ := req.GetArguments()["website"].(string)
	visibility, _ := req.GetArguments()["visibility"].(string)

	opt := gitea_sdk.CreateOrgOption{
		Name:        name,
		Description: description,
		FullName:    fullName,
		Location:    location,
		Website:     website,
		Visibility:  gitea_sdk.VisibleType(visibility),
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	org, _, err := client.CreateOrg(opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create org err: %v", err))
	}
	return to.TextResult(slimOrg(org))
}

func updateOrgFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called updateOrgFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditOrgOption{}
	if description, ok := req.GetArguments()["description"].(string); ok {
		opt.Description = description
	}
	if fullName, ok := req.GetArguments()["full_name"].(string); ok {
		opt.FullName = fullName
	}
	if location, ok := req.GetArguments()["location"].(string); ok {
		opt.Location = location
	}
	if website, ok := req.GetArguments()["website"].(string); ok {
		opt.Website = website
	}
	if visibility, ok := req.GetArguments()["visibility"].(string); ok {
		opt.Visibility = gitea_sdk.VisibleType(visibility)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.EditOrg(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("update org err: %v", err))
	}
	return to.TextResult("Organization updated successfully")
}

func deleteOrgFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called deleteOrgFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteOrg(org)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete org err: %v", err))
	}
	return to.TextResult("Organization deleted successfully")
}

func slimOrg(o *gitea_sdk.Organization) map[string]any {
	if o == nil {
		return nil
	}
	return map[string]any{
		"id":          o.ID,
		"name":        o.Name,
		"full_name":   o.FullName,
		"description": o.Description,
		"avatar_url":  o.AvatarURL,
		"website":     o.Website,
		"location":    o.Location,
		"visibility":  o.Visibility,
	}
}

func slimOrgs(orgs []*gitea_sdk.Organization) []map[string]any {
	out := make([]map[string]any, 0, len(orgs))
	for _, o := range orgs {
		out = append(out, slimOrg(o))
	}
	return out
}

func slimUsers(users []*gitea_sdk.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{
			"id":         u.ID,
			"login":      u.UserName,
			"full_name":  u.FullName,
			"email":      u.Email,
			"avatar_url": u.AvatarURL,
		})
	}
	return out
}

func slimTeams(teams []*gitea_sdk.Team) []map[string]any {
	out := make([]map[string]any, 0, len(teams))
	for _, t := range teams {
		out = append(out, map[string]any{
			"id":          t.ID,
			"name":        t.Name,
			"description": t.Description,
			"permission":  t.Permission,
		})
	}
	return out
}
