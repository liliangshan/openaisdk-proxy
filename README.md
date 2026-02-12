# OpenAI Proxy & Token Optimization Tool

## Background & Pain Points

In large language model (LLM) application development, context window limitations and token consumption costs have always been two core challenges. When developers need to handle complex projects, they often need to let AI read numerous code files to understand the project structure. A medium-sized project may contain dozens of files with tens of thousands of lines of code; sending all this content to AI for every conversation means:

**High cost consumption**: Assuming each request needs to process 50,000 tokens, at GPT-4 pricing (approximately $30-60 per million tokens), a single request costs $1.5-3. If multiple developers are using the project simultaneously or dozens of conversation rounds are needed, monthly bills can easily reach hundreds or even thousands of dollars.

**Severe context waste**: A large amount of code in projects is legacy, explanatory comments, or duplicate implementations. These contents have little relevance to the current task but occupy valuable context space. Worse, as conversation history grows, context gets filled with old code and duplicate information, causing AI to fail to effectively understand the latest requirements.

**Forced context truncation**: When context exceeds limits, AI can only forcibly truncate historical information, potentially losing critical code dependencies or design decision records, affecting code quality and development efficiency.

## Solutions

This tool is designed to solve the pain points mentioned above. It acts as an intelligent proxy layer between your application and the LLM API, significantly reducing costs and improving efficiency through:

**Token Truncation Compression**: When message length exceeds the threshold, automatically remove the earliest conversation content, retaining the most recent N rounds of dialogue. This is a simple but effective compression strategy that avoids forced truncation when context exceeds limits.

**Flexible Compression Policies**: You can customize compression triggers, the number of dialogue rounds to retain, message role types to preserve, and other parameters based on project characteristics and requirements, achieving fine-grained control.

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
- Automatic truncation of overly long messages
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

### Build & Run

```bash
# Build frontend and backend
./build.sh

# Start service
./bin/api
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
2. Automatically compress when exceeding threshold
3. Retain recent user and assistant conversations
4. Configurable compression parameters

### Parameter Description
- `compress_enabled`: Whether to enable token compression
- `compress_truncate_len`: Trigger compression when message length exceeds this value (unit: Token)
- `compress_user_count`: Retain the most recent N rounds of user/assistant dialogue during compression
- `compress_role_types`: Message role types to retain, defaulting to user and assistant

## License

MIT
