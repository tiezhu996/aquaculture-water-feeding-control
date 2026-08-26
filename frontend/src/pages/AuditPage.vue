<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Collection, Document, Search, User } from '@element-plus/icons-vue'
import { auditApi } from '@/api/audit'
import MetricCard from '@/components/common/MetricCard.vue'
import { useQueryParams } from '@/hooks/useQueryParams'
import type { AuditLog } from '@/types/models'
import { errorMessage } from '@/utils/errors'
import { formatDateTime } from '@/utils/format'

const { params } = useQueryParams({ search: '', entityType: '', action: '', page: 1 })
const logs = ref<AuditLog[]>([])
const total = ref(0)
const loading = ref(false)
const detailOpen = ref(false)
const selected = ref<AuditLog | null>(null)

async function load() {
  loading.value = true
  try {
    const result = await auditApi.list({
      page: Number(params.page), pageSize: 20, search: String(params.search),
      entityType: String(params.entityType), action: String(params.action),
    })
    logs.value = result.items
    total.value = result.total
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    loading.value = false
  }
}

function entityLabel(value: string) {
  return ({ pond: '养殖池', water_reading: '水质读数', feeding_plan: '投喂计划', control_execution: '执行记录' } as Record<string, string>)[value] || value
}

function actionLabel(value: string) {
  return ({ create: '创建', update: '更新', delete: '删除', confirm: '确认异常', revise: '版本修订', submit: '提交', approve: '批准', revoke: '撤销', schedule: '安排执行', complete: '完成反馈', execute: '计划执行' } as Record<string, string>)[value] || value
}

function pretty(value: string) {
  if (!value) return '—'
  try { return JSON.stringify(JSON.parse(value), null, 2) } catch { return value }
}

let timer: number | undefined
watch(params, () => { window.clearTimeout(timer); timer = window.setTimeout(load, 200) }, { deep: true })
onMounted(load)
</script>

<template>
  <div class="page-stack">
    <section class="metrics-grid audit-metrics">
      <MetricCard label="审计记录" :value="total" :icon="Collection" hint="按时间倒序" />
      <MetricCard label="页内操作人" :value="new Set(logs.map((item) => item.actor)).size" :icon="User" tone="blue" />
      <MetricCard label="页内实体类型" :value="new Set(logs.map((item) => item.entityType)).size" :icon="Document" tone="green" />
    </section>
    <section class="workspace-panel">
      <div class="panel-toolbar">
        <div class="filters audit-filters">
          <el-input v-model="params.search" clearable placeholder="搜索操作人或原因" :prefix-icon="Search" />
          <el-select v-model="params.entityType" placeholder="全部实体" clearable><el-option label="养殖池" value="pond" /><el-option label="水质读数" value="water_reading" /><el-option label="投喂计划" value="feeding_plan" /><el-option label="执行记录" value="control_execution" /></el-select>
          <el-select v-model="params.action" placeholder="全部操作" clearable><el-option label="创建" value="create" /><el-option label="更新" value="update" /><el-option label="批准" value="approve" /><el-option label="撤销" value="revoke" /><el-option label="完成反馈" value="complete" /><el-option label="删除" value="delete" /></el-select>
        </div>
      </div>
      <el-table v-loading="loading" :data="logs" stripe empty-text="暂无审计记录">
        <el-table-column label="时间" min-width="165"><template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template></el-table-column>
        <el-table-column label="操作人" min-width="130"><template #default="{ row }"><div class="primary-cell"><strong>{{ row.actor }}</strong><small>{{ row.role }}</small></div></template></el-table-column>
        <el-table-column label="操作" width="110"><template #default="{ row }"><el-tag effect="plain">{{ actionLabel(row.action) }}</el-tag></template></el-table-column>
        <el-table-column label="实体" min-width="140"><template #default="{ row }">{{ entityLabel(row.entityType) }} #{{ row.entityId }}</template></el-table-column>
        <el-table-column label="原因 / 说明" prop="reason" min-width="260" show-overflow-tooltip />
        <el-table-column label="请求 ID" prop="requestId" min-width="190" show-overflow-tooltip />
        <el-table-column label="操作" width="80" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="selected = row; detailOpen = true">详情</el-button></template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="params.page" layout="total, prev, pager, next" :total="total" :page-size="20" /></div>
    </section>
    <el-drawer v-model="detailOpen" title="审计详情" size="600px">
      <template v-if="selected">
        <el-descriptions :column="1" border><el-descriptions-item label="操作人">{{ selected.actor }} ({{ selected.role }})</el-descriptions-item><el-descriptions-item label="操作">{{ actionLabel(selected.action) }}</el-descriptions-item><el-descriptions-item label="实体">{{ entityLabel(selected.entityType) }} #{{ selected.entityId }}</el-descriptions-item><el-descriptions-item label="原因">{{ selected.reason }}</el-descriptions-item><el-descriptions-item label="请求 ID">{{ selected.requestId }}</el-descriptions-item></el-descriptions>
        <div class="snapshot-grid"><section><h4>变更前</h4><pre>{{ pretty(selected.fromState) }}</pre></section><section><h4>变更后</h4><pre>{{ pretty(selected.toState) }}</pre></section></div>
      </template>
    </el-drawer>
  </div>
</template>
