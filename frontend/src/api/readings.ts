import { client, type ApiEnvelope } from './client'
import type { PageQuery, PageResult, WaterReading, WaterReadingInput } from '@/types/models'

export const readingApi = {
  async list(params: PageQuery = {}) {
    const response = await client.get<ApiEnvelope<PageResult<WaterReading>>>('/readings', { params })
    return response.data.data
  },
  async create(input: WaterReadingInput) {
    const response = await client.post<ApiEnvelope<WaterReading>>('/readings', input)
    return response.data.data
  },
  async confirm(id: number, note: string) {
    const response = await client.patch<ApiEnvelope<WaterReading>>(`/readings/${id}/confirm`, { note })
    return response.data.data
  },
  async remove(id: number) {
    await client.delete(`/readings/${id}`)
  },
}
