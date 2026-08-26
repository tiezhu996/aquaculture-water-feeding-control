package constants

type PondStatus string

const (
	PondStatusActive     PondStatus = "active"
	PondStatusQuarantine PondStatus = "quarantine"
	PondStatusClosed     PondStatus = "closed"
)

func (s ExecutionStatus) Valid() bool {
	return s == ExecutionScheduled || s == ExecutionRunning || s == ExecutionCompleted || s == ExecutionCancelled
}

func (s ExecutionStatus) CanTransitionTo(next ExecutionStatus) bool {
	switch s {
	case ExecutionScheduled:
		return next == ExecutionScheduled || next == ExecutionCompleted || next == ExecutionCancelled
	case ExecutionRunning:
		return next == ExecutionScheduled || next == ExecutionRunning || next == ExecutionCompleted
	default:
		return false
	}
}

func (s PondStatus) Valid() bool {
	return s == PondStatusActive || s == PondStatusQuarantine
}

type PlanStatus string

const (
	PlanStatusDraft    PlanStatus = "draft"
	PlanStatusPending  PlanStatus = "pending"
	PlanStatusApproved PlanStatus = "approved"
	PlanStatusExecuted PlanStatus = "executed"
)

func (s PlanStatus) Valid() bool {
	return s == PlanStatusDraft || s == PlanStatusPending || s == PlanStatusApproved || s == PlanStatusExecuted
}

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleManager  Role = "manager"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleManager || r == RoleOperator || r == RoleViewer
}

type RiskLevel string

const (
	RiskNormal   RiskLevel = "normal"
	RiskWarning  RiskLevel = "warning"
	RiskCritical RiskLevel = "critical"
)

type ExecutionStatus string

const (
	ExecutionScheduled ExecutionStatus = "scheduled"
	ExecutionRunning   ExecutionStatus = "running"
	ExecutionCompleted ExecutionStatus = "completed"
	ExecutionCancelled ExecutionStatus = "cancelled"
)
