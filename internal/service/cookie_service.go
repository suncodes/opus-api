package service

import (
	"errors"
	"opus-api/internal/model"

	"gorm.io/gorm"
)

var (
	ErrCookieNotFound = errors.New("cookie not found")
)

// CookieService Cookie 管理服务
type CookieService struct {
	db *gorm.DB
}

// NewCookieService 创建 Cookie 服务
func NewCookieService(db *gorm.DB) *CookieService {
	return &CookieService{db: db}
}

// GetDB 获取数据库连接
func (s *CookieService) GetDB() *gorm.DB {
	return s.db
}

// ListCookies 获取用户的所有 Cookie
func (s *CookieService) ListCookies(userID uint) ([]model.MorphCookie, error) {
	var cookies []model.MorphCookie
	err := s.db.Where("user_id = ?", userID).
		Order("priority DESC, created_at DESC").
		Find(&cookies).Error
	return cookies, err
}

// GetCookie 获取单个 Cookie
func (s *CookieService) GetCookie(id, userID uint) (*model.MorphCookie, error) {
	var cookie model.MorphCookie
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&cookie).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCookieNotFound
		}
		return nil, err
	}
	return &cookie, nil
}

// CreateCookie 创建 Cookie
func (s *CookieService) CreateCookie(cookie *model.MorphCookie) error {
	return s.db.Create(cookie).Error
}

// UpdateCookie 更新 Cookie
func (s *CookieService) UpdateCookie(cookie *model.MorphCookie) error {
	return s.db.Save(cookie).Error
}

// DeleteCookie 删除 Cookie
func (s *CookieService) DeleteCookie(id, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.MorphCookie{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCookieNotFound
	}
	return nil
}

// GetStats 获取统计信息
func (s *CookieService) GetStats(userID uint) (*model.CookieStats, error) {
	stats := &model.CookieStats{}

	// 总数
	if err := s.db.Model(&model.MorphCookie{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 有效数量
	if err := s.db.Model(&model.MorphCookie{}).
		Where("user_id = ? AND is_valid = ?", userID, true).
		Count(&stats.ValidCount).Error; err != nil {
		return nil, err
	}

	// 无效数量
	stats.InvalidCount = stats.TotalCount - stats.ValidCount

	// 总使用次数
	var totalUsage *int64
	if err := s.db.Model(&model.MorphCookie{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(usage_count), 0)").
		Scan(&totalUsage).Error; err != nil {
		return nil, err
	}
	if totalUsage != nil {
		stats.TotalUsage = *totalUsage
	}

	return stats, nil
}

// GetValidCookies 获取所有有效的 Cookie
func (s *CookieService) GetValidCookies(userID uint) ([]model.MorphCookie, error) {
	var cookies []model.MorphCookie
	err := s.db.Where("user_id = ? AND is_valid = ?", userID, true).
		Order("priority DESC, usage_count ASC").
		Find(&cookies).Error
	return cookies, err
}

// GetAllValidCookies 获取系统中所有有效的 Cookie（用于轮询）
func (s *CookieService) GetAllValidCookies() ([]model.MorphCookie, error) {
	var cookies []model.MorphCookie
	err := s.db.Where("is_valid = ?", true).
		Order("priority DESC, usage_count ASC").
		Find(&cookies).Error
	return cookies, err
}