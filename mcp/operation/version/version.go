package version

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitea.com/gitea/gitea-mcp/pkg/flag"
	"gitea.com/gitea/gitea-mcp/pkg/gitea"
	"gitea.com/gitea/gitea-mcp/pkg/log"
	"gitea.com/gitea/gitea-mcp/pkg/to"
	"gitea.com/gitea/gitea-mcp/pkg/tool"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var Tool = tool.New()

const (
	GetGiteaMCPServerVersion = "get_gitea_mcp_server_version"
	CheckGiteaVersion        = "check_gitea_version"
)

var GetGiteaMCPServerVersionTool = mcp.NewTool(
	GetGiteaMCPServerVersion,
	mcp.WithDescription("Get Gitea MCP Server Version"),
)

var CheckGiteaVersionTool = mcp.NewTool(
	CheckGiteaVersion,
	mcp.WithDescription("Check the Gitea server version and API capabilities. Returns version string, parsed components, and capability matrix indicating which APIs are available based on the detected version."),
)

// VersionResponse represents the Gitea version API response
type VersionResponse struct {
	Version string `json:"version"`
}

// VersionInfo represents the complete version information with capabilities
type VersionInfo struct {
	Version    string           `json:"version"`
	Major      int              `json:"major"`
	Minor      int              `json:"minor"`
	Patch      int              `json:"patch"`
	Capabilities CapabilityMatrix `json:"capabilities"`
}

// CapabilityMatrix indicates which APIs are available
type CapabilityMatrix struct {
	ActionsAPI     bool `json:"actions_api"`
	SecretsAPI     bool `json:"secrets_api"`
	VariablesAPI   bool `json:"variables_api"`
	RunnersAPI     bool `json:"runners_api"`
	ArtifactsAPI   bool `json:"artifacts_api"`
	CommitStatusAPI bool `json:"commit_status_api"`
}

func init() {
	Tool.RegisterRead(server.ServerTool{
		Tool:    GetGiteaMCPServerVersionTool,
		Handler: GetGiteaMCPServerVersionFn,
	})
	Tool.RegisterRead(server.ServerTool{
		Tool:    CheckGiteaVersionTool,
		Handler: CheckGiteaVersionFn,
	})
}

func GetGiteaMCPServerVersionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called GetGiteaMCPServerVersionFn")
	version := flag.Version
	if version == "" {
		version = "dev"
	}
	return to.TextResult(fmt.Sprintf("Gitea MCP Server version: %v", version))
}

// CheckGiteaVersionFn retrieves Gitea server version and determines capabilities
func CheckGiteaVersionFn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debugf("Called CheckGiteaVersionFn")

	var versionResp VersionResponse
	status, err := gitea.DoJSON(ctx, "GET", "version", nil, nil, &versionResp)
	if err != nil {
		log.Errorf("Failed to get Gitea version: status=%d, err=%v", status, err)
		return to.TextResult(fmt.Sprintf("Error: Failed to get Gitea version (status %d): %v", status, err))
	}

	major, minor, patch, err := parseVersion(versionResp.Version)
	if err != nil {
		log.Errorf("Failed to parse version string '%s': %v", versionResp.Version, err)
		return to.TextResult(fmt.Sprintf("Error: Failed to parse version '%s': %v", versionResp.Version, err))
	}

	capabilities := determineCapabilities(major, minor, patch)

	info := VersionInfo{
		Version:      versionResp.Version,
		Major:        major,
		Minor:        minor,
		Patch:        patch,
		Capabilities: capabilities,
	}

	jsonBytes, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Errorf("Failed to marshal version info: %v", err)
		return to.TextResult(fmt.Sprintf("Error: Failed to format response: %v", err))
	}

	return to.TextResult(string(jsonBytes))
}

// parseVersion parses a version string like "1.22.5" into major, minor, patch
func parseVersion(version string) (int, int, int, error) {
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s (expected major.minor.patch)", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch := 0
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return major, minor, patch, nil
}

// determineCapabilities determines which APIs are available based on version
func determineCapabilities(major, minor, patch int) CapabilityMatrix {
	cm := CapabilityMatrix{}

	// Commit Status API: Available in 1.12+
	if major > 1 || (major == 1 && minor >= 12) {
		cm.CommitStatusAPI = true
	}

	// Secrets API: Available in 1.22.0+
	if major > 1 || (major == 1 && minor >= 22) {
		cm.SecretsAPI = true
	}

	// Variables API: Available in 1.22.0+ (same as secrets)
	if major > 1 || (major == 1 && minor >= 22) {
		cm.VariablesAPI = true
	}

	// Runners API: Available in 1.22.0+ with improvements in 1.23+
	if major > 1 || (major == 1 && minor >= 22) {
		cm.RunnersAPI = true
	}

	// Actions API: Full support in 1.23+, limited in 1.22
	// Artifacts API: Full support in 1.23+
	if major > 1 || (major == 1 && minor >= 23) {
		cm.ActionsAPI = true
		cm.ArtifactsAPI = true
		// Runners API is more complete in 1.23+
		cm.RunnersAPI = true
	}

	return cm
}
