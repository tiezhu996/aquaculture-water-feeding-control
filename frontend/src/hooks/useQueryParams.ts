import { reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useQueryParams(defaults: Record<string, string | number>) {
  const route = useRoute()
  const router = useRouter()
  const params = reactive<Record<string, string | number>>({ ...defaults })

  for (const key of Object.keys(defaults)) {
    const value = route.query[key]
    if (typeof value === 'string') {
      params[key] = typeof defaults[key] === 'number' ? Number(value) || defaults[key] : value
    }
  }

  let timer: number | undefined
  watch(params, () => {
    window.clearTimeout(timer)
    timer = window.setTimeout(() => {
      const query: Record<string, string> = {}
      for (const [key, value] of Object.entries(params)) {
        if (value !== defaults[key] && value !== '') query[key] = String(value)
      }
      void router.replace({ query })
    }, 150)
  }, { deep: true })

  function reset() {
    Object.assign(params, defaults)
  }
  return { params, reset }
}
