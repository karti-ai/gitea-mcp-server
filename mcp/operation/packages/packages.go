package packages

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
	ListPackagesToolName     = "list_packages"
	GetPackageToolName       = "get_package"
	GetLatestPackageToolName = "get_latest_package"
	DeletePackageToolName    = "delete_package"
	ListPackageFilesToolName = "list_package_files"
)

var Tool = tool.New()

var (
	ListPackagesTool = mcp.NewTool(
		ListPackagesToolName,
		mcp.WithDescription("List packages for a user/org"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Owner (user or org) name")),
		mcp.WithString("type", mcp.Description("Package type (container, npm, pypi, etc.)")),
		mcp.WithString("name", mcp.Description("Filter by package name")),
		mcp.WithNumber("page", mcp.Description("page number"), mcp.DefaultNumber(1)),
		mcp.WithNumber("perPage", mcp.Description("results per page"), mcp.DefaultNumber(20)),
	)

	GetPackageTool = mcp.NewTool(
		GetPackageToolName,
		mcp.WithDescription("Get a specific package version"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Owner name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Package type")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
		mcp.WithString("version", mcp.Required(), mcp.Description("Package version")),
	)

	GetLatestPackageTool = mcp.NewTool(
		GetLatestPackageToolName,
		mcp.WithDescription("Get the latest version of a package"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Owner name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Package type")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
	)

	DeletePackageTool = mcp.NewTool(
		DeletePackageToolName,
		mcp.WithDescription("Delete a package or specific version"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Owner name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Package type")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
		mcp.WithString("version", mcp.Description("Specific version to delete (omit to delete all)")),
	)

	ListPackageFilesTool = mcp.NewTool(
		ListPackageFilesToolName,
		mcp.WithDescription("List files in a package"),
		mcp.WithString("owner", mcp.Required(), mcp.Description("Owner name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Package type")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Package name")),
		mcp.WithString("version", mcp.Required(), mcp.Description("Package version")),
	)
)

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListPackagesTool,
		Handler: listPackagesFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetPackageTool,
		Handler: getPackageFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetLatestPackageTool,
		Handler: getLatestPackageFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    ListPackageFilesTool,
		Handler: listPackageFilesFn,
	})
	Tool.RegisterWrite(server.ServerTool{
		Tool:    DeletePackageTool,
		Handler: deletePackageFn,
	})
}

func listPackagesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Packages] Called listPackagesFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	page, pageSize := params.GetPagination(args, 20)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	opt := gitea_sdk.ListPackagesOptions{
		ListOptions: gitea_sdk.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
	}

	pkgs, _, err := client.ListPackages(owner, opt)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list packages err: %v", err))
	}

	return to.TextResult(slimPackages(pkgs))
}

func getPackageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Packages] Called getPackageFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	packageType, err := params.GetString(args, "type")
	if err != nil {
		return to.ErrorResult(err)
	}
	name, err := params.GetString(args, "name")
	if err != nil {
		return to.ErrorResult(err)
	}
	version, err := params.GetString(args, "version")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	pkg, _, err := client.GetPackage(owner, packageType, name, version)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get package err: %v", err))
	}

	return to.TextResult(slimPackage(pkg))
}

func getLatestPackageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Packages] Called getLatestPackageFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	packageType, err := params.GetString(args, "type")
	if err != nil {
		return to.ErrorResult(err)
	}
	name, err := params.GetString(args, "name")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	pkg, _, err := client.GetLatestPackage(owner, packageType, name)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get latest package err: %v", err))
	}

	return to.TextResult(slimPackage(pkg))
}

func deletePackageFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Packages] Called deletePackageFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	packageType, err := params.GetString(args, "type")
	if err != nil {
		return to.ErrorResult(err)
	}
	name, err := params.GetString(args, "name")
	if err != nil {
		return to.ErrorResult(err)
	}
	version, _ := args["version"].(string)

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	_, err = client.DeletePackage(owner, packageType, name, version)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("delete package err: %v", err))
	}

	return to.TextResult("Package deleted successfully")
}

func listPackageFilesFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("[Packages] Called listPackageFilesFn")
	args := req.GetArguments()
	owner, err := params.GetString(args, "owner")
	if err != nil {
		return to.ErrorResult(err)
	}
	packageType, err := params.GetString(args, "type")
	if err != nil {
		return to.ErrorResult(err)
	}
	name, err := params.GetString(args, "name")
	if err != nil {
		return to.ErrorResult(err)
	}
	version, err := params.GetString(args, "version")
	if err != nil {
		return to.ErrorResult(err)
	}

	client, err := gitea.ClientFromContext(ctx)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("get gitea client err: %v", err))
	}

	files, _, err := client.ListPackageFiles(owner, packageType, name, version)
	if err != nil {
		return to.ErrorResult(fmt.Errorf("list package files err: %v", err))
	}

	return to.TextResult(slimPackageFiles(files))
}

func slimPackages(pkgs []*gitea_sdk.Package) []map[string]interface{} {
	result := make([]map[string]interface{}, len(pkgs))
	for i, p := range pkgs {
		result[i] = slimPackage(p)
	}
	return result
}

func slimPackage(p *gitea_sdk.Package) map[string]interface{} {
	return map[string]interface{}{
		"id":           p.ID,
		"name":         p.Name,
		"version":      p.Version,
		"package_type": p.Type,
		"created_at":   p.CreatedAt,
		"owner":        p.Owner.UserName,
	}
}

func slimPackageFiles(files []*gitea_sdk.PackageFile) []map[string]interface{} {
	result := make([]map[string]interface{}, len(files))
	for i, f := range files {
		result[i] = map[string]interface{}{
			"id":     f.ID,
			"name":   f.Name,
			"size":   f.Size,
			"md5":    f.MD5,
			"sha256": f.SHA256,
			"sha512": f.SHA512,
		}
	}
	return result
}
