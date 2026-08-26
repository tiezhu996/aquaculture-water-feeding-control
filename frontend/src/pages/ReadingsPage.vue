<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Bell, CircleCheck, DataAnalysis, Plus, Warning } from '@element-plus/icons-vue'
import { readingApi } from '@/api/readings'
import { pondApi } from '@/api/ponds'
import MetricCard from '@/components/common/MetricCard.vue'
import RiskTag from '@/components/common/RiskTag.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAuth } from '@/hooks/useAuth'
import { useQueryParams } from '@/hooks/useQueryParams'
import type { Pond, WaterReading, WaterReadingInput } from '@/types/models'
import { errorMessage } from '@/utils/errors'
import { formatDateTime, toISO, toLocalInput } from '@/utils/format'

const { canOperate, canReview } = useAuth()
const { params } = useQueryParams({ status: '', pondId: '', page: 1 })
const readings = ref<WaterReading[]>([])
const ponds = ref<Pond[]>([])
const total = ref(0)
const loading = ref(false)
const saving = ref(false)
const editorOpen = ref(false)
const confirmOpen = ref(false)
const deleteOpen = ref(false)
const target = ref<WaterReading | null>(null)
const confirmationNote = ref('')
const measuredAtLocal = ref(toLocalInput())
const form = reactive<WaterReadingInput>({ pondId: 0, dissolvedOxygen: 6, temperature: 26, ph: 7.5, ammonia: 0.1, turbidity: 25, measuredAt: '', source: 'manual' })

const warningCount = computed(() => readings.value.filter((item) => item.riskLevel === 'warning').length)
const criticalCount = computed(() => readings.value.filter((item) => item.riskLevel === 'critical').length)
const unconfirmedCount = computed(() => readings.value.filter((item) => item.riskLevel !== 'normal' && !item.confirmed).length)

async function load() {
  loading.value = true
  try {
    const [result, pondResult] = await Promise.all([
      readingApi.list({ page: Number(params.page), pageSize: 20, status: String(params.status), pondId: Number(params.pondId) || undefined }),
      pondApi.list({ page: 1, pageSize: 100 }),
    ])
    readings.value = result.items
    total.value = result.total
    ponds.value = pondResult.items
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, { pondId: ponds.value[0]?.id || 0, dissolvedOxygen: 6, temperature: 26, ph: 7.5, ammonia: 0.1, turbidity: 25, source: 'manual' })
  measuredAtLocal.value = toLocalInput()
  editorOpen.value = true
}

async function create() {
  if (!form.pondId) {
    ElMessage.warning('请选择养殖池')
    return
  }
  saving.value = true
  try {
    await readingApi.create({ ...form, measuredAt: toISO(measuredAtLocal.value) })
    ElMessage.success('水质读数已录入')
    editorOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

function openConfirm(reading: WaterReading) {
  target.value = reading
  confirmationNote.value = ''
  confirmOpen.value = true
}

async function confirmReading() {
  if (!target.value || confirmationNote.value.trim().length < 2) {
    ElMessage.warning('请填写处置或确认说明')
    return
  }
  saving.value = true
  try {
    await readingApi.confirm(target.value.id, confirmationNote.value)
    ElMessage.success('异常读数已确认')
    confirmOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!target.value) return
  saving.value = true
  try {
    await readingApi.remove(target.value.id)
    ElMessage.success('手工读数已删除')
    deleteOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

let timer: number | undefined
watch(params, () => { window.clearTimeout(timer); timer = window.setTimeout(load, 200) }, { deep: true })
onMounted(load)
</script>

<template>
  <div class="page-stack">
    <section class="metrics-grid">
      <MetricCard label="页内读数" :value="readings.length" :icon="DataAnalysis" hint="按测量时间倒序" />
      <MetricCard label="正常" :value="readings.length - warningCount - criticalCount" :icon="CircleCheck" tone="green" />
      <MetricCard label="预警 / 严重" :value="`${warningCount} / ${criticalCount}`" :icon="Warning" tone="amber" />
      <MetricCard label="待确认异常" :value="unconfirmedCount" :icon="Bell" tone="red" hint="需人工复核" />
    </section>
    <section class="workspace-panel">
      <div class="panel-toolbar">
        <div class="filters">
          <el-select v-model="params.pondId" placeholder="全部养殖池" clearable><el-option v-for="pond in ponds" :key="pond.id" :label="pond.name" :value="String(pond.id)" /></el-select>
          <el-select v-model="params.status" placeholder="全部风险" clearable><el-option label="正常" value="normal" /><el-option label="预警" value="warning" /><el-option label="严重" value="critical" /></el-select>
        </div>
        <el-button v-if="canOperate()" type="primary" :icon="Plus" @click="openCreate">录入读数</el-button>
      </div>
      <el-table v-loading="loading" :data="readings" stripe empty-text="暂无水质读数">
        <el-table-column label="养殖池 / 时间" min-width="190"><template #default="{ row }"><div class="primary-cell"><strong>{{ row.pond?.name || `#${row.pondId}` }}</strong><small>{{ formatDateTime(row.measuredAt) }}</small></div></template></el-table-column>
        <el-table-column label="溶解氧" width="105"><template #default="{ row }">{{ row.dissolvedOxygen }} mg/L</template></el-table-column>
        <el-table-column label="水温" width="85"><template #default="{ row }">{{ row.temperature }}℃</template></el-table-column>
        <el-table-column label="pH" prop="ph" width="70" />
        <el-table-column label="氨氮" width="85"><template #default="{ row }">{{ row.ammonia }}</template></el-table-column>
        <el-table-column label="风险" width="90"><template #default="{ row }"><RiskTag :level="row.riskLevel" /></template></el-table-column>
        <el-table-column label="判定说明" prop="alertMessage" min-width="220" show-overflow-tooltip />
        <el-table-column label="复核" width="90"><template #default="{ row }"><span v-if="row.confirmed" class="confirmed-text">已确认</span><span v-else class="muted">—</span></template></el-table-column>
        <el-table-column v-if="canOperate()" label="操作" width="140" fixed="right"><template #default="{ row }"><el-button v-if="row.riskLevel !== 'normal' && !row.confirmed" link type="primary" @click="openConfirm(row)">确认</el-button><el-button v-if="canReview() && row.source === 'manual' && !row.confirmed" link type="danger" @click="target = row; deleteOpen = true">删除</el-button></template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="params.page" layout="total, prev, pager, next" :total="total" :page-size="20" /></div>
    </section>
    <el-dialog v-model="editorOpen" title="录入水质读数" width="680px">
      <el-alert title="保存后系统将自动评估风险等级" type="info" :closable="false" show-icon />
      <el-form label-position="top" class="form-grid form-with-alert">
        <el-form-item label="养殖池"><el-select v-model="form.pondId"><el-option v-for="pond in ponds" :key="pond.id" :label="`${pond.name} (${pond.code})`" :value="pond.id" /></el-select></el-form-item>
        <el-form-item label="测量时间"><el-date-picker v-model="measuredAtLocal" type="datetime" value-format="YYYY-MM-DDTHH:mm" /></el-form-item>
        <el-form-item label="溶解氧（mg/L）"><el-input-number v-model="form.dissolvedOxygen" :min="0" :max="30" :step="0.1" /></el-form-item>
        <el-form-item label="水温（℃）"><el-input-number v-model="form.temperature" :min="-5" :max="50" :step="0.1" /></el-form-item>
        <el-form-item label="pH"><el-input-number v-model="form.ph" :min="0" :max="14" :step="0.1" /></el-form-item>
        <el-form-item label="氨氮（mg/L）"><el-input-number v-model="form.ammonia" :min="0" :max="20" :step="0.01" /></el-form-item>
        <el-form-item label="浊度（NTU）"><el-input-number v-model="form.turbidity" :min="0" :max="1000" :step="1" /></el-form-item>
        <el-form-item label="来源"><el-select v-model="form.source"><el-option label="手工录入" value="manual" /><el-option label="传感器" value="sensor" /><el-option label="数据导入" value="import" /></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="editorOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="create">保存并评估</el-button></template>
    </el-dialog>
    <el-dialog v-model="confirmOpen" title="确认水质异常" width="520px">
      <div v-if="target" class="risk-summary"><RiskTag :level="target.riskLevel" /><p>{{ target.alertMessage }}</p></div>
      <el-form-item label="处置 / 复核说明"><el-input v-model="confirmationNote" type="textarea" :rows="4" placeholder="记录现场复核情况和已采取的措施" /></el-form-item>
      <template #footer><el-button @click="confirmOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="confirmReading">确认并留痕</el-button></template>
    </el-dialog>
    <ConfirmDialog v-model="deleteOpen" title="删除手工读数" message="删除后仍会保留操作审计，确认继续？" danger :loading="saving" @confirm="remove" />
  </div>
</template>
