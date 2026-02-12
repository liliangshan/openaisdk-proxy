# OpenAI Proxy & Token Optimization Tool

## Background & Pain Points

In large language model (LLM) application development, context window limitations and token consumption costs have always been two core challenges. When developers need to handle complex projects, they often need to let AI read numerous code files to understand the project structure. A medium-sized project may contain dozens of files with tens of thousands of lines of code; sending all this content to AI for every conversation means:

**High cost consumption**: Assuming each request needs to process 50,000 tokens, at GPT-4 pricing (approximately $30-60 per million tokens), a single request costs $1.5-3. If multiple developers are using the project simultaneously or dozens of conversation rounds are needed, monthly bills can easily reach hundreds or even thousands of dollars.

**Severe context waste**: A large amount of code in projects is legacy, explanatory comments, or duplicate implementations. These contents have little relevance to the current task but occupy valuable context space. Worse, as conversation history grows, context gets filled with old code and duplicate information, causing AI to fail to effectively understand the latest requirements.

**Forced context truncation**: When context exceeds limits, AI can only forcibly truncate historical information, potentially losing critical code dependencies or design decision records, affecting code quality and development efficiency.

## Solutions

This tool is designed to solve the pain points mentioned above. It acts as an intelligent proxy layer between your application and the LLM API, significantly reducing costs and improving efficiency through:

**Token Length Condensation**: When the total message length exceeds the threshold, automatically condense the overly long text content in early messages, while retaining the complete content of the most recent N rounds of dialogue. This is an efficient compression strategy that significantly reduces token consumption while preserving dialogue structure.

**Flexible Compression Policies**: You can customize compression triggers, the number of dialogue rounds to retain, message role types to condense, and other parameters based on project characteristics and requirements, achieving fine-grained control.

**Multi-Model & Multi-Provider Support**: Unified management of different AI providers and model configurations, simplifying API calls through aliases without hardcoding complex request parameters in business code.

A lightweight OpenAI-compatible API proxy service with token compression to reduce costs.

## Features

### 1. OpenAI API Proxy
- Compatible with OpenAI Chat Completion API
- Support for multiple model configurations
- Flexible multi-provider, multi-model configuration
- API key management

### 2. Token Compression
- Smart conversation compression
- Automatic condensation of overly long text
- Configurable compression policies
- Context integrity preservation

### 3. Admin Interface
- Model management
- Provider configuration
- User interface
- Real-time monitoring

## Quick Start

### Requirements
- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### Configuration

```yaml
# Configuration file: cmd/api/config.yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  name: "model_system"

server:
  host: "0.0.0.0"
  port: 8080

jwt:
  secret: "your-secret-key"
  expiration: "8760h"
```

### Admin Interface

After starting the service, you can access the admin interface via browser for configuration management.

**Access URL**: `http://127.0.0.1:8080/user`

**Usage Process**:

1. **Register Account**
   - First access to admin interface, click "Register" button
   - Fill in username, email, password to complete registration
   - Automatically logged in after successful registration

2. **Login System**
   - Login using registered account and password
   - Supports JWT Token auto-renewal

3. **Get API Key**
   - After login, go to personal center or settings page
   - Create API Key for interface calls
   - Each user can create multiple API Keys

4. **Configure Models**
   - Add AI provider configuration (name, API URL, key, etc.)
   - Add model configuration (select provider, model ID, set compression parameters, etc.)
   - Enable Token compression feature to reduce calling costs

### Build & Run

```bash
# Build frontend and backend
./build.sh

# Start service
./bin/openaisdk-proxy
```

### API Usage

```bash
# Send request
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "prefix-model-alias",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

## Core Configuration

### Provider Configuration
| Parameter | Description |
|-----------|-------------|
| name | Provider identifier |
| display_name | Display name |
| base_url | API endpoint URL |
| api_prefix | API request prefix |
| api_key | Provider API key |

### Model Configuration
| Parameter | Description |
|-----------|-------------|
| model_id | Source model ID |
| display_name | Model alias (used in requests) |
| context_length | Context length (unit: k) |
| compress_enabled | Whether to enable compression |
| compress_truncate_len | Message length threshold to trigger compression |
| compress_user_count | Number of recent dialogue rounds to retain |
| compress_role_types | Message role types to retain (comma-separated) |

## Compression Strategy

### How It Works
1. Count tokens in request messages
2. Automatically condense overly long text in early messages when exceeding threshold
3. Retain the complete content of the most recent N rounds of dialogue, while condensing the text length of earlier messages
4. Configurable compression parameters

### Parameter Description
- `compress_enabled`: Whether to enable token compression
- `compress_truncate_len`: Trigger compression when message length exceeds this value (unit: Token)
- `compress_user_count`: Retain the complete content of the most recent N rounds of dialogue, earlier messages will be condensed
- `compress_role_types`: Message role types whose text length should be condensed, defaulting to user and assistant

### Compression Effect Example

Suppose there is a conversation history with 10 rounds of dialogue and a total of 100 tokens, with the following configuration:
- `compress_enabled`: true
- `compress_truncate_len`: 10
- `compress_user_count`: 3

The system will:
1. Detect that 100 exceeds the threshold of 10 (unit: Token)
2. Retain the complete content of the most recent 3 rounds of user dialogue and their corresponding assistant responses
3. Condense the long text in earlier conversations (truncate overly long content), preserve message structure without deletion
4. Compress the token count to approximately 10 or less
5. **Cost savings**: Original cost ÷ New token ratio = Cost reduction of 90%

### Real-world Performance Data

The following is actual log data from production environment (2026-02-13):

```
client IP: 3.209.66.12, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 15132 (original tokens: 40730) 
[CONTEXT] Truncated long text before 12th user message (total messages: 58)

client IP: 52.44.113.131, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 15217 (original tokens: 40815)
[CONTEXT] Truncated long text before 12th user message (total messages: 60)

client IP: 52.44.113.131, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 16347 (original tokens: 41945)
[CONTEXT] Truncated long text before 12th user message (total messages: 78)
```

**Actual Compression Efficiency**:
- Average cost savings: **62-63%**
- Original token range: 40,730 - 41,945
- Compressed token range: 15,132 - 16,347
- Actual cost reduction: approximately **3.9x**

## Project Structure

```
.
├── cmd/api/                    # Backend application entry
│   ├── main.go                 # Main program
│   ├── config.yaml             # Configuration file
│   └── internal/
│       ├── handlers/           # HTTP request handlers
│       │   ├── chat.go         # Chat API handler
│       │   ├── models.go       # Model management API
│       │   └── ...
│       ├── service/            # Business logic layer
│       ├── repository/         # Data access layer
│       ├── models/             # Data models
│       └── cache/              # Cache management
├── frontend/                   # Frontend application
│   ├── src/
│   │   ├── views/              # Page components
│   │   ├── components/         # Common components
│   │   └── ...
│   └── package.json
├── bin/                        # Build output directory
│   └── openaisdk-proxy         # Executable file
├── build.sh                    # One-click build script
└── README.md                   # English documentation
```

## API Documentation

### 1. Chat Completion API

```bash
POST /v1/chat/completions

# Request body example
{
  "model": "prefix-model-alias",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is 2+2?"}
  ],
  "temperature": 0.7,
  "max_tokens": 100,
  "top_p": 0.9,
  "stream": false
}

# Response example
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "prefix-model-alias",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "2+2 equals 4."
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

### 2. Streaming Response

Set `"streamtotal_tokens": ": true` to get streaming response:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "prefix-model-alias", "messages": [...], "stream": true}'
```

### 3. Common Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| model | string | Model alias in format `prefix-alias` |
| messages | array | Message list, must contain role and content |
| temperature | float | Generation diversity, range 0-2, default 0.7 |
| max_tokens | int | Maximum generated tokens |
| top_p | float | Nucleus sampling parameter, range 0-1 |
| stream | bool | Whether to return streaming response |

## FAQ

### Q: How do I add a new model?
A: Configure the model information in the admin interface or database, including model ID, alias, provider, etc. The system will automatically cache it.

### Q: Will compression lose important information?
A: The compression strategy retains recent conversation history and only deletes earlier content. You can adjust the `compress_user_count` parameter to control how many dialogue rounds are retained.

### Q: How do I monitor token usage?
A: View real-time logs and statistics in the admin interface, or check the `usage` field in API responses to understand consumption for each request.

### Q: Which LLM providers are supported?
A: Theoretically all OpenAI-compatible APIs are supported, including but not limited to OpenAI, Azure, Anthropic, etc.

### Q: How do I deploy to production?
A: Refer to the deployment guide below. Use Docker, Kubernetes, or system service managers (such as systemd) to run the service.

## Deployment Guide

### Docker Deployment

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN ./build.sh

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/bin/openaisdk-proxy .
COPY --from=builder /app/cmd/api/config.yaml .
EXPOSE 8080
CMD ["./openaisdk-proxy"]
```

### Systemd Service Configuration

Create file `/etc/systemd/system/openaisdk-proxy.service`:

```ini
[Unit]
Description=OpenAI SDK Proxy Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/openaisdk-proxy
ExecStart=/opt/openaisdk-proxy/bin/openaisdk-proxy
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

Then run:
```bash
sudo systemctl daemon-reload
sudo systemctl enable openaisdk-proxy
sudo systemctl start openaisdk-proxy
```

### Environment Variable Configuration

Support the following environment variables to override configuration file:

```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=model_system
API_PORT=8080
JWT_SECRET=your-secret-key
```

## Performance Optimization Recommendations

1. **Database Optimization**
   - Create indexes on `api_keys` and `models` tables
   - Regularly clean up old log data
   - Use connection pools to manage database connections

2. **Cache Optimization**
   - Regularly refresh model cache
   - Set reasonable token compression thresholds
   - Monitor cache hit rates

3. **API Call Optimization**
   - Use connection reuse and Keep-Alive
   - Set reasonable timeout values
   - Implement retry mechanisms and circuit breakers

## Contributing Guide

We welcome Issue and Pull Request submissions!

### Development Environment Setup

```bash
# Clone the project
git clone https://github.com/liliangshan/openaisdk-proxy.git
cd openaisdk-proxy

# Install dependencies
go mod download
cd frontend && npm install

# Run development build
./build.sh

# Start service
./bin/openaisdk-proxy
```

### Code Standards

- Go code follows `gofmt` standards
- Frontend uses Vue 3 + TypeScript
- Run `go vet` and `gofmt` before committing

## Performance Data

| Scenario | Tokens (Before Optimization) | Tokens (After Optimization) | Cost Savings |
|----------|------------------------------|----------------------------|-------------|
| Typical project conversation | 40,730 | 15,132 | 62.9% |
| Long conversation | 41,945 | 16,347 | 61.0% |
| Real-time code review | 41,024 | 15,426 | 62.4% |

## License

MIT

---

## GitHub Project

- [openaisdk-proxy](https://github.com/liliangshan/openaisdk-proxy)
- [Releases](https://github.com/liliangshan/openaisdk-proxy/releases)
