# GitCoffee MCP Setup for OpenClaw

This guide explains how to set up GitCoffee MCP with per-agent Gitea account isolation in OpenClaw.

## Structure

```
gitea-mcp-mcp/
├── mcp/           # MCP server source code (Go)
└── AGENTS.md      # This file - OpenClaw setup guide
```

## Architecture Overview

The GitCoffee MCP server provides Gitea tools via the Model Context Protocol (MCP). Each MCP instance uses a single Gitea token, making it ideal for per-agent isolation.

**Key Features:**
- stdio mode (default): For local MCP clients
- HTTP mode: For remote MCP servers on configurable ports
- 50+ tools: repos, issues, PRs, branches, files, wiki, search, actions, commit status, repo structure

**New Wave 2 Tools:**
- `check_gitea_version` - Check Gitea server version and API capabilities
- `get_workflow_file_content` - Get workflow files from .gitea/workflows/ or .github/workflows/
- `list_repo_structure` - List complete repository structure using Git tree API
- `monitor_workflow_dispatch` - Dispatch and monitor workflows until completion (Gitea 1.23+)
- `list_action_runners` - List self-hosted action runners (Gitea 1.23+)
- `create_commit_status` - Create commit status checks for CI/CD
- `list_action_artifacts` - List and download workflow artifacts (Gitea 1.23+)

**Gitea Version Compatibility:**
- Gitea 1.22.5: Limited Actions API support (no artifacts, runners, or workflow monitoring)
- Gitea 1.23+: Full Actions API support including all Wave 2 tools

## Setup for OpenClaw with Per-Agent Isolation

### Step 1: Build the MCP Server

```bash
cd mcp
make build
cp gitea-mcp ~/.bun/bin/gitea-mcp-mcp
```

### Step 2: Run Multiple MCP Instances

Each agent needs its own MCP server instance with its own token:

```bash
# Friday's MCP (runs on port 8081)
gitea-mcp-mcp --host https://gitea.example.com --port 8081 --token <YOUR_GITEA_TOKEN> &

# Karti's MCP (runs on port 8082)
gitea-mcp-mcp --host https://gitea.example.com --port 8082 --token <YOUR_GITEA_TOKEN> &
```

### Step 3: Configure MCP Servers in OpenClaw

Add to `~/.openclaw/openclaw.json`:

```json
{
  "plugins": {
    "entries": {
      "acpx": {
        "config": {
          "mcpServers": {
            "gitea-mcp-friday": {
              "command": "gitea-mcp-mcp",
              "args": ["--host", "https://gitea.example.com", "--port", "8081"],
              "env": {
                "GITEA_ACCESS_TOKEN": "<YOUR_GITEA_TOKEN>"
              }
            },
            "gitea-mcp-karti": {
              "command": "gitea-mcp-mcp",
              "args": ["--host", "https://gitea.example.com", "--port", "8082"],
              "env": {
                "GITEA_ACCESS_TOKEN": "<YOUR_GITEA_TOKEN>"
              }
            }
          }
        }
      }
    }
  }
}
```

### Step 4: Add Per-Agent Tool Restrictions

The key for isolation is `tools.allow/deny` per agent. This guarantees each agent can only use its own Gitea MCP:

```json
{
  "agents": {
    "list": [
      {
        "id": "friday",
        "name": "Friday Agent",
        "agentDir": "~/.openclaw/agents/friday/agent",
        "workspace": "/Users/karti/.openclaw/agents/friday/workspace",
        "tools": {
          "allow": ["gitea-mcp-friday:*", "group:fs", "group:runtime"],
          "deny": ["gitea-mcp-karti:*"]
        }
      },
      {
        "id": "karti",
        "name": "Karti Agent",
        "agentDir": "~/.openclaw/agents/karti/agent",
        "workspace": "/Users/karti/.openclaw/agents/karti/workspace",
        "tools": {
          "allow": ["gitea-mcp-karti:*", "group:fs", "group:runtime"],
          "deny": ["gitea-mcp-friday:*"]
        }
      }
    ]
  }
}
```

**Key Points:**
- `allow`: What tools the agent CAN use
- `deny`: What tools the agent CANNOT see/use
- `gitea-mcp-*:*` means all tools from that MCP server
- `group:fs`, `group:runtime` are built-in tool groups

### Step 5: Add Agent Instructions (AGENTS.md per agent)

Create per-agent AGENTS.md to enforce behavior:

**Friday's workspace** (`~/.openclaw/agents/friday/workspace/AGENTS.md`):
```markdown
# Friday Agent

You are the Friday Agent. You only have access to the gitea-mcp-friday MCP server.

When performing Git operations, use only the gitea-mcp-friday tools.
Never attempt to use gitea-mcp-karti or any other Gitea MCP server.
```

**Karti's workspace** (`~/.openclaw/agents/karti/workspace/AGENTS.md`):
```markdown
# Karti Agent

You are the Karti Agent. You only have access to the gitea-mcp-karti MCP server.

When performing Git operations, use only the gitea-mcp-karti tools.
Never attempt to use gitea-mcp-friday or any other Gitea MCP server.
```

## Available MCP Tools

Each MCP server provides these tools (prefixed with server name):

| Category | Tools |
|----------|-------|
| User | get_my_user_info, get_user_orgs, search_users |
| Repository | create_repo, fork_repo, list_my_repos, search_repos, list_repo_structure |
| Branches/Tags | create_branch, delete_branch, list_branches, create_tag, list_tags |
| Files | get_file_content, create_file, update_file, delete_file, get_dir_content |
| Issues | create_issue, list_repo_issues, create_issue_comment, edit_issue |
| Pull Requests | create_pull_request, list_repo_pull_requests, get_pull_request_by_index |
| Releases | create_release, list_releases, get_latest_release |
| Wiki | create_wiki_page, update_wiki_page, list_wiki_pages |
| Search | search_repos, search_users, search_org_teams |
| Server | get_gitea_mcp_server_version, check_gitea_version |
| Actions | get_workflow_file_content, monitor_workflow_dispatch, list_action_runners, list_action_artifacts, dispatch_repo_action_workflow, list_repo_action_runs |
| Commit Status | create_commit_status |

**Note:** Tools marked with (1.23+) require Gitea 1.23 or later:
- monitor_workflow_dispatch
- list_action_runners  
- list_action_artifacts

## Testing Per-Agent Isolation

Test with TUI for each agent:

```bash
# Test Friday (should only see gitea-mcp-friday tools)
openclaw tui --session friday --message "list available gitea tools"

# Test Karti (should only see gitea-mcp-karti tools)
openclaw tui --session karti --message "list available gitea tools"
```

## Adding New Agents

To add a new agent (e.g., "edith"):

1. Run a new MCP instance:
```bash
gitea-mcp-mcp --host https://gitea.example.com --port 8083 --token <YOUR_GITEA_TOKEN> &
```

2. Add to mcpServers in openclaw.json:
```json
"gitea-mcp-edith": {
  "command": "gitea-mcp-mcp",
  "args": ["--host", "https://gitea.example.com", "--port", "8083"],
  "env": { "GITEA_ACCESS_TOKEN": "<YOUR_GITEA_TOKEN>" }
}
```

3. Add agent with tool restrictions:
```json
{
  "id": "edith",
  "name": "Edith Agent",
  "tools": {
    "allow": ["gitea-mcp-edith:*", "group:fs", "group:runtime"],
    "deny": ["gitea-mcp-friday:*", "gitea-mcp-karti:*"]
  }
}
```

4. Create AGENTS.md in her workspace

## Troubleshooting

- **Agent can't see MCP tools**: Check `tools.allow` includes the MCP server name
- **Agent sees wrong tools**: Check `tools.deny` excludes other MCP servers
- **MCP not connecting**: Verify port is available and token is correct
- **Token not working**: Test directly: `curl -H "Authorization: token <YOUR_GITEA_TOKEN>" https://gitea.example.com/api/v1/user`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITEA_HOST` | Gitea server URL | https://gitea.com |
| `GITEA_ACCESS_TOKEN` | Access token | (required) |
| `GITEA_READONLY` | Enable read-only mode | false |
| `GITEA_DEBUG` | Enable debug logging | false |
| `GITEA_INSECURE` | Allow insecure TLS | false |