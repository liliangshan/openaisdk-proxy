# OpenAI 接口转发与 Token 精简工具

## 背景与痛点

在大模型应用开发中，上下文窗口的限制和 Token 消耗成本一直是两大核心挑战。当开发者需要处理复杂项目时，往往需要让 AI 读取大量代码文件以理解项目结构。一个中型项目可能包含数十个文件，总代码量达到数万行；每次对话都将这些内容发送给 AI，意味着：

**高昂的成本消耗**：假设每次请求需要处理 5 万 Token，按照 GPT-4 的价格（约 30-60 美元/百万 Token），单次请求成本就高达 1.5-3 美元。如果项目中有多个开发者同时使用，或者需要进行数十轮对话，月度账单轻易达到数百甚至数千美元。

**严重的上下文浪费**：项目中的大量代码是历史遗留、注释说明或重复实现，这些内容对当前任务几乎没有帮助，却占据了宝贵的上下文空间。更糟糕的是，随着对话历史增长，上下文会被旧代码和重复信息填满，导致 AI 无法有效理解最新需求。

**被迫的上下文截断**：当上下文超出限制时，AI 只能被迫截断历史信息，可能丢失关键的代码依赖关系或设计决策记录，影响代码质量和开发效率。

## 解决方案

本工具正是为解决上述痛点而设计。它作为一个智能代理层，位于您的应用与大模型 API 之间，通过以下方式显著降低成本并提升效率：

**Token 长度精简**：当消息总长度超过阈值时，自动精简早期消息中的超长文本内容，保留最近 N 轮对话的完整内容。这是一种高效的压缩策略，在保留对话结构的同时大幅减少 Token 消耗。

**灵活的压缩策略**：您可以根据项目特点和需求，自定义触发压缩的阈值、保留的对话轮数、需要精简的消息角色类型等参数，实现精细化控制。

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
- 自动精简过长文本
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

### 管理界面

启动服务后，可通过浏览器访问管理界面进行配置管理。

**访问地址**：`http://127.0.0.1:8080/user`

**使用流程**：

1. **注册账号**
   - 首次访问管理界面，点击"注册"按钮
   - 填写用户名、邮箱、密码完成注册
   - 注册成功后自动登录

2. **登录系统**
   - 使用注册的账号密码登录
   - 支持 JWT Token 自动续期

3. **获取 API Key**
   - 登录后进入个人中心或设置页面
   - 创建 API Key 用于接口调用
   - 每个用户可创建多个 API Key

4. **配置模型**
   - 添加 AI 厂商配置（名称、API 地址、密钥等）
   - 添加模型配置（选择厂商、模型 ID、设置压缩参数等）
   - 启用 Token 压缩功能降低调用成本

### 构建与运行

```bash
# 一键构建前后端
./build.sh

# 启动服务
./bin/openaisdk-proxy
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
2. 超过阈值时自动精简早期消息的超长文本
3. 保留最近 N 轮对话的完整内容，同时精简更早消息的文本长度
4. 可配置压缩参数

### 参数说明
- `compress_enabled`: 是否启用 Token 压缩功能
- `compress_truncate_len`: 消息总长度超过此值时触发压缩（单位：Token）
- `compress_user_count`: 保留最近 N 轮对话的完整内容，其前的消息会被精简
- `compress_role_types`: 需要精简文本长度的消息角色类型，默认为 user 和 assistant

### 压缩效果示例

假设有一个对话历史包含 10 轮对话，总 Token 数为 100，配置如下：
- `compress_enabled`: true
- `compress_truncate_len`: 10
- `compress_user_count`: 3

系统将：
1. 检测到 100 超过阈值 10（单位为 Token）
2. 保留最近 3 轮用户对话及其对应的助手回复保持完整
3. 精简较早对话中的超长文本（截断过长内容），保留消息结构不删除
4. 将 Token 数压缩至约 10 以内
5. **成本节省**: 原始成本 ÷ 新 Token 数比例 = 成本降低 90%

### 实际运行数据

以下是生产环境中的实际日志数据（2026-02-13）：

```
client IP: 3.209.66.12, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 15132 (原tokens: 40730) 
[CONTEXT] 已截断所有小于第12个user消息的过长文本 (总消息数: 58)

client IP: 52.44.113.131, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 15217 (原tokens: 40815)
[CONTEXT] 已截断所有小于第12个user消息的过长文本 (总消息数: 60)

client IP: 52.44.113.131, model: qn-ch45, model_id: claude-4.5-haiku
body tokens: 16347 (原tokens: 41945)
[CONTEXT] 已截断所有小于第12个user消息的过长文本 (总消息数: 78)
```

**真实压缩效果**：
- 平均成本节省：**62-63%**
- 原始 Token 范围：40,730 - 41,945
- 压缩后 Token 范围：15,132 - 16,347
- 实际成本降低：约 **3.9 倍** 左右

## 项目结构

```
.
├── cmd/api/                    # 后端应用入口
│   ├── main.go                 # 主程序
│   ├── config.yaml             # 配置文件
│   └── internal/
│       ├── handlers/           # HTTP 请求处理
│       │   ├── chat.go         # 聊天接口处理
│       │   ├── models.go       # 模型管理接口
│       │   └── ...
│       ├── service/            # 业务逻辑层
│       ├── repository/         # 数据访问层
│       ├── models/             # 数据模型
│       └── cache/              # 缓存管理
├── frontend/                   # 前端应用
│   ├── src/
│   │   ├── views/              # 页面组件
│   │   ├── components/         # 公共组件
│   │   └── ...
│   └── package.json
├── bin/                        # 构建输出目录
│   └── openaisdk-proxy         # 可执行文件
├── build.sh                    # 一键构建脚本
└── README.md                   # 英文文档
```

## API 文档

### 1. 聊天完成 API

```bash
POST /v1/chat/completions

# 请求体示例
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

# 响应示例
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

### 2. 流式响应

设置 `"stream": true` 以获取流式响应：

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "prefix-model-alias", "messages": [...], "stream": true}'
```

### 3. 常见参数

| 参数 | 类型 | 说明 |
|------|------|------|
| model | string | 模型别名，格式为 `prefix-alias` |
| messages | array | 消息列表，必须包含 role 和 content |
| temperature | float | 生成多样性，范围 0-2，默认 0.7 |
| max_tokens | int | 最大生成 token 数 |
| top_p | float | 核采样参数，范围 0-1 |
| stream | bool | 是否流式返回 |

## 常见问题

### Q: 如何添加新模型？
A: 在管理界面或数据库中配置模型信息，包括模型 ID、别名、厂商等，系统会自动缓存。

### Q: 压缩会丢失重要信息吗？
A: 压缩策略保留最近的对话历史，只删除较早的内容。可以通过调整 `compress_user_count` 参数来控制保留的对话轮数。

### Q: 如何监控 Token 使用情况？
A: 在管理界面查看实时日志和统计数据，或通过 API 响应的 `usage` 字段了解每次请求的消耗。

### Q: 支持哪些 LLM 厂商？
A: 理论上支持所有 OpenAI 兼容的 API，包括但不限于 OpenAI、Azure、Anthropic 等。

### Q: 如何部署到生产环境？
A: 参考下方部署指南，使用 Docker、Kubernetes 或系统服务管理器（如 systemd）来运行。

## 部署指南

### Docker 部署

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

### Systemd 服务配置

创建文件 `/etc/systemd/system/openaisdk-proxy.service`：

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

然后运行：
```bash
sudo systemctl daemon-reload
sudo systemctl enable openaisdk-proxy
sudo systemctl start openaisdk-proxy
```

### 环境变量配置

支持以下环境变量覆盖配置文件：

```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=model_system
API_PORT=8080
JWT_SECRET=your-secret-key
```

## 性能优化建议

1. **数据库优化**
   - 为 `api_keys` 和 `models` 表创建索引
   - 定期清理旧日志数据
   - 使用连接池管理数据库连接

2. **缓存优化**
   - 定期刷新模型缓存
   - 合理设置 Token 压缩阈值
   - 监控缓存命中率

3. **API 调用优化**
   - 使用连接复用和 Keep-Alive
   - 设置合理的超时时间
   - 实现重试机制和熔断保护

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/liliangshan/openaisdk-proxy.git
cd openaisdk-proxy

# 安装依赖
go mod download
cd frontend && npm install

# 运行开发服务
./build.sh

# 启动
./bin/openaisdk-proxy
```

### 代码规范

- Go 代码遵循 `gofmt` 规范
- 前端使用 Vue 3 + TypeScript
- 提交前运行 `go vet` 和 `gofmt`

## 性能数据

| 场景 | Token 数（优化前） | Token 数（优化后） | 成本节省 |
|------|-------------------|-------------------|--------|
| 典型项目对话 | 40,730 | 15,132 | 62.9% |
| 长会话场景 | 41,945 | 16,347 | 61.0% |
| 实时代码审查 | 41,024 | 15,426 | 62.4% |

## License

MIT

---

## GitHub 项目

- [openaisdk-proxy](https://github.com/liliangshan/openaisdk-proxy)
- [Releases](https://github.com/liliangshan/openaisdk-proxy/releases)
