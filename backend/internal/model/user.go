package model

import "time"

type User struct {
	ID                int64      `gorm:"column:id;primaryKey" json:"id"`
	Username          string     `gorm:"column:username" json:"username"`
	Email             *string    `gorm:"column:email" json:"email"`
	PasswordHash      string     `gorm:"column:password_hash" json:"-"`
	Nickname          *string    `gorm:"column:nickname" json:"nickname"`
	Role              string     `gorm:"column:role" json:"role"`
	Status            string     `gorm:"column:status" json:"status"`
	AvatarPath        *string    `gorm:"column:avatar_path" json:"-"`
	AvatarURL         *string    `gorm:"-" json:"avatar_url"`
	StorageQuotaBytes *int64     `gorm:"column:storage_quota_bytes" json:"storage_quota_bytes"`
	StorageUsedBytes  int64      `gorm:"-" json:"storage_used_bytes"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"-"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
}

func (User) TableName() string { return "users" }

type RefreshToken struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	UserAgent *string    `gorm:"column:user_agent"`
	IPAddress *string    `gorm:"column:ip_address"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type AccessClaims struct {
	Subject string `json:"sub"`
	Role    string `json:"role"`
	Type    string `json:"typ"`
	Expires int64  `json:"exp"`
}
