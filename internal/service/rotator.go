package service

import (
	"errors"
	"opus-api/internal/model"
	"sync"
	"time"

	"gorm.io/gorm"
)

// RotationStrategy Cookie 轮询策略
type RotationStrategy string

const (
	StrategyRoundRobin RotationStrategy = "round_robin" // 轮询
	StrategyPriority   RotationStrategy = "priority"    // 优先级
	StrategyLeastUsed  RotationStrategy = "least_used"  // 最少使用
)

var (
	ErrNoCookiesAvailable = errors.New("no valid cookies available")
)

// CookieRotator Cookie 轮询器
type CookieRotator struct {
	service  *CookieService
	strategy RotationStrategy
	mu       sync.RWMutex
	index    int // 用于 round_robin 策略
}

// NewCookieRotator 创建轮询器
func NewCookieRotator(service *CookieService, strategy RotationStrategy) *CookieRotator {
	if strategy == "" {
		strategy = StrategyRoundRobin
	}
	return &CookieRotator{
		service:  service,
		strategy: strategy,
		index:    0,
	}
}

// NextCookie 获取下一个可用的 Cookie
func (r *CookieRotator) NextCookie() (interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cookies, err := r.service.GetAllValidCookies()
	if err != nil {
		return nil, err
	}

	if len(cookies) == 0 {
		return nil, ErrNoCookiesAvailable
	}

	var selected *model.MorphCookie

	switch r.strategy {
	case StrategyRoundRobin:
		selected = r.roundRobin(cookies)
	case StrategyPriority:
		selected = r.priority(cookies)
	case StrategyLeastUsed:
		selected = r.leastUsed(cookies)
	default:
		selected = r.roundRobin(cookies)
	}

	if selected == nil {
		return nil, ErrNoCookiesAvailable
	}

	return selected, nil
}

// roundRobin 轮询策略
func (r *CookieRotator) roundRobin(cookies []model.MorphCookie) *model.MorphCookie {
	if len(cookies) == 0 {
		return nil
	}

	r.index = r.index % len(cookies)
	selected := &cookies[r.index]
	r.index++

	return selected
}

// priority 优先级策略（优先级高的优先使用）
func (r *CookieRotator) priority(cookies []model.MorphCookie) *model.MorphCookie {
	if len(cookies) == 0 {
		return nil
	}

	// cookies 已经按照 priority DESC 排序
	return &cookies[0]
}

// leastUsed 最少使用策略
func (r *CookieRotator) leastUsed(cookies []model.MorphCookie) *model.MorphCookie {
	if len(cookies) == 0 {
		return nil
	}

	// cookies 已经按照 usage_count ASC 排序
	return &cookies[0]
}

// MarkUsed 标记 Cookie 已使用
func (r *CookieRotator) MarkUsed(cookieID uint) error {
	db := r.service.GetDB()
	return db.Model(&model.MorphCookie{}).
		Where("id = ?", cookieID).
		Updates(map[string]interface{}{
			"usage_count": gorm.Expr("usage_count + ?", 1),
			"last_used":   time.Now(),
		}).Error
}

// MarkInvalid 标记 Cookie 无效
func (r *CookieRotator) MarkInvalid(cookieID uint) error {
	db := r.service.GetDB()
	return db.Model(&model.MorphCookie{}).
		Where("id = ?", cookieID).
		Updates(map[string]interface{}{
			"is_valid": false,
		}).Error
}

// MarkError 标记 Cookie 错误
func (r *CookieRotator) MarkError(cookieID uint) error {
	db := r.service.GetDB()

	var cookie model.MorphCookie
	if err := db.First(&cookie, cookieID).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"error_count": gorm.Expr("error_count + ?", 1),
	}

	// 如果错误次数超过阈值，标记为无效
	if cookie.ErrorCount >= 5 {
		updates["is_valid"] = false
	}

	return db.Model(&model.MorphCookie{}).
		Where("id = ?", cookieID).
		Updates(updates).Error
}

// GetStrategy 获取当前策略
func (r *CookieRotator) GetStrategy() RotationStrategy {
	return r.strategy
}

// SetStrategy 设置轮询策略
func (r *CookieRotator) SetStrategy(strategy RotationStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategy = strategy
	r.index = 0 // 重置索引
}