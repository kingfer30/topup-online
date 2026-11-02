<template>
  <div>
    <n-card v-if="panel === 'platforms'" title="Azure 台子">
      <template #header-extra>
        <n-button type="primary" @click="openCreate">新增台子</n-button>
      </template>
      <n-data-table
        :columns="platformColumns"
        :data="platforms"
        :loading="loading"
        :bordered="false"
        :single-line="false"
        :scroll-x="1280"
        :row-key="(row: ApiPlatform) => row.id"
      />
    </n-card>

    <n-modal v-model:show="showPlatformForm" preset="card" :title="editingId ? '编辑台子' : '新增台子'" style="width: 560px">
      <n-form label-placement="left" label-width="90">
        <n-form-item label="台子地址">
          <n-input v-model:value="form.base_url" placeholder="https://example.com" />
        </n-form-item>
        <n-form-item label="API Key">
          <n-input v-model:value="form.api_key" type="password" show-password-on="click" placeholder="访问令牌" />
        </n-form-item>
        <n-form-item label="账号">
          <n-input v-model:value="form.account" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="form.password" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="form.remark" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showPlatformForm = false">取消</n-button>
          <n-button type="primary" :loading="savingPlatform" @click="submitPlatform">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-card v-if="panel === 'channels'">
      <template #header>
        <n-space align="center">
          <n-button size="small" @click="panel = 'platforms'">返回</n-button>
          <span>渠道管理 · {{ currentPlatform?.base_url }}</span>
        </n-space>
      </template>
      <template #header-extra>
        <n-space>
          <n-button @click="openBatchPrices">更新成本价|销售价</n-button>
          <n-button @click="openPull">手动拉取渠道</n-button>
          <n-button type="primary" :loading="syncing" @click="syncChannels">手动同步</n-button>
        </n-space>
      </template>
      <n-space class="mb-3" align="center">
        <n-select
          v-model:value="channelFilter.status"
          :options="channelStatusOptions"
          clearable
          placeholder="状态"
          style="width: 140px"
        />
        <n-input v-model:value="channelFilter.channelId" clearable placeholder="渠道 ID" style="width: 140px" />
        <n-input v-model:value="channelFilter.name" clearable placeholder="渠道名称" style="width: 180px" />
        <n-input v-model:value="channelFilter.model" clearable placeholder="可用模型" style="width: 200px" />
      </n-space>
      <n-data-table
        v-model:checked-row-keys="channelKeys"
        :columns="channelColumns"
        :data="filteredChannels"
        :loading="channelLoading"
        :scroll-x="1980"
        :bordered="false"
        size="small"
        :row-key="(row: ApiPlatformChannel) => row.id"
      />
    </n-card>

    <n-card v-if="panel === 'bills'">
      <template #header>
        <n-space align="center">
          <n-button size="small" @click="panel = 'platforms'">返回</n-button>
          <span>账单管理 · {{ currentPlatform?.base_url }}</span>
        </n-space>
      </template>
      <template #header-extra>
        <n-button type="primary" @click="openCreateBill">生成账单</n-button>
      </template>
      <n-data-table
        :columns="billColumns"
        :data="bills"
        :loading="billLoading"
        :scroll-x="1100"
        size="small"
        :row-key="(row: ApiPlatformBill) => row.id"
      />
    </n-card>

    <n-modal v-model:show="showBatchPrices" preset="card" title="更新成本价|销售价" style="width: 480px">
      <n-text depth="3" class="mb-3 block">将填写的价格写入已勾选的 {{ channelKeys.length }} 个渠道。留空的价格保持原值。</n-text>
      <n-form label-placement="left" label-width="80">
        <n-form-item label="成本价">
          <n-input-number v-model:value="batchPrices.cost_price" :min="0" :precision="4" clearable placeholder="人民币" style="width: 100%" />
        </n-form-item>
        <n-form-item label="销售价">
          <n-input-number v-model:value="batchPrices.sale_price" :min="0" :precision="4" clearable placeholder="人民币" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showBatchPrices = false">取消</n-button>
          <n-button type="primary" :loading="savingBatchPrices" @click="submitBatchPrices">更新</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showPull" preset="card" title="手动拉取渠道" style="width: 1100px">
      <n-space class="mb-3" align="center">
        <n-input-number v-model:value="pullForm.min_id" :min="0" placeholder="最小 ID" clearable style="width: 140px" />
        <n-input-number v-model:value="pullForm.max_id" :min="0" placeholder="最大 ID" clearable style="width: 140px" />
        <n-input v-model:value="pullForm.keyword" placeholder="渠道名称" style="width: 180px" />
        <n-select v-model:value="pullForm.status" :options="statusOptions" style="width: 140px" />
        <n-button type="primary" :loading="pulling" @click="doPull">拉取</n-button>
      </n-space>
      <n-data-table
        v-model:checked-row-keys="pulledKeys"
        :columns="pulledColumns"
        :data="pulled"
        :loading="pulling"
        :scroll-x="1100"
        size="small"
        :max-height="420"
        :row-key="(row: PulledChannel) => row.channel_id"
      />
      <n-space class="mt-3" align="center">
        <span>成本价</span>
        <n-input-number v-model:value="pullPrices.cost_price" :min="0" :precision="4" placeholder="人民币" style="width: 160px" />
        <span>销售价</span>
        <n-input-number v-model:value="pullPrices.sale_price" :min="0" :precision="4" placeholder="人民币" style="width: 160px" />
        <n-button :disabled="pulledKeys.length === 0" @click="applyPullPrices">写入勾选渠道</n-button>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showPull = false">关闭</n-button>
          <n-button type="primary" :loading="savingChannels" :disabled="pulledKeys.length === 0" @click="savePulled">
            保存渠道
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showModels" preset="card" title="更新模型" style="width: 640px">
      <n-space vertical>
        <div>
          <div class="mb-2">现有模型</div>
          <n-space>
            <n-tag v-for="name in editingModels" :key="name" closable @close="removeEditingModel(name)">{{ name }}</n-tag>
            <n-text v-if="editingModels.length === 0" depth="3">暂无模型</n-text>
          </n-space>
        </div>
        <n-space align="center">
          <n-select
            v-model:value="modelsToAdd"
            multiple
            :options="addableModelOptions"
            placeholder="选择要添加的模型"
            style="width: 360px"
          />
          <n-button :disabled="modelsToAdd.length === 0" @click="addEditingModels">添加</n-button>
        </n-space>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModels = false">取消</n-button>
          <n-button type="primary" :loading="savingModels" @click="saveModels">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showTestResult" preset="card" title="测试结果" style="width: 560px">
      <n-descriptions :column="1" bordered size="small">
        <n-descriptions-item label="模型">{{ testResult.model }}</n-descriptions-item>
        <n-descriptions-item label="状态">{{ testResult.status_code }}</n-descriptions-item>
        <n-descriptions-item label="返回">{{ testResult.message }}</n-descriptions-item>
      </n-descriptions>
    </n-modal>

    <n-modal v-model:show="showCreateBill" preset="card" title="选择渠道生成账单" style="width: 860px">
      <n-text depth="3" class="mb-3 block">待支付额度取当前未结算。已在其他待支付账单中的渠道不能再选。</n-text>
      <n-data-table
        v-model:checked-row-keys="billChannelKeys"
        :columns="billPickColumns"
        :data="channels"
        :loading="channelLoading"
        :scroll-x="860"
        :max-height="420"
        size="small"
        :row-key="(row: ApiPlatformChannel) => row.id"
      />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreateBill = false">取消</n-button>
          <n-button type="primary" :loading="creatingBill" :disabled="billChannelKeys.length === 0" @click="submitBill">
            生成账单
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDropdown,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSelect,
  NSpace,
  NTag,
  NText,
  useMessage,
  type DataTableColumns,
  type DropdownOption,
} from 'naive-ui'
import {
  createApiPlatform,
  createApiPlatformBill,
  deleteApiPlatform,
  getAdminApiPlatformSettings,
  getApiPlatform,
  listApiPlatformBills,
  listApiPlatformChannels,
  listApiPlatforms,
  payApiPlatformBill,
  pullApiPlatformChannels,
  saveApiPlatformChannels,
  syncApiPlatformChannels,
  batchUpdateApiPlatformChannelPrices,
  testApiPlatformChannel,
  updateApiPlatform,
  updateApiPlatformChannelModels,
  updateApiPlatformChannelMoney,
  updateApiPlatformChannelStatus,
  voidApiPlatformBill,
  type ApiPlatform,
  type ApiPlatformBill,
  type ApiPlatformChannel,
  type PulledChannel,
} from '@/api/api-platform'

const message = useMessage()
const loading = ref(false)
const platforms = ref<ApiPlatform[]>([])
const testModels = ref<string[]>([])

const showPlatformForm = ref(false)
const savingPlatform = ref(false)
const editingId = ref(0)
const form = ref({ base_url: '', api_key: '', account: '', password: '', remark: '' })

const panel = ref<'platforms' | 'channels' | 'bills'>('platforms')
const currentPlatform = ref<ApiPlatform | null>(null)
const channelLoading = ref(false)
const channels = ref<ApiPlatformChannel[]>([])
const channelKeys = ref<number[]>([])
const channelFilter = ref<{ status: number | null; channelId: string; name: string; model: string }>({
  status: null,
  channelId: '',
  name: '',
  model: '',
})
const channelStatusOptions = [
  { label: '启用', value: 1 },
  { label: '手动禁用', value: 2 },
  { label: '自动禁用', value: 3 },
]
const filteredChannels = computed(() => {
  const channelId = channelFilter.value.channelId.trim()
  const name = channelFilter.value.name.trim().toLowerCase()
  const model = channelFilter.value.model.trim().toLowerCase()
  const status = channelFilter.value.status
  return channels.value.filter((row) => {
    if (status != null && row.status !== status) return false
    if (channelId && !String(row.channel_id).includes(channelId)) return false
    if (name && !row.name.toLowerCase().includes(name)) return false
    if (model && !(row.models || '').toLowerCase().includes(model)) return false
    return true
  })
})
const showBatchPrices = ref(false)
const savingBatchPrices = ref(false)
const batchPrices = ref<{ cost_price: number | null; sale_price: number | null }>({
  cost_price: null,
  sale_price: null,
})
const syncing = ref(false)
const testingId = ref(0)
const statusChangingId = ref(0)

const showPull = ref(false)
const pulling = ref(false)
const savingChannels = ref(false)
const pulled = ref<PulledChannel[]>([])
const pulledKeys = ref<number[]>([])
const pullPrices = ref<{ cost_price: number | null; sale_price: number | null }>({
  cost_price: null,
  sale_price: null,
})
const pullForm = ref<{ min_id: number | null; max_id: number | null; keyword: string; status: string }>({
  min_id: null,
  max_id: null,
  keyword: '',
  status: '',
})
const statusOptions = [
  { label: '不选', value: '' },
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' },
]

const showModels = ref(false)
const savingModels = ref(false)
const editingChannel = ref<ApiPlatformChannel | null>(null)
const editingModels = ref<string[]>([])
const modelsToAdd = ref<string[]>([])

const showTestResult = ref(false)
const testResult = ref({ model: '', status_code: 0, message: '' })

const showCreateBill = ref(false)
const bills = ref<ApiPlatformBill[]>([])
const billLoading = ref(false)
const creatingBill = ref(false)
const billChannelKeys = ref<number[]>([])
const moneySnapshot = ref<Record<number, string>>({})

const fmtTime = (value?: string | null) => {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false, timeZone: 'Asia/Shanghai' })
}
const fmtUSD = (value?: number | null) =>
  `$${Number(value || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
const fmtCNY = (value?: number | null) =>
  `¥${Number(value || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`

const modelOptions = computed<DropdownOption[]>(() => testModels.value.map((name) => ({ label: name, key: name })))
const addableModelOptions = computed(() =>
  testModels.value.filter((name) => !editingModels.value.includes(name)).map((name) => ({ label: name, value: name }))
)

const loadPlatforms = async () => {
  loading.value = true
  try {
    const res = await listApiPlatforms()
    if (res.code === 200) platforms.value = res.data || []
    else message.error(res.message || '加载失败')
  } catch {
    message.error('加载台子失败')
  } finally {
    loading.value = false
  }
}

const loadSettings = async () => {
  try {
    const res = await getAdminApiPlatformSettings()
    if (res.code === 200) testModels.value = res.data?.test_models || []
  } catch {
    testModels.value = []
  }
}

const loadChannels = async () => {
  if (!currentPlatform.value) return
  channelLoading.value = true
  try {
    const res = await listApiPlatformChannels(currentPlatform.value.id)
    if (res.code === 200) {
      channels.value = res.data || []
      const next: Record<number, string> = {}
      for (const row of channels.value) {
      next[row.id] = `${row.cost_price ?? 0}|${row.sale_price ?? 0}`
      }
      moneySnapshot.value = next
    } else message.error(res.message || '加载渠道失败')
  } catch {
    message.error('加载渠道失败')
  } finally {
    channelLoading.value = false
  }
}

const loadBills = async () => {
  if (!currentPlatform.value) return
  billLoading.value = true
  try {
    const res = await listApiPlatformBills(currentPlatform.value.id)
    if (res.code === 200) bills.value = res.data || []
    else message.error(res.message || '加载账单失败')
  } catch {
    message.error('加载账单失败')
  } finally {
    billLoading.value = false
  }
}

const openCreate = () => {
  editingId.value = 0
  form.value = { base_url: '', api_key: '', account: '', password: '', remark: '' }
  showPlatformForm.value = true
}

const openEdit = async (row: ApiPlatform) => {
  editingId.value = row.id
  try {
    const res = await getApiPlatform(row.id)
    if (res.code !== 200 || !res.data) {
      message.error(res.message || '读取台子失败')
      return
    }
    form.value = {
      base_url: res.data.base_url,
      api_key: res.data.api_key || '',
      account: res.data.account || '',
      password: res.data.password || '',
      remark: res.data.remark || '',
    }
    showPlatformForm.value = true
  } catch {
    message.error('读取台子失败')
  }
}

const submitPlatform = async () => {
  savingPlatform.value = true
  try {
    const res = editingId.value
      ? await updateApiPlatform(editingId.value, form.value)
      : await createApiPlatform(form.value)
    if (res.code === 200) {
      message.success('已保存')
      showPlatformForm.value = false
      await loadPlatforms()
    } else message.error(res.message || '保存失败')
  } catch {
    message.error('保存失败')
  } finally {
    savingPlatform.value = false
  }
}

const removePlatform = async (row: ApiPlatform) => {
  try {
    const res = await deleteApiPlatform(row.id)
    if (res.code === 200) {
      message.success('已删除')
      await loadPlatforms()
    } else message.error(res.message || '删除失败')
  } catch {
    message.error('删除失败')
  }
}

const openChannels = async (row: ApiPlatform) => {
  currentPlatform.value = row
  channelFilter.value = { status: null, channelId: '', name: '', model: '' }
  channelKeys.value = []
  panel.value = 'channels'
  await loadChannels()
}

const openBills = async (row: ApiPlatform) => {
  currentPlatform.value = row
  panel.value = 'bills'
  await loadBills()
}

const openPull = () => {
  pullForm.value = { min_id: null, max_id: null, keyword: '', status: '' }
  pullPrices.value = { cost_price: null, sale_price: null }
  pulled.value = []
  pulledKeys.value = []
  showPull.value = true
}

const doPull = async () => {
  if (!currentPlatform.value) return
  pulling.value = true
  try {
    const res = await pullApiPlatformChannels(currentPlatform.value.id, {
      min_id: pullForm.value.min_id || 0,
      max_id: pullForm.value.max_id || 0,
      keyword: pullForm.value.keyword,
      status: pullForm.value.status,
    })
    if (res.code === 200) {
      pulled.value = res.data || []
      pulledKeys.value = []
      message.success(`拉取到 ${pulled.value.length} 条`)
    } else message.error(res.message || '拉取失败')
  } catch {
    message.error('拉取失败')
  } finally {
    pulling.value = false
  }
}

const applyPullPrices = () => {
  if (pulledKeys.value.length === 0) return
  if (pullPrices.value.cost_price == null && pullPrices.value.sale_price == null) {
    message.warning('请先填写成本价或销售价')
    return
  }
  const selected = new Set(pulledKeys.value)
  pulled.value = pulled.value.map((row) => {
    if (!selected.has(row.channel_id)) return row
    return {
      ...row,
      cost_price: pullPrices.value.cost_price ?? row.cost_price,
      sale_price: pullPrices.value.sale_price ?? row.sale_price,
    }
  })
  message.success(`已写入 ${selected.size} 个渠道，保存后生效`)
}

const savePulled = async () => {
  if (!currentPlatform.value) return
  const selected = pulled.value.filter((row) => pulledKeys.value.includes(row.channel_id))
  if (selected.length === 0) return
  savingChannels.value = true
  try {
    const res = await saveApiPlatformChannels(currentPlatform.value.id, {
      filter_name: pullForm.value.keyword.trim(),
      cost_price: pullPrices.value.cost_price,
      sale_price: pullPrices.value.sale_price,
      channels: selected,
    })
    if (res.code === 200) {
      message.success('渠道已保存')
      showPull.value = false
      await Promise.all([loadChannels(), loadPlatforms()])
    } else message.error(res.message || '保存失败')
  } catch {
    message.error('保存失败')
  } finally {
    savingChannels.value = false
  }
}

const openBatchPrices = () => {
  if (channelKeys.value.length === 0) {
    message.warning('请先勾选渠道')
    return
  }
  batchPrices.value = { cost_price: null, sale_price: null }
  showBatchPrices.value = true
}

const submitBatchPrices = async () => {
  if (!currentPlatform.value) return
  if (batchPrices.value.cost_price == null && batchPrices.value.sale_price == null) {
    message.warning('请填写成本价或销售价')
    return
  }
  savingBatchPrices.value = true
  try {
    const res = await batchUpdateApiPlatformChannelPrices(currentPlatform.value.id, {
      channel_row_ids: channelKeys.value,
      cost_price: batchPrices.value.cost_price,
      sale_price: batchPrices.value.sale_price,
    })
    if (res.code === 200) {
      message.success(`已更新 ${res.data?.updated || channelKeys.value.length} 个渠道的价格`)
      showBatchPrices.value = false
      channelKeys.value = []
      await loadChannels()
    } else message.error(res.message || '更新失败')
  } catch {
    message.error('更新失败')
  } finally {
    savingBatchPrices.value = false
  }
}

const syncChannels = async () => {
  if (!currentPlatform.value) return
  syncing.value = true
  try {
    const res = await syncApiPlatformChannels(currentPlatform.value.id)
    if (res.code === 200) {
      message.success(`已更新 ${res.data?.updated || 0} 条，失败 ${res.data?.failed || 0} 条`)
      await Promise.all([loadChannels(), loadPlatforms()])
    } else message.error(res.message || '同步失败')
  } catch {
    message.error('同步失败')
  } finally {
    syncing.value = false
  }
}

const saveMoney = async (row: ApiPlatformChannel) => {
  if (!currentPlatform.value) return
  const signature = `${row.cost_price ?? 0}|${row.sale_price ?? 0}`
  if (moneySnapshot.value[row.id] === signature) return
  try {
    const res = await updateApiPlatformChannelMoney(currentPlatform.value.id, row.id, {
      cost_price: row.cost_price ?? 0,
      sale_price: row.sale_price ?? 0,
    })
    if (res.code === 200) {
      moneySnapshot.value[row.id] = signature
      message.success('成本价和销售价已更新')
    } else message.error(res.message || '保存失败')
  } catch {
    message.error('保存失败')
  }
}

const changeChannelStatus = async (row: ApiPlatformChannel, status: number) => {
  if (!currentPlatform.value || statusChangingId.value) return
  statusChangingId.value = row.id
  try {
    const res = await updateApiPlatformChannelStatus(currentPlatform.value.id, row.id, status)
    if (res.code === 200) {
      message.success(status === 1 ? '渠道已启用' : '渠道已禁用')
      await loadChannels()
    } else message.error(res.message || '更新状态失败')
  } catch {
    message.error('更新状态失败')
  } finally {
    statusChangingId.value = 0
  }
}

const runTest = async (row: ApiPlatformChannel, model: string) => {
  if (!currentPlatform.value || testingId.value) return
  testingId.value = row.id
  try {
    const res = await testApiPlatformChannel(currentPlatform.value.id, row.id, model)
    if (res.code === 200 && res.data) {
      testResult.value = res.data
      showTestResult.value = true
      await loadChannels()
    } else message.error(res.message || '测试失败')
  } catch {
    message.error('测试失败')
  } finally {
    testingId.value = 0
  }
}

const openModels = (row: ApiPlatformChannel) => {
  editingChannel.value = row
  editingModels.value = (row.models || '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
  modelsToAdd.value = []
  showModels.value = true
}

const removeEditingModel = (name: string) => {
  editingModels.value = editingModels.value.filter((item) => item !== name)
}

const addEditingModels = () => {
  const next = modelsToAdd.value.filter((name) => name && !editingModels.value.includes(name))
  if (next.length === 0) return
  editingModels.value = [...editingModels.value, ...next]
  modelsToAdd.value = []
}

const saveModels = async () => {
  if (!currentPlatform.value || !editingChannel.value) return
  savingModels.value = true
  try {
    const res = await updateApiPlatformChannelModels(currentPlatform.value.id, editingChannel.value.id, editingModels.value)
    if (res.code === 200) {
      message.success('模型已更新')
      showModels.value = false
      await loadChannels()
    } else message.error(res.message || '更新失败')
  } catch {
    message.error('更新失败')
  } finally {
    savingModels.value = false
  }
}

const openCreateBill = async () => {
  billChannelKeys.value = []
  await loadChannels()
  showCreateBill.value = true
}

const submitBill = async () => {
  if (!currentPlatform.value || billChannelKeys.value.length === 0) return
  creatingBill.value = true
  try {
    const res = await createApiPlatformBill(currentPlatform.value.id, billChannelKeys.value)
    if (res.code === 200) {
      message.success(`已生成 ${res.data?.bill_no || '账单'}`)
      showCreateBill.value = false
      await Promise.all([loadBills(), loadChannels()])
    } else message.error(res.message || '生成失败')
  } catch {
    message.error('生成失败')
  } finally {
    creatingBill.value = false
  }
}

const markPaid = async (row: ApiPlatformBill) => {
  if (!currentPlatform.value) return
  try {
    const res = await payApiPlatformBill(currentPlatform.value.id, row.id)
    if (res.code === 200) {
      message.success('已标记支付')
      await Promise.all([loadBills(), loadChannels(), loadPlatforms()])
    } else message.error(res.message || '操作失败')
  } catch {
    message.error('操作失败')
  }
}

const markVoid = async (row: ApiPlatformBill) => {
  if (!currentPlatform.value) return
  try {
    const res = await voidApiPlatformBill(currentPlatform.value.id, row.id)
    if (res.code === 200) {
      message.success('账单已作废')
      await Promise.all([loadBills(), loadChannels()])
    } else message.error(res.message || '操作失败')
  } catch {
    message.error('操作失败')
  }
}

const openPdf = (row: ApiPlatformBill) => {
  if (!currentPlatform.value) return
  window.open(`/bill-print/${currentPlatform.value.id}/${row.id}`, '_blank')
}

const moneyInput = (row: ApiPlatformChannel, field: 'cost_price' | 'sale_price') =>
  h(NInputNumber, {
    value: row[field],
    min: 0,
    precision: 4,
    size: 'small',
    showButton: false,
    style: 'width: 110px',
    onUpdateValue: (value: number | null) => {
      row[field] = value ?? 0
    },
    onBlur: () => saveMoney(row),
  })

const platformColumns: DataTableColumns<ApiPlatform> = [
  { title: '台子地址', key: 'base_url', ellipsis: { tooltip: true }, minWidth: 220 },
  { title: 'API Key', key: 'api_key_mask', width: 160 },
  { title: '账号', key: 'account', width: 140 },
  { title: '密码', key: 'password_mask', width: 120 },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true }, width: 160 },
  { title: '消耗总额', key: 'used_amount', width: 130, render: (row) => fmtUSD(row.used_amount) },
  { title: '已结算金额', key: 'settled_amount', width: 130, render: (row) => fmtUSD(row.settled_amount) },
  { title: '未结算金额', key: 'unsettled_amount', width: 130, render: (row) => fmtUSD(row.unsettled_amount) },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 6 }, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => openEdit(row) }, { default: () => '编辑' }),
          h(NButton, { size: 'small', type: 'primary', onClick: () => openChannels(row) }, { default: () => '渠道管理' }),
          h(NButton, { size: 'small', onClick: () => openBills(row) }, { default: () => '账单管理' }),
          h(NPopconfirm, { onPositiveClick: () => removePlatform(row) }, {
            trigger: () => h(NButton, { size: 'small', type: 'error', ghost: true }, { default: () => '删除' }),
            default: () => '台子会从列表移除，渠道和账单仍保留。确认删除？',
          }),
        ],
      }),
  },
]

const channelColumns: DataTableColumns<ApiPlatformChannel> = [
  { type: 'selection' },
  { title: '渠道 ID', key: 'channel_id', width: 90 },
  { title: '渠道名称', key: 'name', ellipsis: { tooltip: true }, minWidth: 180 },
  { title: '筛选名称', key: 'filter_name', width: 120 },
  { title: '可用模型', key: 'models', ellipsis: { tooltip: true }, minWidth: 180 },
  {
    title: '状态',
    key: 'status_label',
    width: 110,
    render: (row) => h(NTag, { size: 'small', type: row.status === 1 ? 'success' : 'warning' }, { default: () => row.status_label }),
  },
  { title: '消耗总额', key: 'used_amount', width: 120, render: (row) => fmtUSD(row.used_amount) },
  { title: '已结算', key: 'settled_amount', width: 120, render: (row) => fmtUSD(row.settled_amount) },
  { title: '未结算', key: 'unsettled_amount', width: 120, render: (row) => fmtUSD(row.unsettled_amount) },
  { title: '成本价', key: 'cost_price', width: 130, render: (row) => moneyInput(row, 'cost_price') },
  { title: '销售价', key: 'sale_price', width: 130, render: (row) => moneyInput(row, 'sale_price') },
  { title: '创建时间', key: 'created_at', width: 170, render: (row) => fmtTime(row.created_at) },
  { title: '更新时间', key: 'updated_at', width: 170, render: (row) => fmtTime(row.updated_at) },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 6 }, {
        default: () => [
          row.status === 1
            ? h(NButton, {
                size: 'small',
                type: 'warning',
                ghost: true,
                loading: statusChangingId.value === row.id,
                onClick: () => changeChannelStatus(row, 2),
              }, { default: () => '禁用' })
            : h(NButton, {
                size: 'small',
                type: 'success',
                ghost: true,
                loading: statusChangingId.value === row.id,
                onClick: () => changeChannelStatus(row, 1),
              }, { default: () => '启用' }),
          h(NDropdown, {
            options: modelOptions.value,
            onSelect: (key: string | number) => runTest(row, String(key)),
          }, {
            default: () => h(NButton, { size: 'small', loading: testingId.value === row.id }, { default: () => '测试' }),
          }),
          h(NButton, { size: 'small', onClick: () => openModels(row) }, { default: () => '更新模型' }),
        ],
      }),
  },
]

const pulledColumns: DataTableColumns<PulledChannel> = [
  { type: 'selection' },
  { title: '渠道 ID', key: 'channel_id', width: 90 },
  { title: '渠道名称', key: 'name', ellipsis: { tooltip: true } },
  { title: '可用模型', key: 'models', ellipsis: { tooltip: true } },
  { title: '状态', key: 'status_label', width: 110 },
  { title: '消耗总额', key: 'used_amount', width: 130, render: (row) => fmtUSD(row.used_amount) },
  { title: '成本价', key: 'cost_price', width: 110, render: (row) => (row.cost_price == null ? '—' : fmtCNY(row.cost_price)) },
  { title: '销售价', key: 'sale_price', width: 110, render: (row) => (row.sale_price == null ? '—' : fmtCNY(row.sale_price)) },
]

const billColumns: DataTableColumns<ApiPlatformBill> = [
  { title: '账单编号', key: 'bill_no', width: 130 },
  { title: '生成时间', key: 'created_at', width: 180, render: (row) => fmtTime(row.created_at) },
  { title: '渠道数', key: 'channel_count', width: 80 },
  { title: '待支付额度', key: 'pay_quota', width: 140, render: (row) => fmtUSD(row.pay_quota) },
  { title: '金额', key: 'amount_cny', width: 140, render: (row) => fmtCNY(row.amount_cny) },
  { title: '状态', key: 'status_label', width: 90 },
  { title: '支付时间', key: 'paid_at', width: 180, render: (row) => fmtTime(row.paid_at) },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render: (row) => {
      const buttons = [
        h(NButton, { size: 'small', onClick: () => openPdf(row) }, { default: () => '生成 PDF' }),
      ]
      if (row.status === 1) {
        buttons.unshift(
          h(NPopconfirm, { onPositiveClick: () => markPaid(row) }, {
            trigger: () => h(NButton, { size: 'small', type: 'primary' }, { default: () => '已支付' }),
            default: () => '按账单冻结的待支付额度入账？',
          }),
          h(NPopconfirm, { onPositiveClick: () => markVoid(row) }, {
            trigger: () => h(NButton, { size: 'small', type: 'warning', ghost: true }, { default: () => '作废' }),
            default: () => '作废后可以重新生成这些渠道的账单。',
          })
        )
      }
      return h(NSpace, { size: 6 }, { default: () => buttons })
    },
  },
]

const billPickColumns: DataTableColumns<ApiPlatformChannel> = [
  { type: 'selection', disabled: (row) => !!row.in_pending_bill },
  { title: '渠道 ID', key: 'channel_id', width: 90 },
  { title: '渠道名称', key: 'name', ellipsis: { tooltip: true } },
  { title: '未结算', key: 'unsettled_amount', width: 120, render: (row) => fmtUSD(row.unsettled_amount) },
  { title: '销售价', key: 'sale_price', width: 100, render: (row) => fmtCNY(row.sale_price) },
  {
    title: '金额',
    key: 'amount',
    width: 120,
    render: (row) => fmtCNY((row.sale_price || 0) * (row.unsettled_amount || 0)),
  },
  {
    title: '说明',
    key: 'in_pending_bill',
    width: 140,
    render: (row) => (row.in_pending_bill ? '已有待支付账单' : ''),
  },
]

onMounted(() => {
  loadPlatforms()
  loadSettings()
})
</script>
