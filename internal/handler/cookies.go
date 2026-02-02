package handler

import (
	"net/http"
	"opus-api/internal/middleware"
	"opus-api/internal/model"
	"opus-api/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CookieHandler Cookie 管理处理器
type CookieHandler struct {
	cookieService *service.CookieService
	validator     *service.CookieValidator
}

// NewCookieHandler 创建 Cookie 处理器
func NewCookieHandler(cookieService *service.CookieService, validator *service.CookieValidator) *CookieHandler {
	return &CookieHandler{
		cookieService: cookieService,
		validator:     validator,
	}
}

// CreateCookieRequest 创建 Cookie 请求
type CreateCookieRequest struct {
	Name       string `json:"name" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
	SessionKey string `json:"session_key"`
	Priority   int    `json:"priority"`
}

// UpdateCookieRequest 更新 Cookie 请求
type UpdateCookieRequest struct {
	Name       string `json:"name"`
	APIKey     string `json:"api_key"`
	SessionKey string `json:"session_key"`
	Priority   *int   `json:"priority"`
	IsValid    *bool  `json:"is_valid"`
}

// CookieResponse Cookie 响应
type CookieResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	APIKey        string `json:"api_key"`
	SessionKey    string `json:"session_key"`
	IsValid       bool   `json:"is_valid"`
	Priority      int    `json:"priority"`
	UsageCount    int64  `json:"usage_count"`
	ErrorCount    int    `json:"error_count"`
	LastUsed      string `json:"last_used,omitempty"`
	LastValidated string `json:"last_validated,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ListCookies 获取 Cookie 列表
func (h *CookieHandler) ListCookies(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cookies, err := h.cookieService.ListCookies(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cookies"})
		return
	}

	responses := make([]CookieResponse, len(cookies))
	for i, cookie := range cookies {
		responses[i] = toCookieResponse(&cookie)
	}

	c.JSON(http.StatusOK, responses)
}

// GetCookie 获取单个 Cookie
func (h *CookieHandler) GetCookie(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cookie, err := h.cookieService.GetCookie(uint(id), userID)
	if err != nil {
		if err == service.ErrCookieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cookie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cookie"})
		return
	}

	c.JSON(http.StatusOK, toCookieResponse(cookie))
}

// CreateCookie 创建 Cookie
func (h *CookieHandler) CreateCookie(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateCookieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cookie := &model.MorphCookie{
		UserID:     userID,
		Name:       req.Name,
		APIKey:     req.APIKey,
		SessionKey: req.SessionKey,
		Priority:   req.Priority,
		IsValid:    true, // 默认有效，可以通过验证接口验证
	}

	if err := h.cookieService.CreateCookie(cookie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cookie"})
		return
	}

	c.JSON(http.StatusCreated, toCookieResponse(cookie))
}

// UpdateCookie 更新 Cookie
func (h *CookieHandler) UpdateCookie(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cookie, err := h.cookieService.GetCookie(uint(id), userID)
	if err != nil {
		if err == service.ErrCookieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cookie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cookie"})
		return
	}

	var req UpdateCookieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		cookie.Name = req.Name
	}
	if req.APIKey != "" {
		cookie.APIKey = req.APIKey
	}
	if req.SessionKey != "" {
		cookie.SessionKey = req.SessionKey
	}
	if req.Priority != nil {
		cookie.Priority = *req.Priority
	}
	if req.IsValid != nil {
		cookie.IsValid = *req.IsValid
	}

	if err := h.cookieService.UpdateCookie(cookie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cookie"})
		return
	}

	c.JSON(http.StatusOK, toCookieResponse(cookie))
}

// DeleteCookie 删除 Cookie
func (h *CookieHandler) DeleteCookie(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.cookieService.DeleteCookie(uint(id), userID); err != nil {
		if err == service.ErrCookieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cookie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete cookie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cookie deleted successfully"})
}

// ValidateCookie 验证单个 Cookie
func (h *CookieHandler) ValidateCookie(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	cookie, err := h.cookieService.GetCookie(uint(id), userID)
	if err != nil {
		if err == service.ErrCookieNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cookie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cookie"})
		return
	}

	isValid := h.validator.ValidateCookie(cookie)

	c.JSON(http.StatusOK, gin.H{
		"id":       cookie.ID,
		"is_valid": isValid,
	})
}

// ValidateAllCookies 验证所有 Cookie
func (h *CookieHandler) ValidateAllCookies(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	results := h.validator.ValidateAllCookies(userID)

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// GetStats 获取统计信息
func (h *CookieHandler) GetStats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.cookieService.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// toCookieResponse 转换为响应格式
func toCookieResponse(cookie *model.MorphCookie) CookieResponse {
	resp := CookieResponse{
		ID:         cookie.ID,
		Name:       cookie.Name,
		APIKey:     maskAPIKey(cookie.APIKey),
		SessionKey: cookie.SessionKey,
		IsValid:    cookie.IsValid,
		Priority:   cookie.Priority,
		UsageCount: cookie.UsageCount,
		ErrorCount: cookie.ErrorCount,
		CreatedAt:  cookie.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  cookie.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if cookie.LastUsed != nil {
		resp.LastUsed = cookie.LastUsed.Format("2006-01-02 15:04:05")
	}
	if cookie.LastValidated != nil {
		resp.LastValidated = cookie.LastValidated.Format("2006-01-02 15:04:05")
	}

	return resp
}

// maskAPIKey 隐藏 API Key 中间部分
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}