package model

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"time"
)

type FeedingPlan struct {
	Base
	PondID            uint                 `gorm:"not null;index" json:"pondId"`
	Pond              *Pond                `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"pond,omitempty"`
	Name              string               `gorm:"size:120;not null" json:"name"`
	Version           int                  `gorm:"not null;default:1" json:"version"`
	DailyAmountKg     float64              `gorm:"not null" json:"dailyAmountKg"`
	FrequencyPerDay   int                  `gorm:"not null" json:"frequencyPerDay"`
	FeedType          string               `gorm:"size:80;not null" json:"feedType"`
	TargetGrowthStage string               `gorm:"size:40;not null" json:"targetGrowthStage"`
	MinOxygen         float64              `gorm:"not null" json:"minOxygen"`
	StartDate         time.Time            `gorm:"not null" json:"startDate"`
	EndDate           time.Time            `gorm:"not null" json:"endDate"`
	Status            constants.PlanStatus `gorm:"size:20;not null;index" json:"status"`
	Rationale         string               `gorm:"type:text" json:"rationale"`
	CreatedBy         string               `gorm:"size:80;not null" json:"createdBy"`
	ReviewedBy        string               `gorm:"size:80" json:"reviewedBy"`
	ReviewedAt        *time.Time           `json:"reviewedAt"`
}
