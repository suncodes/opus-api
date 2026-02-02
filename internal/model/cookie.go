package model

import (
	"time"
)

// MorphCookie Morph Cookie 模型
type MorphCookie struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	User          User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Name          string     `gorm:"size:100;not null" json:"name"`
	APIKey        string     `gorm:"column:api_key;type:text;not null" json:"api_key"`
	SessionKey    string     `gorm:"column:session_key;type:text" json:"session_key"`
	IsValid       bool       `gorm:"default:true;index" json:"is_valid"`
	LastValidated *time.Time `gorm:"column:last_validated" json:"last_validated"`
	LastUsed      *time.Time `gorm:"column:last_used" json:"last_used"`
	Priority      int        `gorm:"default:0;index" json:"priority"`
	UsageCount    int64      `gorm:"default:0" json:"usage_count"`
	ErrorCount    int        `gorm:"default:0" json:"error_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CookieStats Cookie 统计信息
type CookieStats struct {
	TotalCount   int64 `json:"total_count"`
	ValidCount   int64 `json:"valid_count"`
	InvalidCount int64 `json:"invalid_count"`
	TotalUsage   int64 `json:"total_usage"`
}

// TableName 指定表名
func (MorphCookie) TableName() string {
	return "morph_cookies"
}

// MarkUsed 标记 Cookie 已使用
func (c *MorphCookie) MarkUsed() {
	now := time.Now()
	c.LastUsed = &now
	c.UsageCount++
	c.ErrorCount = 0 // 成功使用后重置错误计数
}

// MarkError 标记 Cookie 错误
func (c *MorphCookie) MarkError() {
	c.ErrorCount++
}

// MarkInvalid 标记 Cookie 无效
func (c *MorphCookie) MarkInvalid() {
	c.IsValid = false
	now := time.Now()
	c.LastValidated = &now
}

// MarkValid 标记 Cookie 有效
func (c *MorphCookie) MarkValid() {
	c.IsValid = true
	c.ErrorCount = 0
	now := time.Now()
	c.LastValidated = &now
}