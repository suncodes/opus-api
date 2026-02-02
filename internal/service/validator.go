package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"opus-api/internal/converter"
	"opus-api/internal/model"
	"opus-api/internal/types"
	"time"

	"gorm.io/gorm"
)

// CookieValidator Cookie 验证器
type CookieValidator struct {
	service *CookieService
}

// NewCookieValidator 创建验证器
func NewCookieValidator(service *CookieService) *CookieValidator {
	return &CookieValidator{service: service}
}

// ValidateCookie 验证单个 Cookie
func (v *CookieValidator) ValidateCookie(cookie *model.MorphCookie) bool {
	result := v.testCookie(cookie)

	db := v.service.GetDB()
	if result {
		db.Model(cookie).Updates(map[string]interface{}{
			"is_valid":       true,
			"last_validated": time.Now(),
			"error_count":    0,
		})
	} else {
		db.Model(cookie).Updates(map[string]interface{}{
			"is_valid":       false,
			"last_validated": time.Now(),
			"error_count":    gorm.Expr("error_count + ?", 1),
		})
	}

	return result
}

// ValidateAllCookies 验证用户的所有 Cookie
func (v *CookieValidator) ValidateAllCookies(userID uint) map[uint]bool {
	cookies, err := v.service.ListCookies(userID)
	if err != nil {
		return nil
	}

	results := make(map[uint]bool)
	for _, cookie := range cookies {
		results[cookie.ID] = v.ValidateCookie(&cookie)
	}

	return results
}

// testCookie 测试 Cookie 是否有效（与 /v1/messages 逻辑完全一致）
func (v *CookieValidator) testCookie(cookie *model.MorphCookie) bool {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建 Claude 格式的测试请求
	claudeReq := types.ClaudeRequest{
		Model:     types.DefaultModel,
		MaxTokens: 1024,
		Messages: []types.ClaudeMessage{
			{
				Role:    "user",
				Content: "Hello!",
			},
		},
	}

	// 使用与 /v1/messages 相同的转换逻辑，将 Claude 格式转换为 Morph 格式
	morphReq := converter.ClaudeToMorph(claudeReq)

	reqBody, _ := json.Marshal(morphReq)
	req, err := http.NewRequest("POST", types.MorphAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return false
	}

	// 使用与 /v1/messages 相同的请求头
	for key, value := range types.MorphHeaders {
		req.Header.Set(key, value)
	}
	// 覆盖 Cookie
	req.Header.Set("cookie", cookie.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查响应状态码
	// 200 OK 表示成功，401/403 表示认证失败
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false
	}

	// 尝试读取响应体
	body, _ := io.ReadAll(resp.Body)

	// 检查是否是有效的 API 响应（SSE 流格式）
	bodyStr := string(body)
	
	// Morph API 返回 SSE 流，有效的响应会包含 "data:" 前缀
	if !containsPrefix(bodyStr, "data:") {
		return false
	}

	return true
}

// containsPrefix 检查字符串是否包含指定前缀（忽略空白字符）
func containsPrefix(s, prefix string) bool {
	// 去除前导空白字符
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	if start >= len(s) {
		return false
	}
	
	return len(s[start:]) >= len(prefix) && s[start:start+len(prefix)] == prefix
}

// validateCookieQuiet 静默验证 Cookie（不更新数据库）
func (v *CookieValidator) validateCookieQuiet(cookie *model.MorphCookie) bool {
	return v.testCookie(cookie)
}