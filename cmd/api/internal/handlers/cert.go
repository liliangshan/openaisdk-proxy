package handlers

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// CertVerifyResult 证书验证结果
type CertVerifyResult struct {
	Valid        bool   `json:"valid"`
	Subject      string `json:"subject"`
	Issuer       string `json:"issuer"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
	DaysLeft     int    `json:"days_left"`
	Algorithm    string `json:"algorithm"`
	SerialNumber string `json:"serial_number"`
	Error        string `json:"error,omitempty"`
}

// VerifyCert 验证证书文件
// POST /cert/verify
func (h *Handler) VerifyCert(c echo.Context) error {
	var req struct {
		CertPath string `json:"cert_path"` // 证书文件路径
		CertData string `json:"cert_data"` // PEM格式的证书内容
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误",
		})
	}

	var result CertVerifyResult

	// 优先使用CertData（上传的PEM内容）
	if req.CertData != "" {
		return verifyCertFromData(req.CertData, &result, c)
	}

	// 否则使用CertPath（证书文件路径）
	if req.CertPath == "" {
		result.Valid = false
		result.Error = "证书路径不能为空"
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: result.Error,
			Data:    result,
		})
	}

	return verifyCertFromPath(req.CertPath, &result, c)
}

// verifyCertFromPath 从文件路径验证证书
func verifyCertFromPath(certPath string, result *CertVerifyResult, c echo.Context) error {
	// 读取证书文件
	certData, err := os.ReadFile(certPath)
	if err != nil {
		result.Valid = false
		result.Error = "读取证书文件失败: " + err.Error()
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: result.Error,
			Data:    result,
		})
	}

	return parseAndVerifyCert(certData, result, c)
}

// verifyCertFromData 从PEM数据验证证书
func verifyCertFromData(certData string, result *CertVerifyResult, c echo.Context) error {
	return parseAndVerifyCert([]byte(certData), result, c)
}

// parseAndVerifyCert 解析并验证证书
func parseAndVerifyCert(certData []byte, result *CertVerifyResult, c echo.Context) error {
	// 解码PEM格式
	block, _ := pem.Decode(certData)
	if block == nil {
		result.Valid = false
		result.Error = "无效的PEM格式证书"
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: result.Error,
			Data:    result,
		})
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		result.Valid = false
		result.Error = "解析证书失败: " + err.Error()
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: result.Error,
			Data:    result,
		})
	}

	// 填充验证结果
	result.Subject = cert.Subject.CommonName
	if result.Subject == "" {
		result.Subject = cert.Subject.String()
	}
	result.Issuer = cert.Issuer.CommonName
	if result.Issuer == "" {
		result.Issuer = cert.Issuer.String()
	}

	result.NotBefore = cert.NotBefore.Format("2006-01-02 15:04:05")
	result.NotAfter = cert.NotAfter.Format("2006-01-02 15:04:05")

	// 计算剩余天数
	now := time.Now()
	daysLeft := int(cert.NotAfter.Sub(now).Hours() / 24)
	result.DaysLeft = daysLeft

	// 获取签名算法
	result.Algorithm = cert.SignatureAlgorithm.String()

	// 获取序列号
	result.SerialNumber = cert.SerialNumber.String()

	// 检查证书是否过期
	if now.Before(cert.NotBefore) {
		result.Valid = false
		result.Error = "证书尚未生效，有效期从 " + result.NotBefore + " 开始"
	} else if now.After(cert.NotAfter) {
		result.Valid = false
		result.Error = "证书已过期，有效期至 " + result.NotAfter
	} else if daysLeft < 0 {
		result.Valid = false
		result.Error = "证书已过期"
	} else {
		result.Valid = true
	}

	return c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "验证完成",
		Data:    result,
	})
}

// GetCertInfo 获取证书信息（不验证）
// GET /cert/info
func (h *Handler) GetCertInfo(c echo.Context) error {
	certPath := c.QueryParam("path")
	certData := c.QueryParam("data")

	var result CertVerifyResult

	if certData != "" {
		return verifyCertFromData(certData, &result, c)
	}

	if certPath == "" {
		result.Error = "证书路径不能为空"
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: result.Error,
		})
	}

	return verifyCertFromPath(certPath, &result, c)
}
