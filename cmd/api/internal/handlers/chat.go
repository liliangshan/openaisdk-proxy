package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/model-system/api/internal/cache"
	"github.com/tiktoken-go/tokenizer"
)

// 全局 tokenizer 实例
var globalTokenizer tokenizer.Codec

// 全局 HTTP 客户端，复用连接池
var globalHTTPClient *http.Client

func init() {
	// 初始化全局 HTTP 客户端，配置连接池参数
	transport := &http.Transport{
		MaxIdleConns:        100,              // 全局空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个 host 的空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	// 注意：SSE 流式请求不能设置整体超时，由 context 控制
	// 这里设置的超时只适用于普通请求
	globalHTTPClient = &http.Client{
		Transport: transport,
		Timeout:   300 * time.Second, // 调整为 5 分钟，用于普通请求
	}
}

// SetTokenizer 设置全局 tokenizer 实例
func SetTokenizer(tk tokenizer.Codec) {
	globalTokenizer = tk
}

// writeDynamicFile 写入动态调试文件
func writeDynamicFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// ChatCompletionRequest OpenAI 兼容的聊天补全请求
// 固定字段 + Extra 透传未知字段（如 tools、tool_choice 等）
type ChatCompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    json.RawMessage        `json:"messages"`
	Stream      bool                   `json:"stream"`
	Temperature *float64               `json:"temperature,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Extra       map[string]interface{} `json:"-"` // 其他未知字段，序列化时一并输出
}

// UnmarshalJSON 自定义反序列化，保留所有未知字段到 Extra
func (r *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	type Alias ChatCompletionRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var full map[string]interface{}
	if err := json.Unmarshal(data, &full); err != nil {
		return err
	}
	knownFields := map[string]bool{
		"model": true, "messages": true, "stream": true,
		"temperature": true, "max_tokens": true,
	}
	r.Extra = make(map[string]interface{})
	for key, value := range full {
		if !knownFields[key] {
			r.Extra[key] = value
		}
	}
	return nil
}

// MarshalJSON 自定义序列化，先输出固定字段，再输出 Extra（类似 ... 展开）
func (r ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	result["model"] = r.Model
	result["messages"] = r.Messages
	result["stream"] = r.Stream
	if r.Temperature != nil {
		result["temperature"] = r.Temperature
	}
	if r.MaxTokens != nil {
		result["max_tokens"] = r.MaxTokens
	}
	for key, value := range r.Extra {
		result[key] = value
	}
	return json.Marshal(result)
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role       string                 `json:"role"`
	Name       string                 `json:"name,omitempty"`
	Content    json.RawMessage        `json:"content"`
	ToolCallID *string                `json:"tool_call_id,omitempty"`
	ToolCalls  *json.RawMessage       `json:"tool_calls,omitempty"`
	Extra      map[string]interface{} `json:"-"` // 其他未知字段
}

// UnmarshalJSON 自定义反序列化，保留所有未知字段
func (cm *ChatMessage) UnmarshalJSON(data []byte) error {
	// 先定义一个临时结构体，包含所有已知字段
	type Alias ChatMessage
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(cm),
	}

	// 反序列化到临时结构体
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析整个JSON，提取未知字段
	var full map[string]interface{}
	if err := json.Unmarshal(data, &full); err != nil {
		return err
	}

	// 已知字段列表
	knownFields := map[string]bool{
		"role":          true,
		"name":          true,
		"content":       true,
		"tool_call_id":  true,
		"tool_calls":    true,
		"function_call": true, // 旧格式的函数调用
	}

	// 提取未知字段
	cm.Extra = make(map[string]interface{})
	for key, value := range full {
		if !knownFields[key] {
			cm.Extra[key] = value
		}
	}

	return nil
}

// MarshalJSON 自定义序列化，包含所有字段
func (cm ChatMessage) MarshalJSON() ([]byte, error) {
	// 转换为 map
	result := make(map[string]interface{})

	// 添加已知字段
	if cm.Role != "" {
		result["role"] = cm.Role
	}
	if cm.Name != "" {
		result["name"] = cm.Name
	}
	result["content"] = json.RawMessage(cm.Content)
	if cm.ToolCallID != nil {
		result["tool_call_id"] = cm.ToolCallID
	}
	if cm.ToolCalls != nil {
		result["tool_calls"] = cm.ToolCalls
	}

	// 添加未知字段
	for key, value := range cm.Extra {
		result[key] = value
	}

	return json.Marshal(result)
}

// MarshalMessagesToJSON 将 []ChatMessage 序列化为 json.RawMessage，确保 Extra 完整
func MarshalMessagesToJSON(messages []ChatMessage) json.RawMessage {
	if len(messages) == 0 {
		return json.RawMessage("[]")
	}

	// 直接序列化，复用 ChatMessage.MarshalJSON 逻辑
	resultBytes, err := json.Marshal(messages)
	if err != nil {
		// 如果序列化失败，返回空数组
		return json.RawMessage("[]")
	}
	return resultBytes
}

// ChatCompletionResponse 聊天补全响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatStreamChunk 流式响应块
type ChatStreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice 流式选择
type StreamChoice struct {
	Index int `json:"index"`
	Delta struct {
		Role    string `json:"role,omitempty"`
		Content string `json:"content,omitempty"`
	} `json:"delta"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

// countMessageTokens 计算单条消息的 token 数量
func countMessageTokens(msg ChatMessage, enc tokenizer.Codec) int {
	tokens := 0

	// 计算 role 的 token 数
	roleTokens, _, _ := enc.Encode(msg.Role)
	tokens += len(roleTokens)

	// 添加角色标记的 token（OpenAI 格式）
	tokens += 1 // role 标签

	// 计算 content 的 token 数
	content := string(msg.Content)
	contentTokens, _, _ := enc.Encode(content)
	tokens += len(contentTokens)

	// 添加内容标记的 token
	tokens += 2 // content 标签开始和结束

	// 添加消息结束标记
	tokens += 1 // <|im_end|> 或等价标记

	return tokens
}

// countMessagesTokens 计算 messages 的 token 总数
func countMessagesTokens(messages []ChatMessage) int {
	if globalTokenizer == nil {
		// 回退：简单估算
		total := 0
		for _, msg := range messages {
			total += len(msg.Role) + len(string(msg.Content)) + 10
		}
		return total
	}

	total := 0
	for _, msg := range messages {
		total += countMessageTokens(msg, globalTokenizer)
	}
	return total
}

// truncateLongTexts 截断过长文本
// 找到倒数第N个 user 角色的消息索引，截断从模型第一次调用工具到该消息之间指定类型消息的过长text
// 返回修改后的消息和日志字符串
func truncateLongTexts(messages []ChatMessage, userCount int, truncateLen int, roleTypes string) ([]ChatMessage, string) {
	// 解析角色类型配置，用于截断这些类型消息的 text
	targetRoles := map[string]bool{}
	if roleTypes == "" {
		// 默认截断 user/assistant/tool 类型的消息
		targetRoles["user"] = true
		targetRoles["assistant"] = true
		targetRoles["tool"] = true
	} else {
		for _, role := range strings.Split(roleTypes, ",") {
			role = strings.TrimSpace(role)
			if role != "" {
				targetRoles[role] = true
			}
		}
	}

	// 从前往后找到第一个包含 ToolCalls（模型决定调用工具）的 assistant 消息索引
	targetToolIndex := -1
	for i := 0; i < len(messages); i++ {
		if messages[i].Role == "assistant" && messages[i].ToolCalls != nil {
			targetToolIndex = i
			break
		}
	}

	// 从后往前找到第N个 user 角色的消息，获取其索引
	count := 0
	targetMsgIndex := -1
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			count++
			if count == userCount {
				targetMsgIndex = i
				break
			}
		}
	}

	if targetMsgIndex == -1 {
		return messages, ""
	}
	// 如果没有工具调用或工具调用在目标消息之后，则无需截断
	if targetToolIndex == -1 || targetToolIndex > targetMsgIndex {
		return messages, ""
	}

	// 循环所有消息，截断在 targetToolIndex 和 targetMsgIndex 之间的消息的过长 text
	modified := false
	for i := range messages {
		if messages[i].Role == "system" {
			continue
		}
		if i >= targetMsgIndex {
			continue // 只截断小于目标索引的消息
		}
		if i < targetToolIndex {
			continue // 只截断大于等于工具调用索引的消息
		}
		// 只截断指定类型消息的 text
		if !targetRoles[messages[i].Role] {
			continue
		}

		// 解析 content
		var content []map[string]interface{}
		contentStr := string(messages[i].Content)
		if err := json.Unmarshal([]byte(contentStr), &content); err != nil {
			continue
		}

		for _, item := range content {
			if text, ok := item["text"].(string); ok {
				if len(text) > truncateLen {
					item["text"] = text[:truncateLen]
					modified = true
				}
			}
		}

		if modified {
			newContent, err := json.Marshal(content)
			if err != nil {
				continue
			}
			messages[i].Content = newContent
		}
	}

	var logMsg string
	if modified {
		roleLabel := "user"
		if roleTypes != "" {
			roleLabel = roleTypes
		}
		logMsg = fmt.Sprintf("[CONTEXT] 已截断所有小于第%d个%s消息的过长文本 (总消息数: %d)", targetMsgIndex, roleLabel, len(messages))
	}

	return messages, logMsg
}

// ChatCompletion 聊天补全处理函数
// POST /api/v1/chat/completions
func (h *Handler) ChatCompletion(c echo.Context) error {
	// 从请求头获取API密钥
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		log.Printf("[ERROR] 请求头中没有Authorization")
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "Invalid API key",
		})
	}

	// 提取Bearer token
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		log.Printf("[ERROR] 请求头中没有Bearer token")
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "Invalid API key",
		})
	}

	// 检查API密钥是否在全局缓存中，并获取用户ID
	userID, exists := cache.GetCache().GetUserIDByAPIKey(apiKey)
	if !exists {
		log.Printf("[ERROR] API密钥不存在")
		return c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "Invalid API key",
		})
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("[ERROR] 读取请求体失败: %v", err)
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "读取请求体失败",
		})
	}
	// 解析请求
	var req ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("[ERROR] 解析请求失败: %v", err)
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "解析请求失败",
		})
	}

	// 解析 messages 为 []ChatMessage，确保 Extra 完整
	var messages []ChatMessage
	if err := json.Unmarshal(req.Messages, &messages); err != nil {
		log.Printf("[WARN] 解析 messages 失败: %v", err)
	}

	// 调试模式：输出请求内容的最后一条消息
	if h.cfg.Debug && len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		if lastMsgJSON, err := json.Marshal(lastMsg); err == nil {
			log.Printf("[DEBUG] 最后一条消息: %s", string(lastMsgJSON))
		}
	}

	// 验证模型参数
	if req.Model == "" {
		log.Printf("[ERROR] 模型参数不能为空")
		return c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "模型参数不能为空",
		})
	}

	// 从缓存中查找模型
	modelItem, err := h.findModelByName(req.Model)
	if err != nil {
		log.Printf("[ERROR] 模型不存在: %v", err)
		return c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
	}

	// 检查模型是否属于该API密钥的用户
	if modelItem.Model.UserID != userID {
		log.Printf("[ERROR] 模型不属于该API密钥的用户")
		return c.JSON(http.StatusForbidden, Response{
			Code:    403,
			Message: "This model does not belong to your account",
		})
	}

	// 定义日志附加信息字符串
	var logExtra string

	// 在判断截断之前计算原始 token 数
	originalTokenCount := countMessagesTokens(messages)

	// 根据模型配置压缩/截断消息
	if modelItem.Model.CompressEnabled {
		// 截断过长文本：找到指定数量 user 角色的消息，截断其content中过长的text
		truncatedMessages, truncateLog := truncateLongTexts(messages, modelItem.Model.CompressUserCount, modelItem.Model.CompressTruncateLen, modelItem.Model.CompressRoleTypes)
		messages = truncatedMessages
		if truncateLog != "" {
			logExtra += " " + truncateLog
		}
	}

	// 计算 token 数
	tokenCount := countMessagesTokens(messages)

	// 输出请求日志
	log.Printf("client IP: %s, model: %s, model_id: %s, body tokens: %d (原tokens: %d)%s", c.RealIP(), req.Model, modelItem.Model.ModelID, tokenCount, originalTokenCount, logExtra)

	// 获取提示词（三级优先级：API Key工具提示词 > API Key默认提示词 > 模型 > 厂商）
	prompt := ""

	// 从 req.Extra["tools"] 中提取工具名
	toolName := extractToolNameFromExtra(req.Extra)

	if toolName != "" {
		// 如果有工具名，优先使用工具提示词
		if toolPrompt, ok := cache.GetCache().GetAPIKeyPromptByTool(apiKey, toolName); ok && toolPrompt != "" {
			prompt = toolPrompt
		}
	}

	// 如果没有找到工具提示词，使用默认优先级
	if prompt == "" {
		if apiKeyPrompt, ok := cache.GetCache().GetAPIKeyPrompt(apiKey); ok && apiKeyPrompt != "" {
			// API Key 级别提示词（最高优先级）
			prompt = apiKeyPrompt
		}
	}

	// 如果有提示词，添加到消息数组最后，并添加缓存功能
	// 但如果最后一条 user 消息已经是工具执行结果，则不追加提示词
	// 直接使用 messages
	if len(messages) > 0 {
		// 检测最后一条消息的 content 是否包含 "user_query"
		if isUserQueryFromMessages(messages) {
			// 获取 API Key 的所有工具提示词
			allPrompts := cache.GetCache().GetAPIKeyAllPrompts(apiKey)

			// 提取请求中的工具列表转成字符串，用于检查是否包含工具名
			toolsStr := extractToolsStringFromExtra(req.Extra)

			// 最终要追加的提示词列表
			var promptsToAppend []string

			for toolName, toolPrompt := range allPrompts {
				if toolPrompt == "" {
					continue
				}

				if toolName == "" {
					// 工具名为空，直接加入
					promptsToAppend = append(promptsToAppend, toolPrompt)
				} else {
					// 工具名不为空，检查工具列表是否包含该工具名
					if strings.Contains(toolsStr, toolName) {
						promptsToAppend = append(promptsToAppend, toolPrompt)
					}
				}
			}

			// 如果没有匹配的工具提示词，使用之前的默认提示词逻辑
			if len(promptsToAppend) == 0 && prompt != "" {
				promptsToAppend = append(promptsToAppend, prompt)
			}

			// 追加所有匹配的提示词
			if len(promptsToAppend) > 0 {
				for _, p := range promptsToAppend {
					// 创建带缓存功能的提示词消息
					promptContent := []map[string]interface{}{
						{
							"type":          "text",
							"text":          p,
							"cache_control": map[string]string{"type": "ephemeral"},
						},
					}

					// 创建 ChatMessage
					contentBytes, _ := json.Marshal(promptContent)
					promptMsg := ChatMessage{
						Role:    "user",
						Content: contentBytes,
					}
					// 将提示词作为 user 消息追加到消息数组最后
					messages = append(messages, promptMsg)
				}
			}
		}
	}

	// 准备转发到厂商的请求
	providerURL := modelItem.ProviderBaseURL + "/chat/completions"
	providerKey := modelItem.ProviderKey

	// 更新 messages 和 model
	req.Messages = MarshalMessagesToJSON(messages)
	req.Model = modelItem.Model.ModelID

	// 直接序列化 req（包含所有已修改的消息和 Extra 字段）
	providerReqBody, err := req.MarshalJSON()
	if err != nil {
		log.Printf("[ERROR] 序列化请求失败: %v", err)
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "序列化请求失败",
		})
	}

	if h.cfg.Debug {
		if err := writeDynamicFile("providerReqBody.json", providerReqBody); err != nil {
			log.Printf("[WARN] 写入调试文件失败: %v", err)
		} else {
			log.Printf("[DEBUG] 序列化请求体: providerReqBody.json")
		}
	}

	// 发送请求到厂商
	providerReq, err := http.NewRequest("POST", providerURL, bytes.NewReader(providerReqBody))
	if err != nil {
		log.Printf("[ERROR] 创建请求失败: %v", err)
		return c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建请求失败",
		})
	}

	// 设置请求头
	providerReq.Header.Set("Content-Type", "application/json")
	providerReq.Header.Set("Authorization", "Bearer "+providerKey)

	// 发送请求
	resp, err := globalHTTPClient.Do(providerReq)
	if err != nil {
		log.Printf("[ERROR] 请求厂商失败: %v", err)
		return c.JSON(http.StatusBadGateway, Response{
			Code:    502,
			Message: "请求厂商失败: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] 厂商返回错误 (status: %d): %s", resp.StatusCode, string(respBody))
		return c.JSON(http.StatusBadGateway, Response{
			Code:    502,
			Message: "厂商返回错误: " + string(respBody),
		})
	}

	// 如果不流式，直接返回响应
	if !req.Stream {
		respBody, _ := io.ReadAll(resp.Body)
		// 直接返回厂商的响应
		c.Response().Header().Set("Content-Type", "application/json")
		return c.String(http.StatusOK, string(respBody))
	}

	// 流式响应
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Transfer-Encoding", "chunked")

	// 先发送 HTTP 状态码 200 给客户端
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	// 使用 bufio.Reader 实时读取并转发每个数据块
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 正常结束
				break
			}
			// 连接错误
			return nil
		}

		// 检测是否包含 ListMcpResources 工具调用
		lineStr := string(line)
		if strings.Contains(lineStr, "ListMcpResources") {
			// 关闭当前流式响应
			resp.Body.Close()

			// 将 MCP 工具列表信息作为字符串追加到消息中
			mcpToolsInfo := "MCP工具列表查询：以下是可用的 MCP 工具函数列表"

			// 更新请求体中的消息，追加 MCP 工具信息
			var reqData map[string]interface{}
			if err := json.Unmarshal(body, &reqData); err == nil {
				if messages, ok := reqData["messages"].([]interface{}); ok {
					// 创建提示词消息
					promptContent := []map[string]interface{}{
						{
							"type":          "text",
							"text":          mcpToolsInfo,
							"cache_control": map[string]string{"type": "ephemeral"},
						},
					}

					// 创建 user 消息
					userMsg := map[string]interface{}{
						"role":    "user",
						"content": promptContent,
					}
					// 将提示词作为 user 消息追加到消息数组最后
					messages = append(messages, userMsg)
					reqData["messages"] = messages

					// 重新编码请求体
					if newBody, err := json.Marshal(reqData); err == nil {
						body = newBody
					}
				}
			}

			// 重新请求模型
			h.sendProviderRequest(c, modelItem, req)
			return nil
		}

		// 转发数据块到客户端
		if _, writeErr := c.Response().Writer.Write([]byte(line)); writeErr != nil {
			// 客户端断开连接
			return nil
		}
		c.Response().Flush() // 强制立即发送
	}

	return nil
}

// findModelByName 根据厂商前缀-模型ID查找模型
func (h *Handler) findModelByName(modelName string) (*cache.ModelCacheItem, error) {
	cache := h.modelService.GetCache()

	// 直接使用 modelName 作为完整缓存键查找（如 "mm-gpt-4"）
	item, ok := cache.GetModelByCacheKey(modelName)
	if !ok {
		return nil, fmt.Errorf("模型不存在: %s", modelName)
	}

	return item, nil
}

// isUserQueryFromMessages 检查 messages 是否需要追加提示词
func isUserQueryFromMessages(messages []ChatMessage) bool {
	if len(messages) == 0 {
		return false
	}

	lastMsg := messages[len(messages)-1]
	var content []map[string]interface{}
	if err := json.Unmarshal(lastMsg.Content, &content); err != nil {
		return false
	}

	contentStr, _ := json.Marshal(content)
	return strings.Contains(string(contentStr), "user_query")
}

// extractToolNameFromExtra 从 Extra 中提取第一个工具名
func extractToolNameFromExtra(extra map[string]interface{}) string {
	toolsRaw, ok := extra["tools"]
	if !ok {
		return ""
	}

	toolsArr, ok := toolsRaw.([]interface{})
	if !ok {
		return ""
	}

	for _, tool := range toolsArr {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		// 兼容 function 类型的工具（OpenAI 格式）
		if fn, ok := toolMap["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				return name
			}
		}

		// 兼容直接写 name 的工具
		if name, ok := toolMap["name"].(string); ok {
			return name
		}
	}
	return ""
}

// extractToolsStringFromExtra 从 Extra 中提取所有工具名，返回逗号分隔字符串
func extractToolsStringFromExtra(extra map[string]interface{}) string {
	var toolNames []string

	toolsRaw, ok := extra["tools"]
	if !ok {
		return ""
	}

	toolsArr, ok := toolsRaw.([]interface{})
	if !ok {
		return ""
	}

	for _, tool := range toolsArr {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		// 兼容 function 类型的工具（OpenAI 格式）
		if fn, ok := toolMap["function"].(map[string]interface{}); ok {
			if name, ok := fn["name"].(string); ok {
				toolNames = append(toolNames, name)
			}
		}

		// 兼容直接写 name 的工具
		if name, ok := toolMap["name"].(string); ok {
			toolNames = append(toolNames, name)
		}
	}

	return strings.Join(toolNames, ",")
}

// sendProviderRequest 发送请求到厂商并处理响应
func (h *Handler) sendProviderRequest(c echo.Context, modelItem *cache.ModelCacheItem, req ChatCompletionRequest) {
	// 准备转发到厂商的请求
	providerURL := modelItem.ProviderBaseURL + "/chat/completions"
	providerKey := modelItem.ProviderKey

	// 更新 model 字段为厂商实际的模型ID
	req.Model = modelItem.Model.ModelID

	// 直接序列化 req（包含所有已修改的消息）
	providerReqBody, _ := req.MarshalJSON()

	// 发送请求到厂商
	client := globalHTTPClient
	providerReq, err := http.NewRequest("POST", providerURL, bytes.NewReader(providerReqBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建请求失败",
		})
		return
	}

	// 设置请求头
	providerReq.Header.Set("Content-Type", "application/json")
	providerReq.Header.Set("Authorization", "Bearer "+providerKey)

	// 发送请求
	resp, err := client.Do(providerReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, Response{
			Code:    502,
			Message: "请求厂商失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadGateway, Response{
			Code:    502,
			Message: "厂商返回错误: " + string(respBody),
		})
		return
	}

	// 如果不流式，直接返回响应
	if !req.Stream {
		respBody, _ := io.ReadAll(resp.Body)
		c.Response().Header().Set("Content-Type", "application/json")
		c.String(http.StatusOK, string(respBody))
		return
	}

	// 流式响应
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Transfer-Encoding", "chunked")

	// 先发送 HTTP 状态码 200 给客户端
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	// 使用 bufio.Reader 实时读取并转发每个数据块
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 正常结束
				break
			}
			// 连接错误
			return
		}

		// 转发数据块到客户端
		if _, writeErr := c.Response().Writer.Write([]byte(line)); writeErr != nil {
			// 客户端断开连接
			return
		}
		c.Response().Flush() // 强制立即发送
	}
}
