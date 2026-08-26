<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleCheck, Grid, Plus, Search, Warning } from '@element-plus/icons-vue'
import { pondApi } from '@/api/ponds'
import { readingApi } from '@/api/readings'
import MetricCard from '@/components/common/MetricCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import RiskTag from '@/components/common/RiskTag.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useAuth } from '@/hooks/useAuth'
import { useQueryParams } from '@/hooks/useQueryParams'
import type { Pond, PondInput, WaterReading } from '@/types/models'
import type { RiskLevel } from '@/types/enums'
import { formatNumber } from '@/utils/format'
import { errorMessage } from '@/utils/errors'

const { canManagePonds } = useAuth()
const { params } = useQueryParams({ search: '', status: '', page: 1 })
const loading = ref(false)
const saving = ref(false)
const ponds = ref<Pond[]>([])
const total = ref(0)
const readings = ref<WaterReading[]>([])
const editorOpen = ref(false)
const deleteOpen = ref(false)
const editingId = ref<number | null>(null)
const target = ref<Pond | null>(null)
const emptyForm = (): PondInput => ({ code: '', name: '', species: '', areaSquareMeters: 0, capacityKg: 0, growthStage: '', status: 'active', manager: '', notes: '' })
const form = reactive<PondInput>(emptyForm())

const activeCount = computed(() => ponds.value.filter((item) => item.status === 'active').length)
const quarantineCount = computed(() => ponds.value.filter((item) => item.status === 'quarantine').length)
const capacity = computed(() => ponds.value.reduce((sum, item) => sum + item.capacityKg, 0))
const latestRisk = computed(() => {
  const map = new Map<number, RiskLevel>()
  for (const item of readings.value) if (!map.has(item.pondId)) map.set(item.pondId, item.riskLevel)
  return map
})

async function load() {
  loading.value = true
  try {
    const [pondResult, readingResult] = await Promise.all([
      pondApi.list({ page: Number(params.page), pageSize: 20, search: String(params.search), status: String(params.status) }),
      readingApi.list({ page: 1, pageSize: 100 }),
    ])
    ponds.value = pondResult.items
    total.value = pondResult.total
    readings.value = readingResult.items
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, emptyForm())
  editorOpen.value = true
}

function openEdit(pond: Pond) {
  editingId.value = pond.id
  Object.assign(form, {
    code: pond.code, name: pond.name, species: pond.species, areaSquareMeters: pond.areaSquareMeters,
    capacityKg: pond.capacityKg, growthStage: pond.growthStage, status: pond.status, manager: pond.manager, notes: pond.notes,
  })
  editorOpen.value = true
}

async function save() {
  if (!form.code || !form.name || !form.species || form.areaSquareMeters <= 0 || form.capacityKg <= 0) {
    ElMessage.warning('请完整填写养殖池必填项')
    return
  }
  saving.value = true
  try {
    if (editingId.value) await pondApi.update(editingId.value, { ...form })
    else await pondApi.create({ ...form })
    ElMessage.success(editingId.value ? '养殖池已更新' : '养殖池已创建')
    editorOpen.value = false
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
    await pondApi.remove(target.value.id)
    ElMessage.success('养殖池已删除')
    deleteOpen.value = false
    await load()
  } catch (error) {
    ElMessage.error(errorMessage(error))
  } finally {
    saving.value = false
  }
}

let refreshTimer: number | undefined
watch(params, () => {
  window.clearTimeout(refreshTimer)
  refreshTimer = window.setTimeout(load, 250)
}, { deep: true })
onMounted(load)
</script>

<template>
  <div class="page-stack">
    <section class="metrics-grid">
      <MetricCard label="养殖池总数" :value="total" :icon="Grid" hint="当前管辖" />
      <MetricCard label="运行中" :value="activeCount" :icon="CircleCheck" tone="blue" hint="可执行投喂" />
      <MetricCard label="隔离观察" :value="quarantineCount" :icon="Warning" tone="amber" hint="需续评水质" />
      <MetricCard label="页内养殖容量" :value="`${formatNumber(capacity / 1000)} t`" :icon="Grid" tone="green" />
    </section>
    <section class="workspace-panel">
      <div class="panel-toolbar">
        <div class="filters">
          <el-input v-model="params.search" clearable placeholder="搜索编码、名称或品种" :prefix-icon="Search" />
          <el-select v-model="params.status" placeholder="全部状态" clearable>
            <el-option label="运行中" value="active" /><el-option label="隔离观察" value="quarantine" /><el-option label="已关闭" value="closed" />
          </el-select>
        </div>
        <el-button v-if="canManagePonds()" type="primary" :icon="Plus" @click="openCreate">新建养殖池</el-button>
      </div>
      <el-table v-loading="loading" :data="ponds" stripe empty-text="暂无养殖池数据">
        <el-table-column label="养殖池" min-width="210"><template #default="{ row }"><div class="primary-cell"><strong>{{ row.name }}</strong><small>{{ row.code }} · {{ row.species }}</small></div></template></el-table-column>
        <el-table-column label="阶段" prop="growthStage" min-width="100" />
        <el-table-column label="面积 / 容量" min-width="150"><template #default="{ row }">{{ formatNumber(row.areaSquareMeters, 0) }} ㎡ / {{ formatNumber(row.capacityKg / 1000) }} t</template></el-table-column>
        <el-table-column label="负责人" prop="manager" min-width="100" />
        <el-table-column label="水质风险" width="100"><template #default="{ row }"><RiskTag :level="latestRisk.get(row.id) || (row.status === 'quarantine' ? 'warning' : 'normal')" /></template></el-table-column>
        <el-table-column label="状态" width="110"><template #default="{ row }"><StatusBadge :status="row.status" /></template></el-table-column>
        <el-table-column v-if="canManagePonds()" label="操作" width="130" fixed="right"><template #default="{ row }"><el-button link type="primary" @click="openEdit(row)">编辑</el-button><el-button link type="danger" @click="target = row; deleteOpen = true">删除</el-button></template></el-table-column>
      </el-table>
      <div class="pagination"><el-pagination v-model:current-page="params.page" layout="total, prev, pager, next" :total="total" :page-size="20" /></div>
    </section>
    <el-dialog v-model="editorOpen" :title="editingId ? '编辑养殖池' : '新建养殖池'" width="660px">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="编码"><el-input v-model="form.code" placeholder="P-C02" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="养殖品种"><el-input v-model="form.species" /></el-form-item>
        <el-form-item label="生长阶段"><el-input v-model="form.growthStage" /></el-form-item>
        <el-form-item label="面积（㎡）"><el-input-number v-model="form.areaSquareMeters" :min="1" :controls="false" /></el-form-item>
        <el-form-item label="容量（kg）"><el-input-number v-model="form.capacityKg" :min="1" :controls="false" /></el-form-item>
        <el-form-item label="负责人"><el-input v-model="form.manager" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="运行中" value="active" /><el-option label="隔离观察" value="quarantine" /><el-option label="已关闭" value="closed" /></el-select></el-form-item>
        <el-form-item label="备注" class="form-span"><el-input v-model="form.notes" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="editorOpen = false">取消</el-button><el-button type="primary" :loading="saving" @click="save">保存</el-button></template>
    </el-dialog>
    <ConfirmDialog v-model="deleteOpen" title="删除养殖池" :message="`确认删除「${target?.name || ''}」？已有关联数据时系统将拒绝。`" danger :loading="saving" @confirm="remove" />
  </div>
</template>
