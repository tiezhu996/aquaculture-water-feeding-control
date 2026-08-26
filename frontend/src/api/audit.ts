import { client, type ApiEnvelope } from './client'
import type { AuditLog, PageQuery, PageResult } from '@/types/models'

export const auditApi = {
  async list(params: PageQuery = {}) {
    const response = await client.get<ApiEnvelope<PageResult<AuditLog>>>('/audit', { params })
    return response.data.data
  },
}
