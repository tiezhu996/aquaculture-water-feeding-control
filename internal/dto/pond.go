package dto

import "aquaculture-water-feeding-control/backend/internal/constants"

type PondInput struct {
	Code             string               `json:"code" binding:"required,min=2,max=32"`
	Name             string               `json:"name" binding:"required,min=2,max=100"`
	Species          string               `json:"species" binding:"required,max=80"`
	AreaSquareMeters float64              `json:"areaSquareMeters" binding:"required,gt=0"`
	CapacityKg       float64              `json:"capacityKg" binding:"required,gt=0"`
	GrowthStage      string               `json:"growthStage" binding:"required,max=40"`
	Status           constants.PondStatus `json:"status" binding:"required"`
	Manager          string               `json:"manager" binding:"max=80"`
	Notes            string               `json:"notes" binding:"max=1000"`
}
