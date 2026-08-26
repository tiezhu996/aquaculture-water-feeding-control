<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleCheck, Clock, DocumentChecked, Plus, Search, TrendCharts } from '@element-plus/icons-vue'
import { planApi } from '@/api/plans'
import { pondApi } from '@/api/ponds'
import MetricCard from '@/components/common/MetricCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import PlanDrawer from '@/components/common/PlanDrawer.vue'
import { useAuth } from '@/hooks/useAuth'
import { useQueryParams } from '@/hooks/useQueryParams'
import type { FeedingPlan, FeedingPlanInput, FeedingRecommendation, Pond } from '@/types/models'
import { errorMessage } from '@/utils/errors'
import { formatDateTime, formatNumber, toISO, toLocalInput } from '@/utils/format'

const { canOperate, canReview } = useAuth()
const { params } = useQueryParams({ search: '', status: '', pondId: '', page: 1 })
const plans = ref<FeedingPlan[]>([])
const ponds = ref<Pond[]>([])
const total = ref(0)
const loading = ref(false)
const saving = ref(false)
const editorOpen = ref(false)
const drawerOpen = ref(false)
const transitionOpen = ref(false)
const recommendationOpen = ref(false)
const editingId = ref<number | null>(null)
const selected = ref<FeedingPlan | null>(null)
const transitionAction = ref<'submit' | 'approve' | 'revoke'>('submit')
const transitionReason = ref('')
const startLocal = ref(toLocalInput())
const endLocal = ref(toLocalInput(new Date(Date.now() + 30 * 86400000)))
const weather = ref('晴朗，微风')
const recommendation = ref<FeedingRecommendation | null>(null)
const emptyForm = (): FeedingPlanInput => ({ pondId: 0, name: '', dailyAmountKg: 0, frequencyPerDay: 3, feedType: '', targetGrowthStage: '', minOxygen: 5, startDate: '', endDate: '', rationale: '' })
const form = reactive<FeedingPlanInput>(emptyForm())

const pendingCount = computed(() => plans.value.filter((item) => item.status === 'pending').length)
const approvedCount = computed(() => plans.value.filter((item) => item.status === 'approved').length)
const dailyTotal = computed(() => plans.value.filter((item) => item.status === 'approved').reduce((sum, item) => sum + item.dailyAmountKg, 0))
const transitionTitle = computed(() => ({ submit: '提交审核', approve: '批准计划', revoke: '撤销计划' }[transitionAction.value]))

async function load() {
  loading.value = true
  try {
    const [result, pondResult] = await Promise.all([
      planApi.list({ page: Number(params.page), pageSize: 20, search: String(params.search), status: String(params.status), pondId: Number(params.pondId) || undefined }),
      pondApi.list({ page: 1, pageSize: 100 }),
    ])
    plans.value = result.items
    total.value = result.total
    ponds.value = pondResult.items
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, emptyForm(), { pondId: ponds.value.find((item) => item.status !== 'closed')?.id || 0 })
  startLocal.value = toLocalInput()
  endLocal.value = toLocalInput(new Date(Date.now() + 30 * 86400000))
  editorOpen.value = true
}

function openEdit(plan: FeedingPlan) {
  editingId.value = plan.id
  Object.assign(form, {
    pondId: plan.pondId, name: plan.name, dailyAmountKg: plan.dailyAmountKg, frequencyPerDay: plan.frequencyPerDay,
    feedType: plan.feedType, targetGrowthStage: plan.targetGrowthStage, minOxygen: plan.minOxygen,
    startDate: plan.startDate, endDate: plan.endDate, rationale: plan.rationale,
  })
  startLocal.value = toLocalInput(new Date(plan.startDate))
  endLocal.value = toLocalInput(new Date(plan.endDate))
  editorOpen.value = true
}

async function save() {
  if (!form.pondId || !form.name || !form.feedType || !form.rationale || form.dailyAmountKg <= 0) {
    ElMessage.warning('请完整填写投喂计划')
    return
  }
  saving.value = true
  try {
    const payload = { ...form, startDate: toISO(startLocal.value), endDate: toISO(endLocal.value) }
    if (editingId.value) await planApi.update(editingId.value, payload)
    else await planApi.create(payload)
    ElMessage.success(editingId.value ? '计划已更新并生成新版本' : '计划草稿已创建')
    editorOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

function openTransition(plan: FeedingPlan, action: 'submit' | 'approve' | 'revoke') {
  selected.value = plan
  transitionAction.value = action
  transitionReason.value = ''
  transitionOpen.value = true
}

async function transition() {
  if (!selected.value || transitionReason.value.trim().length < 2) {
    ElMessage.warning('请填写状态变更原因')
    return
  }
  saving.value = true
  try {
    await planApi.transition(selected.value.id, transitionAction.value, transitionReason.value)
    ElMessage.success(`${transitionTitle.value}成功`)
    transitionOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

async function remove(plan: FeedingPlan) {
  try {
    await planApi.remove(plan.id)
    ElMessage.success('草稿计划已删除')
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  }
}

async function generateRecommendation(plan: FeedingPlan) {
  selected.value = plan
  recommendation.value = null
  recommendationOpen.value = true
  saving.value = true
  try {
    recommendation.value = await planApi.recommendation(plan.pondId, weather.value)
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
      <MetricCard label="计划总数" :value="total" :icon="DocumentChecked" />
      <MetricCard label="待审核" :value="pendingCount" :icon="Clock" tone="amber" hint="等待主管复核" />
      <MetricCard label="已批准" :value="approvedCount" :icon="CircleCheck" tone="blue" />
      <MetricCard label="已批准日投喂" :value="`${formatNumber(dailyTotal)} kg`" :icon="TrendCharts" tone="green" />
    </section>
    <section class="workspace-panel">
      <div class="panel-toolbar">
        <div class="filters">
          <el-input v-model="params.search" clearable placeholder="搜索计划名称" :prefix-icon="Search" />
          <el-select v-model="params.pondId" placeholder="全部养殖池" clearable><el-option v-for="pond in ponds" :key="pond.id" :label="pond.name" :value="String(pond.id)" /></el-select>
          <el-select v-model="params.status" placeholder="全部状态" clearable><el-option label="草稿" value="draft" /><el-option label="待审核" value="pending" /><el-option label="已批准" value="approved" /><el-option label="已执行" value="executed" /></el-select>
        </div>
        <el-button v-if="canOperate()" type="primary" :icon="Plus" @click="openCreate">新建计划</el-button>
      </div>
      <el-table v-loading="loading" :data="plans" stripe empty-text="暂无投喂计划">
        <el-table-column label="计划" min-width="230"><template #default="{ row }"><button class="table-link primary-cell" @click="selected = row; drawerOpen = true"><strong>{{ row.name }}</strong><small>{{ row.pond?.name }} · v{{ row.version }}</small></button></template></el-table-column>
        <el-table-column label="日投喂 / 频次" min-width="150"><template #default="{ row }">{{ row.dailyAmountKg }} kg / {{ row.frequencyPerDay }} 次</template></el-table-column>
        <el-table-column label="饲料 / 阶段" min-width="160"><template #default="{ row }"><div class="primary-cell"><span>{{ row.feedType }}</span><small>{{ row.targetGrowthStage }}</small></div></template></el-table-column>
        <el-table-column label="计划周期" min-width="170"><template #default="{ row }">{{ formatDateTime(row.startDate).slice(0, 10) }} 至 {{ formatDateTime(row.endDate).slice(0, 10) }}</template></el-table-column>
        <el-table-column label="状态" width="105"><template #default="{ row }"><StatusBadge :status="row.status" /></template></el-table-column>
        <el-table-column v-if="canOperate()" label="操作" min-width="250" fixed="right"><template #default="{ row }">
          <el-button v-if="row.status === 'draft'" link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="row.status === 'draft'" link type="primary" @click="openTransition(row, 'submit')">提交</el-button>
          <el-button v-if="row.status === 'pending' && canReview()" link type="success" @click="openTransition(row, 'approve')">批准</el-button>
          <el-button v-if="(row.status === 'pending' || row.status === 'approved') && canReview()" link type="warning" @click="openTransition(row, 'revoke')">撤销</el-button>
          <el-button v-if="row.status === 'approved'" link type="primary" @click="generateRecommendation(row)">投喂建议</el-button>
          <el-popconfirm v-if="row.status === 'draft' && canReview()" title="确认删除该草稿？" @confirm="remove(row)"><template #reference><el-button link type="danger">删除</el-button></template></el-popconfirm>
        </template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="params.page" layout="total, prev, pager, next" :total="total" :page-size="20" /></div>
    </section>
    <el-dialog v-model="editorOpen" :title="editingId ? '编辑计划新版本' : '新建投喂计划'" width="720px">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="养殖池"><el-select v-model="form.pondId"><el-option v-for="pond in ponds.filter((item) => item.status !== 'closed')" :key="pond.id" :label="`${pond.name} · ${pond.growthStage}`" :value="pond.id" /></el-select></el-form-item>
        <el-form-item label="计划名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="日投喂量（kg）"><el-input-number v-model="form.dailyAmountKg" :min="0.1" :step="1" /></el-form-item>
        <el-form-item label="每日频次"><el-input-number v-model="form.frequencyPerDay" :min="1" :max="12" /></el-form-item>
        <el-form-item label="饲料类型"><el-input v-model="form.feedType" /></el-form-item>
        <el-form-item label="目标生长阶段"><el-input v-model="form.targetGrowthStage" /></el-form-item>
        <el-form-item label="最低溶解氧（mg/L）"><el-input-number v-model="form.minOxygen" :min="0" :max="20" :step="0.1" /></el-form-item>
        <span />
        <el-form-item label="开始时间"><el-date-picker v-model="startLocal" type="datetime" value-format="YYYY-MM-DDTHH:mm" /></el-form-item>
        <el-form-item label="结束时间"><el-date-picker v-model="endLocal" type="datetime" value-format="YYYY-MM-DDTHH:mm" /></el-form-item>
        <el-form-item label="制定依据" class="form-span"><el-input v-model="form.rationale" type="textarea" :rows="4" placeholder="说明生长阶段、水质依据和投喂目标" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="editorOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存草稿</el-button></template>
    </el-dialog>
    <el-dialog v-model="transitionOpen" :title="transitionTitle" width="500px">
      <el-alert v-if="transitionAction === 'approve'" title="批准时将校验养殖池状态、最新水质和溶解氧阈值" type="warning" :closable="false" show-icon />
      <el-form-item class="dialog-field" label="变更原因"><el-input v-model="transitionReason" type="textarea" :rows="4" /></el-form-item>
      <template #footer><el-button @click="transitionOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="transition">确认{{ transitionTitle }}</el-button></template>
    </el-dialog>
    <el-dialog v-model="recommendationOpen" title="实时投喂建议" width="560px">
      <el-form-item label="天气窗口"><div class="inline-action"><el-input v-model="weather" placeholder="例：晴朗、小雨、暴雨" /><el-button :loading="saving" @click="selected && generateRecommendation(selected)">重新计算</el-button></div></el-form-item>
      <div v-if="recommendation" class="recommendation-result" :data-action="recommendation.action">
        <div class="recommendation-main"><span>{{ recommendation.action === 'hold' ? '暂停投喂' : recommendation.action === 'reduce' ? '减量投喂' : '按计划投喂' }}</span><strong>{{ recommendation.dailyAmountKg }} kg/日</strong><small>{{ recommendation.amountPerFeedingKg }} kg × {{ recommendation.frequencyPerDay }} 次</small></div>
        <ul><li v-for="reason in recommendation.reasons" :key="reason">{{ reason }}</li></ul>
      </div>
      <el-skeleton v-else :rows="4" animated />
    </el-dialog>
    <PlanDrawer v-model="drawerOpen" :plan="selected" />
  </div>
</template>
