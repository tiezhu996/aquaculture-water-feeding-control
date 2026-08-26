package dto

import "time"

type FeedingPlanInput struct {
	PondID            uint      `json:"pondId" binding:"required"`
	Name              string    `json:"name" binding:"required,min=2,max=120"`
	DailyAmountKg     float64   `json:"dailyAmountKg" binding:"required,gt=0"`
	FrequencyPerDay   int       `json:"frequencyPerDay" binding:"required,gte=1,lte=12"`
	FeedType          string    `json:"feedType" binding:"required,max=80"`
	TargetGrowthStage string    `json:"targetGrowthStage" binding:"required,max=40"`
	MinOxygen         float64   `json:"minOxygen" binding:"gte=0,lte=20"`
	StartDate         time.Time `json:"startDate" binding:"required"`
	EndDate           time.Time `json:"endDate" binding:"required"`
	Rationale         string    `json:"rationale" binding:"required,min=5,max=2000"`
}

type TransitionPlanInput struct {
	Reason string `json:"reason" binding:"required,min=2,max=500"`
}

type FeedingRecommendation struct {
	PondID             uint      `json:"pondId"`
	PlanID             uint      `json:"planId"`
	PlanVersion        int       `json:"planVersion"`
	GeneratedAt        time.Time `json:"generatedAt"`
	ReadingMeasuredAt  time.Time `json:"readingMeasuredAt"`
	Weather            string    `json:"weather"`
	Action             string    `json:"action"`
	DailyAmountKg      float64   `json:"dailyAmountKg"`
	AmountPerFeedingKg float64   `json:"amountPerFeedingKg"`
	FrequencyPerDay    int       `json:"frequencyPerDay"`
	AdjustmentPercent  float64   `json:"adjustmentPercent"`
	Reasons            []string  `json:"reasons"`
}
