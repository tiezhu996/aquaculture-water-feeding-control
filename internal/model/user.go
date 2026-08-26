package model

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"time"
)

type Base struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	Base
	Username     string         `gorm:"size:64;uniqueIndex;not null" json:"username"`
	DisplayName  string         `gorm:"size:100;not null" json:"displayName"`
	PasswordHash string         `gorm:"size:100;not null" json:"-"`
	Role         constants.Role `gorm:"size:20;not null;index" json:"role"`
	Active       bool           `gorm:"not null;default:true" json:"active"`
}
