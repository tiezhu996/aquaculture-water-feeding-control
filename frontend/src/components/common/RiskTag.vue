<script setup lang="ts">
import { computed } from 'vue'
import type { RiskLevel } from '@/types/enums'
import { riskLabels } from '@/types/enums'

const props = defineProps<{ level?: RiskLevel }>()
const level = computed(() => props.level || 'normal')
const type = computed(() => level.value === 'critical' ? 'danger' : level.value === 'warning' ? 'warning' : 'success')
</script>

<template>
  <el-tag :type="type" effect="dark" size="small"><span class="risk-dot" />{{ riskLabels[level] }}</el-tag>
</template>
