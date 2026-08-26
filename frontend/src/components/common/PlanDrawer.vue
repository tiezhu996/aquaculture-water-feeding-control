<script setup lang="ts">
import StatusBadge from './StatusBadge.vue'
import type { FeedingPlan } from '@/types/models'
import { formatDateTime, formatNumber } from '@/utils/format'

defineProps<{ modelValue: boolean; plan: FeedingPlan | null }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>

<template>
  <el-drawer :model-value="modelValue" title="投喂计划详情" size="440px" @close="emit('update:modelValue', false)">
    <template v-if="plan">
      <div class="drawer-heading">
        <div><span class="eyebrow">v{{ plan.version }}</span><h3>{{ plan.name }}</h3></div>
        <StatusBadge :status="plan.status" />
      </div>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="养殖池">{{ plan.pond?.name || `#${plan.pondId}` }}</el-descriptions-item>
        <el-descriptions-item label="日投喂量">{{ formatNumber(plan.dailyAmountKg) }} kg</el-descriptions-item>
        <el-descriptions-item label="频次">{{ plan.frequencyPerDay }} 次/日</el-descriptions-item>
        <el-descriptions-item label="饲料">{{ plan.feedType }}</el-descriptions-item>
        <el-descriptions-item label="最低溶解氧">{{ plan.minOxygen }} mg/L</el-descriptions-item>
        <el-descriptions-item label="周期">{{ formatDateTime(plan.startDate) }} 至 {{ formatDateTime(plan.endDate) }}</el-descriptions-item>
        <el-descriptions-item label="创建人">{{ plan.createdBy }}</el-descriptions-item>
        <el-descriptions-item label="审核人">{{ plan.reviewedBy || '—' }}</el-descriptions-item>
      </el-descriptions>
      <section class="drawer-section"><h4>制定依据</h4><p>{{ plan.rationale }}</p></section>
    </template>
  </el-drawer>
</template>
