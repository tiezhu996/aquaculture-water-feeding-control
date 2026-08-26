package model

import "aquaculture-water-feeding-control/backend/internal/constants"

type Pond struct {
	Base
	Code             string               `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name             string               `gorm:"size:100;not null" json:"name"`
	Species          string               `gorm:"size:80;not null" json:"species"`
	AreaSquareMeters float64              `gorm:"not null" json:"areaSquareMeters"`
	CapacityKg       float64              `gorm:"not null" json:"capacityKg"`
	GrowthStage      string               `gorm:"size:40;not null" json:"growthStage"`
	Status           constants.PondStatus `gorm:"size:20;not null;index" json:"status"`
	Manager          string               `gorm:"size:80" json:"manager"`
	Notes            string               `gorm:"type:text" json:"notes"`
}
