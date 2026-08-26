<script setup lang="ts">
import { computed } from 'vue'
import { executionStatusLabels, planStatusLabels, pondStatusLabels } from '@/types/enums'

const props = defineProps<{ status: string }>()
const labels: Record<string, string> = { ...pondStatusLabels, ...planStatusLabels, ...executionStatusLabels }
const kind = computed(() => ({
  active: 'success', approved: 'success', completed: 'success',
  quarantine: 'warning', pending: 'warning', running: 'warning',
  closed: 'info', draft: 'info', scheduled: 'info',
  executed: 'success', cancelled: 'danger',
}[props.status] || 'info'))
</script>

<template>
  <el-tag :type="kind" effect="light" round>{{ labels[status] || status }}</el-tag>
</template>
