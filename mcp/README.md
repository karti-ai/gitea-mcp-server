# Gitea MCP Server

**Gitea MCP Server** is an integration plugin designed to connect Gitea with Model Context Protocol (MCP) systems. This allows for seamless command execution and repository management through an MCP-compatible chat interface.

[![Install with Docker in VS Code](https://img.shields.io/badge/VS_Code-Install_Server-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=gitea&inputs=[{%22id%22:%22gitea_token%22,%22type%22:%22promptString%22,%22description%22:%22Gitea%20Personal%20Access%20Token%22,%22password%22:true}]&config={%22command%22:%22docker%22,%22args%22:[%22run%22,%22-i%22,%22--rm%22,%22-e%22,%22GITEA_ACCESS_TOKEN%22,%22docker.gitea.com/gitea-mcp-server%22],%22env%22:{%22GITEA_ACCESS_TOKEN%22:%22${input:gitea_token}%22}}) [![Install with Docker in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Install_Server-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=gitea&inputs=[{%22id%22:%22gitea_token%22,%22type%22:%22promptString%22,%22description%22:%22Gitea%20Personal%20Access%20Token%22,%22password%22:true}]&config={%22command%22:%22docker%22,%22args%22:[%22run%22,%22-i%22,%22--rm%22,%22-e%22,%22GITEA_ACCESS_TOKEN%22,%22docker.gitea.com/gitea-mcp-server%22],%22env%22:{%22GITEA_ACCESS_TOKEN%22:%22${input:gitea_token}%22}}&quality=insiders)

## Table of Contents

- [Gitea MCP Server](#gitea-mcp-server)
  - [Table of Contents](#table-of-contents)
  - [What is Gitea?](#what-is-gitea)
  - [What is MCP?](#what-is-mcp)
  - [🚧 Installation](#-installation)
    - [Usage with Claude Code](#usage-with-claude-code)
    - [Usage with VS Code](#usage-with-vs-code)
    - [📥 Download the official binary release](#-download-the-official-binary-release)
    - [🔧 Build from Source](#-build-from-source)
    - [📁 Add to PATH](#-add-to-path)
  - [🚀 Usage](#-usage)
  - [✅ Available Tools](#-available-tools)
  - [🐛 Debugging](#-debugging)
  - [🛠 Troubleshooting](#-troubleshooting)

## What is Gitea?

Gitea is a community-managed lightweight code hosting solution written in Go. It is published under the MIT license. Gitea provides Git hosting including a repository viewer, issue tracking, pull requests, and more.

## What is MCP?

Model Context Protocol (MCP) is a protocol that allows for the integration of various tools and systems through a chat interface. It enables seamless command execution and management of repositories, users, and other resources.

## 🚧 Installation

### Usage with Claude Code

This method uses `go run` and requires [Go](https://go.dev) to be installed.

```bash
claude mcp add --transport stdio --scope user gitea \
  --env GITEA_ACCESS_TOKEN=token \
  --env GITEA_HOST=https://gitea.com \
  -- go run gitea.com/gitea/gitea-mcp@latest -t stdio
```

### Usage with VS Code

For quick installation, use one of the one-click install buttons at the top of this README.

For manual installation, add the following JSON block to your User Settings (JSON) file in VS Code. You can do this by pressing `Ctrl + Shift + P` and typing `Preferences: Open User Settings (JSON)`.

Optionally, you can add it to a file called `.vscode/mcp.json` in your workspace. This will allow you to share the configuration with others.

> Note that the `mcp` key is not needed in the `.vscode/mcp.json` file.

```json
{
  "mcp": {
    "inputs": [
      {
        "type": "promptString",
        "id": "gitea_token",
        "description": "Gitea Personal Access Token",
        "password": true
      }
    ],
    "servers": {
      "gitea-mcp": {
        "command": "docker",
        "args": [
          "run",
          "-i",
          "--rm",
          "-e",
          "GITEA_ACCESS_TOKEN",
          "docker.gitea.com/gitea-mcp-server"
        ],
        "env": {
          "GITEA_ACCESS_TOKEN": "${input:gitea_token}"
        }
      }
    }
  }
}
```

### 📥 Download the official binary release

You can download the official release from [official Gitea MCP binary releases](https://gitea.com/gitea/gitea-mcp/releases).

### 🔧 Build from Source

You can download the source code by cloning the repository using Git:

```bash
git clone https://gitea.com/gitea/gitea-mcp.git
```

Before building, make sure you have the following installed:

- make
- Golang (Go 1.24 or later recommended)

Then run:

```bash
make install
```

### 📁 Add to PATH

After installing, copy the binary gitea-mcp to a directory included in your system's PATH. For example:

```bash
cp gitea-mcp /usr/local/bin/
```

## 🚀 Usage

This example is for Cursor, you can also use plugins in VSCode.
To configure the MCP server for Gitea, add the following to your MCP configuration file:

- **stdio mode**

```json
{
  "mcpServers": {
    "gitea": {
      "command": "gitea-mcp",
      "args": [
        "-t",
        "stdio",
        "--host",
        "https://gitea.com"
        // "--token", "<your personal access token>"
      ],
      "env": {
        // "GITEA_HOST": "https://gitea.com",
        // "GITEA_INSECURE": "true",
        "GITEA_ACCESS_TOKEN": "<your personal access token>"
      }
    }
  }
}
```

- **http mode**

```json
{
  "mcpServers": {
    "gitea": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer <your personal access token>"
      }
    }
  }
}
```

**Default log path**: `$HOME/.gitea-mcp/gitea-mcp.log`

> [!NOTE]
> You can provide your Gitea host and access token either as command-line arguments or environment variables.
> Command-line arguments have the highest priority

Once everything is set up, try typing the following in your MCP-compatible chatbox:

```text
list all my repositories
```

## ✅ Available Tools

The Gitea MCP Server supports the following tools:

|               Tool                |    Scope     |                       Description                        |
| :-------------------------------: | :----------: | :------------------------------------------------------: |
|         get_my_user_info          |     User     |      Get the information of the authenticated user       |
|           get_user_orgs           |     User     | Get organizations associated with the authenticated user |
|            create_repo            |  Repository  |                 Create a new repository                  |
|             fork_repo             |  Repository  |                    Fork a repository                     |
|           list_my_repos           |  Repository  |  List all repositories owned by the authenticated user   |
|           create_branch           |    Branch    |                   Create a new branch                    |
|           delete_branch           |    Branch    |                     Delete a branch                      |
|           list_branches           |    Branch    |            List all branches in a repository             |
|          create_release           |   Release    |           Create a new release in a repository           |
|          delete_release           |   Release    |            Delete a release from a repository            |
|            get_release            |   Release    |                      Get a release                       |
|        get_latest_release         |   Release    |          Get the latest release in a repository          |
|           list_releases           |   Release    |            List all releases in a repository             |
|            create_tag             |     Tag      |                     Create a new tag                     |
|            delete_tag             |     Tag      |                       Delete a tag                       |
|              get_tag              |     Tag      |                        Get a tag                         |
|             list_tags             |     Tag      |              List all tags in a repository               |
|         list_repo_commits         |    Commit    |             List all commits in a repository             |
|         get_file_content          |     File     |          Get the content and metadata of a file          |
|          get_dir_content          |     File     |           Get a list of entries in a directory           |
|            create_file            |     File     |                    Create a new file                     |
|            update_file            |     File     |                 Update an existing file                  |
|            delete_file            |     File     |                      Delete a file                       |
|        get_issue_by_index         |    Issue     |                Get an issue by its index                 |
|         list_repo_issues          |    Issue     |             List all issues in a repository              |
|           create_issue            |    Issue     |                    Create a new issue                    |
|       create_issue_comment        |    Issue     |               Create a comment on an issue               |
|            edit_issue             |    Issue     |                       Edit a issue                       |
|        edit_issue_comment         |    Issue     |                Edit a comment on an issue                |
|    get_issue_comments_by_index    |    Issue     |          Get comments of an issue by its index           |
|     get_pull_request_by_index     | Pull Request |             Get a pull request by its index              |
|       get_pull_request_diff       | Pull Request |                 Get a pull request diff                  |
|      list_repo_pull_requests      | Pull Request |          List all pull requests in a repository          |
|        create_pull_request        | Pull Request |                Create a new pull request                 |
|   create_pull_request_reviewer    | Pull Request |             Add reviewers to a pull request              |
|   delete_pull_request_reviewer    | Pull Request |           Remove reviewers from a pull request           |
|     list_pull_request_reviews     | Pull Request |           List all reviews for a pull request            |
|      get_pull_request_review      | Pull Request |               Get a specific review by ID                |
| list_pull_request_review_comments | Pull Request |            List inline comments for a review             |
|    create_pull_request_review     | Pull Request |      Create a review with optional inline comments       |
|    submit_pull_request_review     | Pull Request |                 Submit a pending review                  |
|    delete_pull_request_review     | Pull Request |                     Delete a review                      |
|    dismiss_pull_request_review    | Pull Request |          Dismiss a review with optional message          |
|        merge_pull_request         | Pull Request |                  Merge a pull request                    |
|        check_gitea_version        |    Server    | Check Gitea server version and API capabilities |
|      get_workflow_file_content      |   Actions    | Get workflow file content from .gitea/workflows/ or .github/workflows/ |
|        list_repo_structure        |  Repository  | List complete directory and file structure using Git tree API |
|      monitor_workflow_dispatch      |   Actions    | Dispatch and monitor workflow until completion (requires Gitea 1.23+) |
|        create_commit_status         |  Repository  | Create a commit status check for CI/CD integration |
|           search_users            |     User     |                     Search for users                     |
|         search_org_teams          | Organization |           Search for teams in an organization            |
|          list_org_labels          | Organization |        List labels defined at organization level         |
|         create_org_label          | Organization |            Create a label in an organization             |
|          edit_org_label           | Organization |             Edit a label in an organization              |
|         delete_org_label          | Organization |            Delete a label in an organization             |
|           search_repos            |  Repository  |                 Search for repositories                  |
|     list_repo_action_secrets      |   Actions    |     List repository Actions secrets (metadata only)      |
|     upsert_repo_action_secret     |   Actions    |    Create/update (upsert) a repository Actions secret    |
|     delete_repo_action_secret     |   Actions    |            Delete a repository Actions secret            |
|      list_org_action_secrets      |   Actions    |    List organization Actions secrets (metadata only)     |
|     upsert_org_action_secret      |   Actions    |  Create/update (upsert) an organization Actions secret   |
|     delete_org_action_secret      |   Actions    |          Delete an organization Actions secret           |
|    list_repo_action_variables     |   Actions    |            List repository Actions variables             |
|     get_repo_action_variable      |   Actions    |            Get a repository Actions variable             |
|    create_repo_action_variable    |   Actions    |           Create a repository Actions variable           |
|    update_repo_action_variable    |   Actions    |           Update a repository Actions variable           |
|    delete_repo_action_variable    |   Actions    |           Delete a repository Actions variable           |
|     list_org_action_variables     |   Actions    |           List organization Actions variables            |
|      get_org_action_variable      |   Actions    |           Get an organization Actions variable           |
|    create_org_action_variable     |   Actions    |         Create an organization Actions variable          |
|    update_org_action_variable     |   Actions    |         Update an organization Actions variable          |
|    delete_org_action_variable     |   Actions    |         Delete an organization Actions variable          |
|    list_repo_action_workflows     |   Actions    |            List repository Actions workflows             |
|     get_repo_action_workflow      |   Actions    |            Get a repository Actions workflow             |
|   dispatch_repo_action_workflow   |   Actions    |     Trigger (dispatch) a repository Actions workflow     |
|       list_repo_action_runs       |   Actions    |               List repository Actions runs               |
|        get_repo_action_run        |   Actions    |               Get a repository Actions run               |
|      cancel_repo_action_run       |   Actions    |             Cancel a repository Actions run              |
|       rerun_repo_action_run       |   Actions    |              Rerun a repository Actions run              |
|       list_repo_action_jobs       |   Actions    |               List repository Actions jobs               |
|     list_repo_action_run_jobs     |   Actions    |               List Actions jobs for a run                |
|          list_action_runners          |   Actions    | List self-hosted action runners (requires Gitea 1.23+) |
|     list_action_artifacts      |   Actions    | List and download artifacts from workflow runs (requires Gitea 1.23+) |
|    get_repo_action_job_log_preview    |   Actions    |           Get a job log preview (tail/limited)           |
|   download_repo_action_job_log    |   Actions    |               Download a job log to a file               |
|   get_gitea_mcp_server_version    |    Server    |         Get the version of the Gitea MCP Server          |
|          list_wiki_pages          |     Wiki     |           List all wiki pages in a repository            |
|           get_wiki_page           |     Wiki     |           Get a wiki page content and metadata           |
|        get_wiki_revisions         |     Wiki     |           Get revisions history of a wiki page           |
|         create_wiki_page          |     Wiki     |                  Create a new wiki page                  |
|         update_wiki_page          |     Wiki     |               Update an existing wiki page               |
|         delete_wiki_page          |     Wiki     |                    Delete a wiki page                    |

## 🆕 Wave 2 Tools

The following tools were added in Wave 2 and provide enhanced functionality for repository management, workflow operations, and CI/CD integration.

### Server Tools

#### check_gitea_version

Check the Gitea server version and API capabilities. Returns version string, parsed components, and a capability matrix indicating which APIs are available.

**Example:**
```json
{
  "version": "1.23.1",
  "major": 1,
  "minor": 23,
  "patch": 1,
  "capabilities": {
    "actions_api": true,
    "secrets_api": true,
    "variables_api": true,
    "runners_api": true,
    "artifacts_api": true,
    "commit_status_api": true
  }
}
```

### Repository Tools

#### list_repo_structure

List the complete directory and file structure of a repository using the Git tree API. Supports recursive listing, pattern filtering, and pagination.

**Parameters:**
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `ref`: Git reference (branch, tag, or commit SHA). Defaults to default branch
- `pattern`: Glob pattern to filter files (e.g., '*.yml', '.gitea/*', 'src/**/*.go')
- `recursive`: List contents recursively (default: true)
- `page`: Page number for pagination (default: 1)
- `per_page`: Items per page (default: 100, max: 1000)

**Example:**
```bash
# List all workflow files
list_repo_structure owner="gitea" repo="gitea-mcp" pattern=".gitea/workflows/*"

# List Go source files recursively
list_repo_structure owner="gitea" repo="gitea-mcp" pattern="**/*.go" recursive=true
```

#### create_commit_status

Create a commit status check for CI/CD integration. Adds a new status context to a commit without overwriting existing statuses.

**Parameters:**
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `sha` (required): Commit SHA (full 40-character or short SHA)
- `state` (required): Status state (pending, success, error, failure)
- `target_url`: URL with more details (e.g., review environment link)
- `context`: Status context identifier (default: "default")
- `description`: Short description of the status

**Example:**
```bash
# Set CI status to success
create_commit_status owner="gitea" repo="my-project" sha="abc123..." state="success" context="ci/build" description="Build passed"

# Mark deployment as pending
create_commit_status owner="gitea" repo="my-project" sha="abc123..." state="pending" context="deploy/review" target_url="https://review.example.com"
```

### Actions Tools

#### get_workflow_file_content

Get workflow file content from `.gitea/workflows/` or `.github/workflows/` directories. Auto-discovers workflow files and returns parsed YAML as JSON.

**Parameters:**
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `ref`: Git reference (branch/tag/commit). Defaults to default branch
- `pattern`: File pattern to match (e.g., '*.yml', 'build-*.yml')
- `filename`: Specific workflow filename to retrieve (ignores pattern if provided)

**Example:**
```bash
# Get all workflow files
get_workflow_file_content owner="gitea" repo="gitea-mcp"

# Get specific workflow file
get_workflow_file_content owner="gitea" repo="gitea-mcp" filename="build.yml"

# Get workflow files matching pattern
get_workflow_file_content owner="gitea" repo="gitea-mcp" pattern="test-*.yml"
```

#### monitor_workflow_dispatch

Dispatch a workflow and monitor its execution until completion. Returns full execution summary including run ID, status, conclusion, duration, and logs.

**Note:** Requires Gitea 1.23+. Not available in Gitea 1.22.5.

**Parameters:**
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `workflow_id` (required): Workflow ID or filename
- `ref` (required): Git reference (branch/tag) to run workflow on
- `inputs`: Workflow inputs object
- `timeout_seconds`: Polling timeout in seconds (default: 300 = 5 minutes)
- `poll_interval_seconds`: Poll interval in seconds (default: 10)

**Example:**
```bash
# Dispatch and monitor a workflow
monitor_workflow_dispatch owner="gitea" repo="my-project" workflow_id="build.yml" ref="main"

# Dispatch with inputs
monitor_workflow_dispatch owner="gitea" repo="my-project" workflow_id="deploy.yml" ref="main" inputs='{"environment": "staging"}'

# Custom timeout and poll interval
monitor_workflow_dispatch owner="gitea" repo="my-project" workflow_id="build.yml" ref="main" timeout_seconds=600 poll_interval_seconds=30
```

#### list_action_runners

List self-hosted action runners for a repository. Shows runner status, labels, and availability.

**Note:** Requires Gitea 1.23+. Returns empty list with message on Gitea 1.22.5.

**Parameters:**
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `status`: Filter by status (online, offline, busy, idle)
- `page`: Page number (default: 1)
- `perPage`: Results per page (default: 30)

**Example:**
```bash
# List all runners
list_action_runners owner="gitea" repo="my-project"

# List only online runners
list_action_runners owner="gitea" repo="my-project" status="online"

# List busy runners
list_action_runners owner="gitea" repo="my-project" status="busy"
```

#### list_action_artifacts

List and download artifacts from workflow runs. Supports listing artifacts, getting specific artifact details, and downloading artifact content.

**Note:** Requires Gitea 1.23+. Returns empty list with message on Gitea 1.22.5.

**Methods:** `list`, `get`, `download`

**Parameters:**
- `method` (required): Operation to perform (list, get, download)
- `owner` (required): Repository owner
- `repo` (required): Repository name
- `run_id`: Run ID to filter artifacts (optional for list, required for get/download)
- `artifact_name`: Artifact name to filter (optional)
- `artifact_id`: Artifact ID (required for get and download methods)
- `output_path`: Output file path for download method
- `max_size`: Maximum artifact size in bytes (default: 100MB)
- `page`: Page number (default: 1)
- `perPage`: Results per page (default: 30)

**Example:**
```bash
# List all artifacts for a run
list_action_artifacts method="list" owner="gitea" repo="my-project" run_id=123

# Get specific artifact details
list_action_artifacts method="get" owner="gitea" repo="my-project" artifact_id=456

# Download artifact
list_action_artifacts method="download" owner="gitea" repo="my-project" artifact_id=456 output_path="./build-artifact.zip"
```

## 🔧 Gitea Version Compatibility

Gitea MCP Server supports different feature sets depending on your Gitea server version. Use the `check_gitea_version` tool to detect available APIs.

### Gitea 1.22.5

The following features are **NOT available** in Gitea 1.22.5:

| Feature | Status | Notes |
|---------|--------|-------|
| Actions API | Limited | Basic workflow support only |
| Artifacts API | Not available | Returns empty list with message |
| Runners API | Not available | Returns empty list with message |
| Workflow monitoring | Not available | Use basic dispatch only |

**Compatible tools:**
- `dispatch_repo_action_workflow` (basic dispatch without monitoring)
- `list_repo_action_runs` (limited support)
- `get_repo_action_run`
- `cancel_repo_action_run`
- `rerun_repo_action_run`

### Gitea 1.23+

Full Actions API support including:

| Feature | Status | Notes |
|---------|--------|-------|
| Actions API | Full | Complete workflow support |
| Artifacts API | Full | List, get, and download artifacts |
| Runners API | Full | List and manage self-hosted runners |
| Workflow monitoring | Full | Dispatch with monitoring until completion |
| Secrets API | Full | Repository and organization secrets |
| Variables API | Full | Repository and organization variables |

**All Wave 2 tools are fully supported on Gitea 1.23+.**

### Version Detection

Always check your Gitea version before using Actions-related tools:

```bash
check_gitea_version
```

The response includes a `capabilities` object that indicates which APIs are available.

## 🛡️ Enhanced Error Handling

Wave 1 introduced comprehensive error handling improvements that provide clearer, more actionable error messages.

### Error Categories

Errors are now categorized for better handling:

| Category | Description | Example |
|----------|-------------|---------|
| `auth` | Authentication/authorization errors | Invalid token, insufficient permissions |
| `not_found` | Resource not found | Repository, file, or issue doesn't exist |
| `network` | Network connectivity issues | Connection timeout, DNS failures |
| `timeout` | Operation timeouts | Request took too long |
| `server` | Server-side errors | 5xx errors from Gitea |
| `validation` | Input validation errors | Invalid parameters, missing required fields |
| `file` | File operation errors | File not found, path issues |
| `actions` | Actions API errors | Workflow not found, API unavailable |

### Error Context

All errors now include context information to help with debugging:

```json
{
  "error": "File or directory not found",
  "category": "file",
  "operation": "GetFile",
  "timestamp": "2024-01-15T10:30:00Z",
  "context": {
    "owner": "gitea",
    "path": "README.md"
  }
}
```

### Helper Functions

The enhanced error system provides helper functions for checking error types:

- `IsNotFound(err)` - Check if error is a 404/not found
- `IsAuthError(err)` - Check if error is auth-related (401/403)
- `IsTimeout(err)` - Check if error is a timeout
- `IsNetworkError(err)` - Check if error is network-related
- `IsUnauthorized(err)` - Check for 401 specifically
- `IsForbidden(err)` - Check for 403 specifically
- `IsServerError(err)` - Check for 5xx errors
- `IsActionsAPIUnavailable(err)` - Check if Actions API is unavailable

### Error Formatting

Errors can be formatted for different purposes:

```go
// Human-readable format
enhancedErr.Format()
// Output: "Operation: GetFile | Error: File not found | Category: file | Context: owner=gitea, path=README.md"

// Detailed JSON-like format
enhancedErr.FormatDetailed()
// Output: Full JSON with all error details
```

## 🐛 Debugging

To enable debug mode, add the `-d` flag when running the Gitea MCP Server with http mode:

```sh
./gitea-mcp -t http [--port 8080] --token <your personal access token> -d
```

## 🛠 Troubleshooting

If you encounter any issues, here are some common troubleshooting steps:

1. **Check your PATH**: Ensure that the `gitea-mcp` binary is in a directory included in your system's PATH.
2. **Verify dependencies**: Make sure you have all the required dependencies installed, such as `make` and `Golang`.
3. **Review configuration**: Double-check your MCP configuration file for any errors or missing information.
4. **Consult logs**: Check the logs for any error messages or warnings that can provide more information about the issue.

Enjoy exploring and managing your Gitea repositories via chat!
