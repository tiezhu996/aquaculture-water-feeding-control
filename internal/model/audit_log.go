package model

type AuditLog struct {
	Base
	Actor      string `gorm:"size:80;not null;index" json:"actor"`
	Role       string `gorm:"size:20;not null" json:"role"`
	Action     string `gorm:"size:60;not null;index" json:"action"`
	EntityType string `gorm:"size:40;not null;index" json:"entityType"`
	EntityID   uint   `gorm:"not null;index" json:"entityId"`
	FromState  string `gorm:"type:text" json:"fromState"`
	ToState    string `gorm:"type:text" json:"toState"`
	Reason     string `gorm:"type:text" json:"reason"`
	RequestID  string `gorm:"size:64;index" json:"requestId"`
}
