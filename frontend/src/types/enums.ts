export type PondStatus = 'active' | 'quarantine' | 'closed'
export type PlanStatus = 'draft' | 'pending' | 'approved' | 'executed'
export type RiskLevel = 'normal' | 'warning' | 'critical'
export type ExecutionStatus = 'scheduled' | 'running' | 'completed' | 'cancelled'
export type UserRole = 'admin' | 'manager' | 'operator' | 'viewer'

export const pondStatusLabels: Record<PondStatus, string> = {
  active: '运行中',
  quarantine: '隔离观察',
  closed: '已关闭',
}

export const planStatusLabels: Record<PlanStatus, string> = {
  draft: '草稿',
  pending: '待审核',
  approved: '已批准',
  executed: '已执行',
}

export const riskLabels: Record<RiskLevel, string> = {
  normal: '正常',
  warning: '预警',
  critical: '严重',
}

export const executionStatusLabels: Record<ExecutionStatus, string> = {
  scheduled: '待执行',
  running: '执行中',
  completed: '已完成',
  cancelled: '已取消',
}
