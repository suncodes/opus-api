package types

import (
	"errors"
	"strings"
)

const (
	MorphAPIURL = "https://www.morphllm.com/api/warpgrep-chat"
	LogDir      = "./logs"
)

var MorphHeaders = map[string]string{
	"accept":             "*/*",
	"accept-language":    "zh-CN,zh;q=0.9",
	"cache-control":      "no-cache",
	"content-type":       "application/json",
	"origin":             "https://www.morphllm.com",
	"pragma":             "no-cache",
	"priority":           "u=1, i",
	"referer":            "https://www.morphllm.com/playground/na/warpgrep?repo=tiangolo%2Ffastapi",
	"sec-ch-ua":          `"Not(A:Brand";v="8", "Chromium";v="144", "Google Chrome";v="144"`,
	"sec-ch-ua-mobile":   "?0",
	"sec-ch-ua-platform": `"macOS"`,
	"sec-fetch-dest":     "empty",
	"sec-fetch-mode":     "cors",
	"sec-fetch-site":     "same-origin",
	"user-agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
}

var DebugMode = true

// CookieRotatorInstance is a global reference to the cookie rotator service
// It's set in main.go after initialization
var CookieRotatorInstance interface {
	NextCookie() (cookie interface{}, err error)
	MarkUsed(cookieID uint) error
	MarkError(cookieID uint) error
}

// GetNextCookieFromRotator 从轮询器获取下一个 Cookie
// 这是一个辅助函数，用于从全局轮询器实例获取 Cookie
func GetNextCookieFromRotator() (interface{}, error) {
	if CookieRotatorInstance == nil {
		return nil, errors.New("cookie rotator not initialized")
	}
	return CookieRotatorInstance.NextCookie()
}

type ParsedToolCall struct {
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ========== 模型配置 ==========

// DefaultModel 默认使用的模型
const DefaultModel = "claude-opus-4-5-20251101"

// SupportedModels 支持的模型列表
var SupportedModels = []string{
	"claude-opus-4-5-20251101",
}

// IsModelSupported 检查模型是否在支持列表中
func IsModelSupported(model string) bool {
	if model == "" {
		return false
	}
	for _, supported := range SupportedModels {
		if strings.EqualFold(supported, model) {
			return true
		}
	}
	return false
}

// ========== 认证请求/响应类型 ==========

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
