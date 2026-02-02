package model

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserSession 用户会话模型
type UserSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TokenHash string    `gorm:"uniqueIndex;size:255;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SetPassword 设置用户密码（加密）
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}

// CreateDefaultAdmin 创建默认管理员用户
func CreateDefaultAdmin(db *gorm.DB) error {
	username := os.Getenv("DEFAULT_ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if password == "" {
		password = "changeme123"
	}

	// 检查用户是否已存在
	var existingUser User
	result := db.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		// 用户已存在
		log.Printf("Default admin user '%s' already exists", username)
		return nil
	}

	// 创建新用户
	user := &User{
		Username: username,
	}
	if err := user.SetPassword(password); err != nil {
		return err
	}

	if err := db.Create(user).Error; err != nil {
		return err
	}

	log.Printf("Default admin user '%s' created successfully", username)
	return nil
}