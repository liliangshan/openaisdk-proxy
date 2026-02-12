package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	certFile = "ssl.pem"
	keyFile  = "ssl.key"
)

// Config 应用配置
type Config struct {
	App struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Domain    string `yaml:"domain"`
		CertPort  int    `yaml:"cert_port"`
	} `yaml:"app"`
	SSL struct {
		Email     string `yaml:"email"`
		CertDir   string `yaml:"cert_dir"`
		Challenge string `yaml:"challenge"`
	} `yaml:"ssl"`
}

// loadConfig 加载配置文件
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if cfg.App.CertPort == 0 {
		cfg.App.CertPort = 80
	}
	if cfg.SSL.CertDir == "" {
		cfg.SSL.CertDir = "./cert"
	}
	if cfg.SSL.Challenge == "" {
		cfg.SSL.Challenge = "http"
	}

	return &cfg, nil
}

func main() {
	// 定义命令行参数
	action := flag.String("action", "", "操作类型: create(创建证书), renew(续期证书), revoke(吊销证书)")
	staging := flag.Bool("staging", false, "使用测试环境（不产生真实证书）")
	configFile := flag.String("config", "config.yaml", "配置文件路径")
	help := flag.Bool("help", false, "显示帮助信息")

	flag.Usage = func() {
		fmt.Println("Let's Encrypt 证书管理工具")
		fmt.Println()
		fmt.Println("使用方法:")
		fmt.Println("  cert-tool -action=create")
		fmt.Println("  cert-tool -action=renew")
		fmt.Println()
		fmt.Println("参数说明:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("从配置文件读取:")
		fmt.Println("  -domain: 从 config.yaml 中的 app.domain 读取")
		fmt.Println("  -email: 从 config.yaml 中的 ssl.email 读取")
		fmt.Println("  -challenge: 从 config.yaml 中的 ssl.challenge 读取")
		fmt.Println("  -port: 从 config.yaml 中的 app.cert_port 读取")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  创建证书: cert-tool -action=create")
		fmt.Println("  测试模式: cert-tool -action=create -staging")
		fmt.Println("  续期证书: cert-tool -action=renew")
		fmt.Println("  吊销证书: cert-tool -action=revoke")
	}

	flag.Parse()

	if *help || *action == "" {
		flag.Usage()
		os.Exit(0)
	}

	// 验证操作类型
	validActions := map[string]bool{
		"create": true,
		"renew": true,
		"revoke": true,
	}
	if !validActions[*action] {
		log.Fatalf("错误: 无效的操作类型 '%s'，支持的操作: create, renew, revoke", *action)
	}

	// 加载配置
	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证域名
	if cfg.App.Domain == "" {
		log.Fatal("错误: config.yaml 中未配置 app.domain")
	}

	// 验证邮箱
	if cfg.SSL.Email == "" {
		log.Fatal("错误: config.yaml 中未配置 ssl.email")
	}

	// 创建证书目录
	certPath := cfg.SSL.CertDir
	if err := os.MkdirAll(certPath, 0755); err != nil {
		log.Fatalf("创建证书目录失败: %v", err)
	}

	// 执行相应操作
	switch *action {
	case "create":
		createCertificate(cfg)
	case "renew":
		renewCertificate(cfg)
	case "revoke":
		revokeCertificate(cfg)
	}
}

// createCertificate 创建 Let's Encrypt 证书
func createCertificate(cfg *Config) {
	log.Printf("开始为域名 %s 创建证书（验证方式: %s）...", cfg.App.Domain, cfg.SSL.Challenge)

	certPath := cfg.SSL.CertDir

	if cfg.SSL.Challenge == "http" {
		createCertificateHTTP(cfg, certPath)
	} else {
		createCertificateDNS(cfg, certPath)
	}
}

// createCertificateHTTP 使用文件验证模式创建证书
func createCertificateHTTP(cfg *Config, certPath string) {
	args := []string{
		"certonly",
		"--manual",
		"--preferred-challenges", "http",
		"--agree-tos",
		"--email", cfg.SSL.Email,
		"-d", cfg.App.Domain,
	}

	if *staging {
		args = append(args, "--staging")
		log.Println("注意: 使用测试环境，生成的证书无效")
	}

	// 输出路径
	args = append(args,
		"--cert-path", filepath.Join(certPath, certFile),
		"--key-path", filepath.Join(certPath, keyFile),
	)

	// 检查 certbot 是否安装
	if err := exec.Command("which", "certbot").Run(); err != nil {
		log.Fatal("错误: certbot 未安装，请先安装: sudo apt install certbot")
	}

	log.Printf("执行命令: certbot %s", strings.Join(args, " "))

	// 创建临时目录存放验证文件
	tmpDir, err := os.MkdirTemp("", "certbot-")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 设置环境变量
	os.Setenv("CERTBOT_VALIDATION", tmpDir)
	os.Setenv("CERTBOT_HTTP01_ADDRESS", fmt.Sprintf(":%d", cfg.App.CertPort))

	// 启动文件验证服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			filePath := filepath.Join(tmpDir, r.URL.Path)
			data, err := os.ReadFile(filePath)
			if err != nil {
				http.Error(w, "文件不存在", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(data)
			return
		}
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.CertPort),
		Handler: mux,
	}

	go func() {
		log.Printf("文件验证服务器已启动: http://localhost:%d", cfg.App.CertPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器错误: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 执行 certbot
	cmd := exec.Command("certbot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Printf("证书创建失败: %v", err)
		log.Println("提示: 请确保 80 端口已开放，且域名已解析到当前服务器")
		server.Close()
		return
	}

	server.Close()

	log.Printf("证书创建成功！")
	log.Printf("证书文件: %s", filepath.Join(certPath, certFile))
	log.Printf("密钥文件: %s", filepath.Join(certPath, keyFile))
}

// createCertificateDNS 使用 DNS 验证模式创建证书
func createCertificateDNS(cfg *Config, certPath string) {
	args := []string{
		"certonly",
		"--manual",
		"--preferred-challenges", "dns",
		"--dns-cloudflare",
		"--dns-cloudflare-credentials", "/dev/stdin",
		"--agree-tos",
		"--email", cfg.SSL.Email,
		"-d", cfg.App.Domain,
	}

	if *staging {
		args = append(args, "--staging")
		log.Println("注意: 使用测试环境，生成的证书无效")
	}

	args = append(args,
		"--cert-path", filepath.Join(certPath, certFile),
		"--key-path", filepath.Join(certPath, keyFile),
	)

	log.Printf("执行命令: certbot %s", strings.Join(args, " "))

	log.Println("请提供 Cloudflare API 密钥文件内容（Ctrl+D 结束）:")
	cmd := exec.Command("certbot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("证书创建失败: %v", err)
	}

	log.Printf("证书创建成功！")
	log.Printf("证书文件: %s", filepath.Join(certPath, certFile))
	log.Printf("密钥文件: %s", filepath.Join(certPath, keyFile))
}

// renewCertificate 续期证书
func renewCertificate(cfg *Config) {
	log.Printf("开始续期域名 %s 的证书...", cfg.App.Domain)

	certPath := cfg.SSL.CertDir

	args := []string{
		"renew",
		"--cert-path", filepath.Join(certPath, certFile),
		"--key-path", filepath.Join(certPath, keyFile),
	}

	if *staging {
		args = append(args, "--staging")
	}

	cmd := exec.Command("certbot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("续期失败或无需续期: %v", err)
	}

	log.Printf("证书续期处理完成！")
}

// revokeCertificate 吊销证书
func revokeCertificate(cfg *Config) {
	log.Printf("开始吊销域名 %s 的证书...", cfg.App.Domain)

	certPath := cfg.SSL.CertDir

	args := []string{
		"revoke",
		"--cert-path", filepath.Join(certPath, certFile),
		"--key-path", filepath.Join(certPath, keyFile),
	}

	if *staging {
		args = append(args, "--staging")
	}

	cmd := exec.Command("certbot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("证书吊销失败: %v", err)
	}

	log.Printf("证书吊销处理完成！")
}
