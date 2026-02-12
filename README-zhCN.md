# OpenAI 接口转发与 Token 精简工具

## 背景与痛点

在大模型应用开发中，上下文窗口的限制和 Token 消耗成本一直是两大核心挑战。当开发者需要处理复杂项目时，往往需要让 AI 读取大量代码文件以理解项目结构。一个中型项目可能包含数十个文件，总代码量达到数万行；每次对话都将这些内容发送给 AI，意味着：

**高昂的成本消耗**：假设每次请求需要处理 5 万 Token，按照 GPT-4 的价格（约 30-60 美元/百万 Token），单次请求成本就高达 1.5-3 美元。如果项目中有多个开发者同时使用，或者需要进行数十轮对话，月度账单轻易达到数百甚至数千美元。

**严重的上下文浪费**：项目中的大量代码是历史遗留、注释说明或重复实现，这些内容对当前任务几乎没有帮助，却占据了宝贵的上下文空间。更糟糕的是，随着对话历史增长，上下文会被旧代码和重复信息填满，导致 AI 无法有效理解最新需求。

**被迫的上下文截断**：当上下文超出限制时，AI 只能被迫截断历史信息，可能丢失关键的代码依赖关系或设计决策记录，影响代码质量和开发效率。

## 解决方案

本工具正是为解决上述痛点而设计。它作为一个智能代理层，位于您的应用与大模型 API 之间，通过以下方式显著降低成本并提升效率：

**Token 截断压缩**：当消息长度超过阈值时，自动移除最早的对话内容，保留最近的 N 轮对话。这是一种简单但有效的压缩策略，避免上下文超出限制而被强制截断。

**灵活的压缩策略**：您可以根据项目特点和需求，自定义触发压缩的阈值、保留的对话轮数、需要保留的消息角色类型等参数，实现精细化控制。

**多模型与多厂商支持**：统一管理不同 AI 厂商和模型配置，通过别名简化 API 调用，无需在业务代码中硬编码复杂的请求参数。

一个轻量级的 OpenAI 兼容接口转发服务，支持 Token 压缩以降低成本。

## 功能特性

### 1. OpenAI 接口转发
- 兼容 OpenAI Chat Completion API
- 支持多种模型配置
- 多厂商、多模型灵活配置
- API 密钥管理

### 2. Token 压缩
- 智能对话压缩
- 自动截断过长消息
- 可配置压缩策略
- 保留上下文完整性

### 3. 管理界面
- 模型管理
- 厂商配置
- 用户界面
- 实时监控

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### 配置说明

```yaml
# 配置文件: cmd/api/config.yaml
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

### 构建与运行

```bash
# 一键构建前后端
./build.sh

# 启动服务
./bin/api
```

### API 使用

```bash
# 发送请求
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "prefix-model-alias",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

## 核心配置

### 厂商配置
| 参数 | 说明 |
|-----|------|
| name | 厂商标识符 |
| display_name | 显示名称 |
| base_url | 接口地址 |
| api_prefix | API 请求前缀 |
| api_key | 厂商密钥 |

### 模型配置
| 参数 | 说明 |
|------|------|
| model_id | 源模型 ID |
| display_name | 模型别名（请求时使用） |
| context_length | 上下文长度（单位 k） |
| compress_enabled | 是否启用压缩 |
| compress_truncate_len | 触发压缩的消息长度阈值 |
| compress_user_count | 保留最近 N 轮对话 |
| compress_role_types | 保留的角色类型（多值用逗号分隔） |

## 压缩策略

### 工作原理
1. 统计请求消息 Token 数量
2. 超过阈值时自动压缩
3. 保留最近的用户和助手对话
4. 可配置压缩参数

### 参数说明
- `compress_enabled`: 是否启用 Token 压缩功能
- `compress_truncate_len`: 消息总长度超过此值时触发压缩（单位：Token）
- `compress_user_count`: 压缩时保留最近 N 轮 user/assistant 对话
- `compress_role_types`: 需要保留的消息角色类型，默认为 user 和 assistant

## License

MIT
