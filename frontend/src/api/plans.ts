import { client, type ApiEnvelope } from './client'
import type { FeedingPlan, FeedingPlanInput, FeedingRecommendation, PageQuery, PageResult } from '@/types/models'

export const planApi = {
  async list(params: PageQuery = {}) {
    const response = await client.get<ApiEnvelope<PageResult<FeedingPlan>>>('/plans', { params })
    return response.data.data
  },
  async create(input: FeedingPlanInput) {
    const response = await client.post<ApiEnvelope<FeedingPlan>>('/plans', input)
    return response.data.data
  },
  async update(id: number, input: FeedingPlanInput) {
    const response = await client.put<ApiEnvelope<FeedingPlan>>(`/plans/${id}`, input)
    return response.data.data
  },
  async transition(id: number, action: 'submit' | 'approve' | 'revoke', reason: string) {
    const response = await client.patch<ApiEnvelope<FeedingPlan>>(`/plans/${id}/${action}`, { reason })
    return response.data.data
  },
  async recommendation(pondId: number, weather: string) {
    const response = await client.get<ApiEnvelope<FeedingRecommendation>>('/plans/recommendation', { params: { pondId, weather } })
    return response.data.data
  },
  async remove(id: number) {
    await client.delete(`/plans/${id}`)
  },
}
