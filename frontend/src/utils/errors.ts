import axios from 'axios'
import type { ApiErrorBody } from '@/api/client'

export function errorMessage(error: unknown): string {
  if (axios.isAxiosError<ApiErrorBody>(error)) {
    return error.response?.data?.error?.message || error.message || '请求失败'
  }
  return error instanceof Error ? error.message : '未知错误'
}
