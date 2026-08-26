package dto

import "time"

type WaterReadingInput struct {
	PondID          uint      `json:"pondId" binding:"required"`
	DissolvedOxygen float64   `json:"dissolvedOxygen" binding:"gte=0,lte=30"`
	Temperature     float64   `json:"temperature" binding:"gte=-5,lte=50"`
	PH              float64   `json:"ph" binding:"gte=0,lte=14"`
	Ammonia         float64   `json:"ammonia" binding:"gte=0,lte=20"`
	Turbidity       float64   `json:"turbidity" binding:"gte=0,lte=1000"`
	MeasuredAt      time.Time `json:"measuredAt" binding:"required"`
	Source          string    `json:"source" binding:"required,oneof=sensor manual import"`
}

type ConfirmReadingInput struct {
	Note string `json:"note" binding:"required,min=2,max=500"`
}
