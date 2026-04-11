# Gitea MCP Server

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](mcp/.gitea/workflows)

A comprehensive MCP (Model Context Protocol) server for Gitea repository management and automation.

## Features

- **50+ Tools**: Complete Gitea API coverage - repos, issues, PRs, branches, files, wiki, search, actions, and more
- **Multi-Platform**: Linux, Windows, macOS (AMD64 & ARM64)
- **Multiple Modes**: stdio (default) for local MCP clients, HTTP mode for remote servers
- **Per-Agent Isolation**: Each MCP instance uses a single Gitea token - perfect for multi-agent setups
- **Wave 2 Features**: Workflow monitoring, action runners, commit status checks, repository structure listing

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/karti-ai/gitea-mcp-server.git
cd gitea-mcp-server/mcp

# Build
make build

# Or install to $GOPATH/bin
make install
```

### Usage

```bash
# stdio mode (default)
gitea-mcp --host https://gitea.example.com --token YOUR_TOKEN

# HTTP mode
gitea-mcp --host https://gitea.example.com --token YOUR_TOKEN --port 8080
```

## Documentation

- **[Full Documentation](mcp/README.md)** - Complete usage guide, all 50+ tools, examples
- **[Building Guide](mcp/BUILDING.md)** - Windows build instructions
- **[Agent Setup](AGENTS.md)** - OpenClaw integration with per-agent isolation
- **[Chinese Docs](mcp/README.zh-cn.md)** - 简体中文文档
- **[Chinese (Traditional)](mcp/README.zh-tw.md)** - 繁體中文文件

## Project Structure

```
gitea-mcp-server/
├── mcp/                    # MCP server source code
│   ├── README.md          # Full documentation
│   ├── Makefile           # Build automation
│   ├── Dockerfile         # Container build
│   └── ...
├── AGENTS.md              # OpenClaw setup guide
└── LICENSE                # MIT License
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/karti-ai">Karti AI</a>
</p>
