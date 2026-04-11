package team

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
	TeamReadToolName  = "team_read"
	TeamWriteToolName = "team_write"
)

var (
	TeamReadTool = mcp.NewTool(
		TeamReadToolName,
		mcp.WithDescription("Read team information. Use method 'get' to get team details, 'list_members' to list team members, 'list_repos' to list team repositories."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("get", "list_members", "list_repos")),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("id", mcp.Description("team ID (required for 'get', 'list_members', 'list_repos')")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(30)),
	)

	TeamWriteTool = mcp.NewTool(
		TeamWriteToolName,
		mcp.WithDescription("Create, update, or delete teams, manage team members."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("create", "edit", "delete", "add_member", "remove_member", "add_repo", "remove_repo")),
		mcp.WithString("org", mcp.Required(), mcp.Description("organization name")),
		mcp.WithNumber("id", mcp.Description("team ID (required for 'edit', 'delete', 'add_member', 'remove_member', 'add_repo', 'remove_repo')")),
		mcp.WithString("name", mcp.Description("team name (required for 'create', optional for 'edit')")),
		mcp.WithString("description", mcp.Description("team description")),
		mcp.WithString("permission", mcp.Description("permission level"), mcp.Enum("read", "write", "admin", "owner")),
		mcp.WithString("user", mcp.Description("username (required for 'add_member', 'remove_member')")),
		mcp.WithString("repo", mcp.Description("repository name (required for 'add_repo', 'remove_repo')")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    TeamReadTool,
		Handler: teamReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    TeamWriteTool,
		Handler: teamWriteFn,
	})
}

func teamReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "get":
		return getTeamFn(ctx, req)
	case "list_members":
		return listTeamMembersFn(ctx, req)
	case "list_repos":
		return listTeamReposFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func teamWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "create":
		return createTeamFn(ctx, req)
	case "edit":
		return editTeamFn(ctx, req)
	case "delete":
		return deleteTeamFn(ctx, req)
	case "add_member":
		return addTeamMemberFn(ctx, req)
	case "remove_member":
		return removeTeamMemberFn(ctx, req)
	case "add_repo":
		return addTeamRepoFn(ctx, req)
	case "remove_repo":
		return removeTeamRepoFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func getTeamFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called getTeamFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	team, _, err := client.GetTeam(id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get team err: %v", err))
	}
	return to.TextResult(slimTeam(team))
}

func listTeamMembersFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listTeamMembersFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	opt := gitea_sdk.ListTeamMembersOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	members, _, err := client.ListTeamMembers(id, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list team members err: %v", err))
	}
	return to.TextResult(slimUsers(members))
}

func listTeamReposFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listTeamReposFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(req.GetArguments(), 30)
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	opt := gitea_sdk.ListTeamRepositoriesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}
	repos, _, err := client.ListTeamRepositories(id, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list team repos err: %v", err))
	}
	return to.TextResult(slimRepos(repos))
}

func createTeamFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called createTeamFn")
	org, err := params.GetString(req.GetArguments(), "org")
	if err != nil {
		return to.ErrorResult(err)
	}
	name, err := params.GetString(req.GetArguments(), "name")
	if err != nil {
		return to.ErrorResult(err)
	}
	description, _ := req.GetArguments()["description"].(string)
	permission, _ := req.GetArguments()["permission"].(string)

	opt := gitea_sdk.CreateTeamOption{
		Name:        name,
		Description: description,
		Permission:  gitea_sdk.AccessMode(permission),
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	team, _, err := client.CreateTeam(org, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("create team err: %v", err))
	}
	return to.TextResult(slimTeam(team))
}

func editTeamFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called editTeamFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}

	opt := gitea_sdk.EditTeamOption{}
	if name, ok := req.GetArguments()["name"].(string); ok && name != "" {
		opt.Name = name
	}
	if description, ok := req.GetArguments()["description"].(string); ok && description != "" {
		opt.Description = &description
	}
	if permission, ok := req.GetArguments()["permission"].(string); ok {
		opt.Permission = gitea_sdk.AccessMode(permission)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.EditTeam(id, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("edit team err: %v", err))
	}
	return to.TextResult("Team updated successfully")
}

func deleteTeamFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called deleteTeamFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.DeleteTeam(id)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete team err: %v", err))
	}
	return to.TextResult("Team deleted successfully")
}

func addTeamMemberFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called addTeamMemberFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	user, err := params.GetString(req.GetArguments(), "user")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.AddTeamMember(id, user)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add team member err: %v", err))
	}
	return to.TextResult("Team member added successfully")
}

func removeTeamMemberFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called removeTeamMemberFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	user, err := params.GetString(req.GetArguments(), "user")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	_, err = client.RemoveTeamMember(id, user)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("remove team member err: %v", err))
	}
	return to.TextResult("Team member removed successfully")
}

func addTeamRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called addTeamRepoFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	org, err := params.GetString(req.GetArguments(), "org")
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
	_, err = client.AddTeamRepository(id, org, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("add team repo err: %v", err))
	}
	return to.TextResult("Team repository added successfully")
}

func removeTeamRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called removeTeamRepoFn")
	id, err := params.GetIndex(req.GetArguments(), "id")
	if err != nil {
		return to.ErrorResult(err)
	}
	org, err := params.GetString(req.GetArguments(), "org")
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
	_, err = client.RemoveTeamRepository(id, org, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("remove team repo err: %v", err))
	}
	return to.TextResult("Team repository removed successfully")
}

func slimTeam(t *gitea_sdk.Team) map[string]any {
	if t == nil {
		return nil
	}
	orgName := ""
	if t.Organization != nil {
		orgName = t.Organization.Name
	}
	return map[string]any{
		"id":          t.ID,
		"name":        t.Name,
		"description": t.Description,
		"permission":  t.Permission,
		"org_name":    orgName,
	}
}

func slimUsers(users []*gitea_sdk.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{
			"id":         u.ID,
			"login":      u.UserName,
			"full_name":  u.FullName,
			"avatar_url": u.AvatarURL,
		})
	}
	return out
}

func slimRepos(repos []*gitea_sdk.Repository) []map[string]any {
	out := make([]map[string]any, 0, len(repos))
	for _, r := range repos {
		out = append(out, map[string]any{
			"id":        r.ID,
			"name":      r.Name,
			"full_name": r.FullName,
		})
	}
	return out
}
