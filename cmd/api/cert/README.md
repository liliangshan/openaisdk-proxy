# Let's Encrypt 证书管理工具

## 快速开始

```bash
cd cmd/api/cert
go build -o ../cert-tool main.go
```

然后在 `cmd/api` 目录下执行：

```bash
# 创建证书
../cert-tool -action=create

# 续期证书
../cert-tool -action=renew

# 吊销证书
../cert-tool -action=revoke
```

## 配置方式

所有配置都在 `cmd/api/config.yaml` 中：

```yaml
# 应用配置
app:
  host: "0.0.0.0"
  port: 28080
  domain: "ai.luobbin.cn"    # ✅ SSL证书域名
  cert_port: 80              # ✅ 文件验证端口

# SSL证书配置
ssl:
  email: "admin@luobbin.cn"   # 证书通知邮箱
  cert_dir: "./cert"        # 证书保存目录
  challenge: "http"          # 验证方式: dns 或 http
```

## 使用方法

```bash
# 创建证书（从配置文件读取域名、邮箱、验证方式）
cert-tool -action=create

# 测试模式（不产生真实证书）
cert-tool -action=create -staging

# 续期证书
cert-tool -action=renew

# 吊销证书
cert-tool -action=revoke

# 指定配置文件
cert-tool -action=create -config=other.yaml
```

## 验证方式

| 方式 | 配置 | 说明 |
|------|------|------|
| 文件验证 | `challenge: http` | 需要 80 端口开放 |
| DNS 验证 | `challenge: dns` | 需要 Cloudflare API |

## 输出文件

证书保存在 `cmd/api/cert` 目录：
```
cmd/api/cert/
├── ssl.pem   # 证书文件
└── ssl.key   # 密钥文件
```

## Docker

```bash
cd cmd/api/cert
docker build -t cert-tool .

# 运行
docker run --rm -p 80:80 \
  -v $(pwd)/../cert:/app/cert \
  cert-tool \
  -action=create
```

## 前置要求

```bash
# 安装 certbot
sudo apt install certbot

# DNS 验证需要 Cloudflare 插件
pip install certbot-dns-cloudflare
```

## 自动续期

```bash
# Cron 定时任务
0 0 * * * /path/to/cert-tool -action=renew
```

## 注意事项

1. 测试时使用 `-staging` 避免速率限制
2. 文件验证需要 80 端口开放
3. DNS 验证需要 Cloudflare API 密钥
4. Let's Encrypt 证书有效期 90 天
