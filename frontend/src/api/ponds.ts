import { client, type ApiEnvelope } from './client'
import type { PageQuery, PageResult, Pond, PondInput } from '@/types/models'

export const pondApi = {
  async list(params: PageQuery = {}) {
    const response = await client.get<ApiEnvelope<PageResult<Pond>>>('/ponds', { params })
    return response.data.data
  },
  async get(id: number) {
    const response = await client.get<ApiEnvelope<Pond>>(`/ponds/${id}`)
    return response.data.data
  },
  async create(input: PondInput) {
    const response = await client.post<ApiEnvelope<Pond>>('/ponds', input)
    return response.data.data
  },
  async update(id: number, input: PondInput) {
    const response = await client.put<ApiEnvelope<Pond>>(`/ponds/${id}`, input)
    return response.data.data
  },
  async remove(id: number) {
    await client.delete(`/ponds/${id}`)
  },
}
