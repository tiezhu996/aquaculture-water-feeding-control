package model

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"time"
)

type WaterReading struct {
	Base
	PondID           uint                `gorm:"not null;index" json:"pondId"`
	Pond             *Pond               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"pond,omitempty"`
	DissolvedOxygen  float64             `gorm:"not null" json:"dissolvedOxygen"`
	Temperature      float64             `gorm:"not null" json:"temperature"`
	PH               float64             `gorm:"column:ph;not null" json:"ph"`
	Ammonia          float64             `gorm:"not null" json:"ammonia"`
	Turbidity        float64             `gorm:"not null" json:"turbidity"`
	MeasuredAt       time.Time           `gorm:"not null;index" json:"measuredAt"`
	Source           string              `gorm:"size:30;not null" json:"source"`
	RiskLevel        constants.RiskLevel `gorm:"size:20;not null;index" json:"riskLevel"`
	AlertMessage     string              `gorm:"type:text" json:"alertMessage"`
	Confirmed        bool                `gorm:"not null;default:false" json:"confirmed"`
	ConfirmedBy      string              `gorm:"size:80" json:"confirmedBy"`
	ConfirmedAt      *time.Time          `json:"confirmedAt"`
	ConfirmationNote string              `gorm:"type:text" json:"confirmationNote"`
}
