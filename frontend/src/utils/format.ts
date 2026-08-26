export function formatDateTime(value?: string): string {
  if (!value) return '—'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false,
  }).format(new Date(value))
}

export function toLocalInput(value = new Date()): string {
  const date = new Date(value)
  date.setMinutes(date.getMinutes() - date.getTimezoneOffset())
  return date.toISOString().slice(0, 16)
}

export function toISO(value: string): string {
  return new Date(value).toISOString()
}

export function formatNumber(value: number, digits = 1): string {
  return new Intl.NumberFormat('zh-CN', { maximumFractionDigits: digits }).format(value)
}
