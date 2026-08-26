import { client, type ApiEnvelope } from './client'
import type { ControlExecution, ExecutionInput, PageQuery, PageResult } from '@/types/models'
import type { ExecutionStatus } from '@/types/enums'

export const executionApi = {
  async list(params: PageQuery = {}) {
    const response = await client.get<ApiEnvelope<PageResult<ControlExecution>>>('/executions', { params })
    return response.data.data
  },
  async create(input: ExecutionInput) {
    const response = await client.post<ApiEnvelope<ControlExecution>>('/executions', input)
    return response.data.data
  },
  async update(id: number, input: { scheduledAt: string; plannedAmountKg: number; weather: string; status: ExecutionStatus }) {
    const response = await client.put<ApiEnvelope<ControlExecution>>(`/executions/${id}`, input)
    return response.data.data
  },
  async complete(id: number, input: { actualAmountKg: number; oxygenSnapshot: number; feedback: string }) {
    const response = await client.patch<ApiEnvelope<ControlExecution>>(`/executions/${id}/complete`, input)
    return response.data.data
  },
  async remove(id: number) {
    await client.delete(`/executions/${id}`)
  },
}
