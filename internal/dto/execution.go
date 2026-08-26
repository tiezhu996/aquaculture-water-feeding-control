package dto

import (
	"aquaculture-water-feeding-control/backend/internal/constants"
	"time"
)

type ExecutionInput struct {
	PondID          uint      `json:"pondId" binding:"required"`
	FeedingPlanID   uint      `json:"feedingPlanId" binding:"required"`
	ScheduledAt     time.Time `json:"scheduledAt" binding:"required"`
	PlannedAmountKg float64   `json:"plannedAmountKg" binding:"required,gt=0"`
	Weather         string    `json:"weather" binding:"max=120"`
}

type UpdateExecutionInput struct {
	ScheduledAt     time.Time                 `json:"scheduledAt" binding:"required"`
	PlannedAmountKg float64                   `json:"plannedAmountKg" binding:"required,gt=0"`
	Weather         string                    `json:"weather" binding:"max=120"`
	Status          constants.ExecutionStatus `json:"status" binding:"required,oneof=scheduled running cancelled"`
}

type CompleteExecutionInput struct {
	ActualAmountKg float64 `json:"actualAmountKg" binding:"required,gt=0"`
	OxygenSnapshot float64 `json:"oxygenSnapshot" binding:"gte=0,lte=30"`
	Feedback       string  `json:"feedback" binding:"required,min=2,max=1000"`
}
