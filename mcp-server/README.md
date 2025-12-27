# AIGateway MCP Server

Model Context Protocol (MCP) server for AIGateway that exposes model management capabilities to Claude and other MCP clients.

## Features

- **List Models** - View available built-in models and your custom model mappings
- **Create Custom Mapping** - Create user-owned model aliases (e.g., `my-gpt` → `gpt-4`)
- **Update Custom Mapping** - Modify existing custom model mappings

## Prerequisites

- Python 3.8+
- AIGateway server running on http://localhost:8088 (or configured URL)
- Valid AIGateway API key (format: `ak_...`)

## Installation

1. Clone or copy this repository
2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

3. Create `.env` file (copy from `.env.example`):
   ```bash
   cp .env.example .env
   ```

4. Edit `.env` and configure:
   ```env
   AIGATEWAY_URL=http://localhost:8088
   AIGATEWAY_API_KEY=ak_your_api_key_here
   ```

## Running the Server

### As Stdio Server (for Claude Desktop/Code)

```bash
python -m src.mcp_server
```

The server will start and wait for MCP client connections via stdin/stdout.

### Direct Execution

```bash
python src/mcp_server.py
```

You'll see logs like:
```
2025-12-27 10:30:45 - __main__ - INFO - Starting AIGateway MCP Server
2025-12-27 10:30:45 - __main__ - INFO - AIGateway URL: http://localhost:8088
2025-12-27 10:30:45 - __main__ - INFO - API Key configured: True
```

## Configuration with Claude Desktop

### macOS
Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "aigateway": {
      "command": "python",
      "args": ["-m", "src.mcp_server"],
      "env": {
        "AIGATEWAY_URL": "http://localhost:8088",
        "AIGATEWAY_API_KEY": "ak_your_api_key_here"
      }
    }
  }
}
```

### Windows
Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "aigateway": {
      "command": "python",
      "args": ["-m", "src.mcp_server"],
      "env": {
        "AIGATEWAY_URL": "http://localhost:8088",
        "AIGATEWAY_API_KEY": "ak_your_api_key_here"
      }
    }
  }
}
```

### Linux
Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "aigateway": {
      "command": "python",
      "args": ["-m", "src.mcp_server"],
      "env": {
        "AIGATEWAY_URL": "http://localhost:8088",
        "AIGATEWAY_API_KEY": "ak_your_api_key_here"
      }
    }
  }
}
```

## Configuration with Claude Code

Edit `~/.config/claude-code/mcp.json`:

```json
{
  "mcpServers": {
    "aigateway": {
      "command": "python",
      "args": ["-m", "src.mcp_server"],
      "env": {
        "AIGATEWAY_URL": "http://localhost:8088",
        "AIGATEWAY_API_KEY": "ak_your_api_key_here"
      }
    }
  }
}
```

## Available Tools

### 1. list_models

Lists all available AI models including your custom mappings.

**Parameters:**
- `api_key` (string, required) - Your AIGateway API key (ak_...)

**Response:**
```
{
  "built_in_models": [
    {"alias": "claude-opus-4-1", "provider_id": "antigravity", ...},
    {"alias": "gpt-4", "provider_id": "openai", ...},
    ...
  ],
  "custom_mappings": [
    {"alias": "my-gpt", "provider_id": "openai", "model_name": "gpt-4", ...}
  ],
  "total": 15
}
```

### 2. create_mapping

Create a new custom model mapping (user-owned).

**Parameters:**
- `api_key` (string, required) - Your AIGateway API key
- `alias` (string, required) - Unique alias for the mapping (e.g., "my-gpt")
- `provider_id` (string, required) - Provider: `antigravity`, `openai`, or `glm`
- `model_name` (string, required) - Actual model name at the provider
- `description` (string, optional) - Description of this mapping

**Valid Providers:**
- `antigravity` - Anthropic Claude models, Google Gemini
- `openai` - OpenAI GPT models
- `glm` - Zhipu GLM models

**Example:**
```
alias: my-gpt
provider_id: openai
model_name: gpt-4-turbo
description: My favorite GPT-4 Turbo instance
```

### 3. update_mapping

Update an existing custom model mapping.

**Parameters:**
- `api_key` (string, required) - Your AIGateway API key
- `alias` (string, required) - Current alias to update
- `new_alias` (string, optional) - New alias name
- `provider_id` (string, optional) - New provider
- `model_name` (string, optional) - New model name
- `description` (string, optional) - New description
- `enabled` (boolean, optional) - Enable/disable mapping

**Note:** Only update fields you want to change. Other fields remain unchanged.

## Security

### API Key Handling
- API keys are passed as tool parameters
- Never stored or logged (only first 12 characters logged for debugging)
- Transmitted to AIGateway over HTTPS in production

### Permissions
- Each user can only access their own custom mappings
- Cannot modify or delete global (admin-created) mappings
- Built-in models are visible to all users

## Troubleshooting

### Connection Failed
```
Error: Failed to list models: [Errno -2] Name or service not known
```
**Solution:** Check that AIGateway is running on the configured URL:
```bash
curl http://localhost:8088/health
```

### Invalid API Key
```
Error: Failed to list models: 401 Client Error
```
**Solution:** Verify your API key is correct and active in AIGateway

### Module Not Found
```
ModuleNotFoundError: No module named 'fastmcp'
```
**Solution:** Install dependencies:
```bash
pip install -r requirements.txt
```

## Development

### Project Structure
```
aigateway-mcp-python/
├── src/
│   └── mcp_server.py          # MCP server implementation
├── requirements.txt            # Python dependencies
├── .env.example               # Configuration template
└── README.md                  # This file
```

### Adding More Tools

To add new tools, follow the FastMCP pattern in `mcp_server.py`:

```python
@mcp.tool()
async def new_tool(api_key: str, param: str) -> str:
    """
    Tool description shown to Claude.

    Args:
        api_key: User's API key from AIGateway (ak_...)
        param: Parameter description

    Returns:
        JSON string with results
    """
    try:
        # Implementation
        return str(result)
    except Exception as e:
        error_msg = f"Failed: {str(e)}"
        logger.error(error_msg)
        return f"Error: {error_msg}"
```

## Testing

### Manual Test with curl

1. Start the MCP server:
   ```bash
   python src/mcp_server.py
   ```

2. Verify it's running by sending a test JSON-RPC request via stdin (advanced)

### Integration Testing

Test with Claude Desktop:
1. Configure MCP server in `claude_desktop_config.json`
2. Restart Claude Desktop
3. In conversation, ask Claude to use AIGateway tools:
   - "List all available models"
   - "Create a custom mapping called 'my-model' for gpt-4"
   - "Show me my custom mappings"

## License

Part of AIGateway project.
