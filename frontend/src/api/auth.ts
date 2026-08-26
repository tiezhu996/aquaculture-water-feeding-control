import { client, type ApiEnvelope } from './client'
import type { User } from '@/types/models'

export const authApi = {
  async login(username: string, password: string) {
    const response = await client.post<ApiEnvelope<{ token: string; user: User }>>('/auth/login', { username, password })
    return response.data.data
  },
  async me() {
    const response = await client.get<ApiEnvelope<User>>('/auth/me')
    return response.data.data
  },
}
