package stars

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
	StarsReadToolName  = "stars_read"
	StarsWriteToolName = "stars_write"
)

var (
	StarsReadTool = mcp.NewTool(
		StarsReadToolName,
		mcp.WithDescription("Read stars information. Use method 'list_stargazers' to list repo stargazers, 'list_starred' to list user's starred repos, 'my_starred' for your starred repos, 'check' to check if user starred a repo."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("list_stargazers", "list_starred", "my_starred", "check")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
		mcp.WithString("user", mcp.Description("username (for 'list_starred', 'check')")),
	)

	StarsWriteTool = mcp.NewTool(
		StarsWriteToolName,
		mcp.WithDescription("Star or unstar a repository."),
		mcp.WithString("method", mcp.Required(), mcp.Description("operation to perform"), mcp.Enum("star", "unstar")),
		mcp.WithString("owner", mcp.Required(), mcp.Description("repository owner")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("repository name")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    StarsReadTool,
		Handler: starsReadFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    StarsWriteTool,
		Handler: starsWriteFn,
	})
}

func starsReadFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "list_stargazers":
		return listStargazersFn(ctx, req)
	case "list_starred":
		return listStarredFn(ctx, req)
	case "my_starred":
		return myStarredFn(ctx, req)
	case "check":
		return checkStarFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func starsWriteFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	method, err := params.GetString(req.GetArguments(), "method")
	if err != nil {
		return to.ErrorResult(err)
	}
	switch method {
	case "star":
		return starRepoFn(ctx, req)
	case "unstar":
		return unstarRepoFn(ctx, req)
	default:
		return to.ErrorResult(fmt.Errorf("unknown method: %s", method))
	}
}

func listStargazersFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listStargazersFn")
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
	users, _, err := client.ListRepoStargazers(owner, repo, gitea_sdk.ListStargazersOptions{})
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list stargazers err: %v", err))
	}
	return to.TextResult(slimUsers(users))
}

func listStarredFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called listStarredFn")
	user, err := params.GetString(req.GetArguments(), "user")
	if err != nil {
		return to.ErrorResult(err)
	}
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	repos, _, err := client.GetStarredRepos(user)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list starred repos err: %v", err))
	}
	return to.TextResult(slimRepos(repos))
}

func myStarredFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called myStarredFn")
	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}
	repos, _, err := client.GetMyStarredRepos()
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list my starred repos err: %v", err))
	}
	return to.TextResult(slimRepos(repos))
}

func checkStarFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called checkStarFn")
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
	starred, _, err := client.IsRepoStarring(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("check star err: %v", err))
	}
	return to.TextResult(map[string]any{"starred": starred})
}

func starRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called starRepoFn")
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
	_, err = client.StarRepo(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("star repo err: %v", err))
	}
	return to.TextResult("Repository starred successfully")
}

func unstarRepoFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called unstarRepoFn")
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
	_, err = client.UnStarRepo(owner, repo)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("unstar repo err: %v", err))
	}
	return to.TextResult("Repository unstarred successfully")
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
			"private":   r.Private,
		})
	}
	return out
}
