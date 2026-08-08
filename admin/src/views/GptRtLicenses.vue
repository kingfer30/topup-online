<template>
  <div class="gpt-rt-licenses">
    <n-card title="GPT RT 许可证">
      <template #header-extra>
        <n-button type="primary" @click="showCreateModal = true">颁发许可证</n-button>
      </template>

      <n-space vertical :size="16">
        <n-space>
          <n-select
            v-model:value="searchStatus"
            :options="statusOptions"
            placeholder="状态"
            style="width: 130px"
          />
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索激活码/客户/IP"
            clearable
            style="width: 280px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <n-data-table
          remote
          :columns="columns"
          :data="list"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: GptRtLicense) => row.id"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
          :scroll-x="1500"
        />
      </n-space>
    </n-card>

    <n-modal
      v-model:show="showCreateModal"
      preset="card"
      title="颁发许可证"
      :style="{ width: '480px' }"
      :mask-closable="false"
    >
      <n-form ref="createFormRef" :model="createForm" label-placement="left" label-width="100px">
        <n-form-item label="客户备注">
          <n-input v-model:value="createForm.customer" placeholder="可选" />
        </n-form-item>
        <n-form-item label="应用标识">
          <n-input v-model:value="createForm.app_id" placeholder="默认 gpt-rt-register" />
        </n-form-item>
        <n-form-item label="有效期" path="months" :rule="{ required: true, type: 'number', message: '请选择有效期' }">
          <n-select v-model:value="createForm.months" :options="monthOptions" />
        </n-form-item>
        <n-form-item
          label="可用数量"
          path="available_count"
          :rule="{ required: true, type: 'number', min: 1, message: '可用数量必须为正整数' }"
        >
          <n-input-number v-model:value="createForm.available_count" :min="1" :precision="0" style="width: 100%" />
        </n-form-item>
        <n-form-item
          label="可绑设备数"
          path="max_devices"
          :rule="{ required: true, type: 'number', min: 1, message: '可绑定设备数必须为正整数' }"
        >
          <n-input-number v-model:value="createForm.max_devices" :min="1" :precision="0" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" :loading="createLoading" @click="handleCreate">颁发</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showCreatedKeyModal"
      preset="card"
      title="激活码已生成"
      :style="{ width: '520px' }"
    >
      <n-alert type="success" style="margin-bottom: 12px">请复制保存，关闭后可在列表中查看</n-alert>
      <n-input :value="createdLicenseKey" readonly type="textarea" :rows="2" />
      <template #footer>
        <n-space justify="end">
          <n-button type="primary" @click="copyCreatedKey">复制激活码</n-button>
          <n-button @click="showCreatedKeyModal = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showRenewModal"
      preset="card"
      title="续期"
      :style="{ width: '400px' }"
      :mask-closable="false"
    >
      <n-form label-placement="left" label-width="90px">
        <n-form-item label="续期月数">
          <n-select v-model:value="renewMonths" :options="monthOptions" />
        </n-form-item>
        <n-form-item label="可用数量">
          <span>续期后重置为 500</span>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRenewModal = false">取消</n-button>
          <n-button type="primary" :loading="renewLoading" @click="handleRenew">确认续期</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showAddCountModal"
      preset="card"
      title="增加可用数量"
      :style="{ width: '420px' }"
      :mask-closable="false"
    >
      <n-form label-placement="left" label-width="100px">
        <n-form-item label="当前可用">
          <span>{{ addCountCurrent }}</span>
        </n-form-item>
        <n-form-item label="增加数量">
          <n-input-number v-model:value="addCountValue" :min="1" :precision="0" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddCountModal = false">取消</n-button>
          <n-button type="primary" :loading="addCountLoading" @click="handleAddCount">确认增加</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showMaxDevicesModal"
      preset="card"
      title="修改绑定数量"
      :style="{ width: '420px' }"
      :mask-closable="false"
    >
      <n-form label-placement="left" label-width="100px">
        <n-form-item label="已绑设备">
          <span>{{ maxDevicesBound }}</span>
        </n-form-item>
        <n-form-item label="绑定上限">
          <n-input-number v-model:value="maxDevicesValue" :min="1" :precision="0" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showMaxDevicesModal = false">取消</n-button>
          <n-button type="primary" :loading="maxDevicesLoading" @click="handleMaxDevices">确认修改</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showDevicesModal"
      preset="card"
      :title="devicesModalTitle"
      :style="{ width: '860px' }"
    >
      <n-data-table
        :columns="deviceColumns"
        :data="deviceList"
        :loading="deviceLoading"
        :bordered="false"
        :single-line="false"
        :row-key="(row: GptRtLicenseDevice) => row.id"
        :scroll-x="800"
      />
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import {
  NCard, NSpace, NButton, NInput, NSelect, NDataTable, NModal, NForm, NFormItem,
  NTag, NAlert, NInputNumber, useMessage, useDialog, type PaginationProps, type FormInst, type DataTableColumns
} from 'naive-ui'
import {
  getGptRtLicenseList, createGptRtLicense, updateGptRtLicense, deleteGptRtLicense,
  getGptRtLicenseDevices,
  type GptRtLicense, type GptRtLicenseDevice
} from '@/api/gpt-rt-license'

const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const list = ref<GptRtLicense[]>([])
const searchStatus = ref(0)
const searchKeyword = ref('')

const statusOptions = [
  { label: '全部', value: 0 },
  { label: '正常', value: 1 },
  { label: '禁用', value: 2 },
]

const monthOptions = [
  { label: '1 个月', value: 1 },
  { label: '3 个月', value: 3 },
  { label: '6 个月', value: 6 },
  { label: '12 个月', value: 12 },
]

const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

function licenseStatus(row: GptRtLicense): { label: string; type: 'success' | 'warning' | 'error' | 'default' } {
  if (row.status !== 1) return { label: '禁用', type: 'error' }
  if (new Date(row.expires_at) < new Date()) return { label: '已过期', type: 'warning' }
  return { label: '正常', type: 'success' }
}

function fmtTime(s?: string) {
  if (!s) return '—'
  try {
    return new Date(s).toLocaleString('zh-CN')
  } catch {
    return s
  }
}

const showDevicesModal = ref(false)
const deviceLoading = ref(false)
const deviceList = ref<GptRtLicenseDevice[]>([])
const devicesTitleLicense = ref('')
const devicesModalTitle = computed(() =>
  devicesTitleLicense.value ? `已绑定设备 - ${devicesTitleLicense.value}` : '已绑定设备'
)

const deviceColumns: DataTableColumns<GptRtLicenseDevice> = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '机器 ID',
    key: 'machine_id',
    width: 200,
    ellipsis: { tooltip: true },
  },
  { title: '登录 IP', key: 'login_ip', width: 140 },
  {
    title: 'User-Agent',
    key: 'user_agent',
    width: 220,
    ellipsis: { tooltip: true },
  },
  {
    title: '绑定时间',
    key: 'bound_at',
    width: 170,
    render: (row) => fmtTime(row.bound_at),
  },
]

async function openDevices(row: GptRtLicense) {
  devicesTitleLicense.value = row.license_key
  showDevicesModal.value = true
  deviceLoading.value = true
  deviceList.value = []
  try {
    const res = await getGptRtLicenseDevices(row.id)
    if (res.code === 200 && res.data) {
      deviceList.value = res.data.list || []
    } else {
      message.error(res.message || '加载设备失败')
    }
  } catch (e: any) {
    message.error(e?.message || '加载设备失败')
  } finally {
    deviceLoading.value = false
  }
}

const columns: DataTableColumns<GptRtLicense> = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '激活码',
    key: 'license_key',
    width: 180,
    render: (row) => {
      const s = licenseStatus(row)
      return h(
        'div',
        { style: 'display:flex;flex-direction:column;gap:6px;align-items:flex-start' },
        [
          h('code', { style: 'font-size:12px;word-break:break-all' }, row.license_key),
          h(NTag, { type: s.type, size: 'small' }, () => s.label),
        ]
      )
    },
  },
  { title: '客户', key: 'customer', width: 120, ellipsis: { tooltip: true } },
  { title: 'AppId', key: 'app_id', width: 140 },
  {
    title: '到期时间',
    key: 'expires_at',
    width: 180,
    render: (row) => fmtTime(row.expires_at),
  },
  {
    title: '已用/可用',
    key: 'used_count',
    width: 110,
    render: (row) => `${row.used_count ?? 0} / ${row.available_count ?? 0}`,
  },
  {
    title: '设备',
    key: 'bound_device_count',
    width: 75,
    render: (row) => {
      const bound = row.bound_device_count ?? 0
      const max = row.max_devices ?? 1
      return h(
        NButton,
        {
          text: true,
          type: 'primary',
          onClick: () => openDevices(row),
        },
        () => `${bound} / ${max}`
      )
    },
  },
  {
    title: 'IP',
    key: 'last_using_ip',
    width: 132,
    render: (row) => row.last_using_ip || '—',
  },
  {
    title: '最后验证',
    key: 'last_verified_at',
    width: 180,
    render: (row) => fmtTime(row.last_verified_at),
  },
  {
    title: '操作',
    key: 'actions',
    width: 380,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          {
            size: 'small',
            type: row.status === 1 ? 'warning' : 'success',
            onClick: () => toggleStatus(row),
          },
          () => (row.status === 1 ? '禁用' : '启用')
        ),
        h(NButton, { size: 'small', type: 'info', onClick: () => openAddCount(row) }, () => '加号'),
        h(NButton, { size: 'small', onClick: () => openMaxDevices(row) }, () => '绑定'),
        h(NButton, { size: 'small', type: 'primary', onClick: () => openRenew(row) }, () => '续期'),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
      ]),
  },
]

async function loadList() {
  loading.value = true
  try {
    const res = await getGptRtLicenseList({
      page: pagination.value.page as number,
      page_size: pagination.value.pageSize as number,
      status: searchStatus.value,
      keyword: searchKeyword.value.trim(),
    })
    if (res.code === 200 && res.data) {
      list.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '加载失败')
    }
  } catch (e: any) {
    message.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.value.page = 1
  loadList()
}

function handleReset() {
  searchStatus.value = 0
  searchKeyword.value = ''
  pagination.value.page = 1
  loadList()
}

function handlePageChange(page: number) {
  pagination.value.page = page
  loadList()
}

function handlePageSizeChange(size: number) {
  pagination.value.pageSize = size
  pagination.value.page = 1
  loadList()
}

const showCreateModal = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInst | null>(null)
const createForm = ref({ customer: '', app_id: '', months: 1, available_count: 500, max_devices: 1 })

const showCreatedKeyModal = ref(false)
const createdLicenseKey = ref('')

async function handleCreate() {
  try {
    await createFormRef.value?.validate()
  } catch {
    return
  }
  createLoading.value = true
  try {
    const res = await createGptRtLicense({
      customer: createForm.value.customer.trim(),
      app_id: createForm.value.app_id.trim() || undefined,
      months: createForm.value.months,
      available_count: createForm.value.available_count,
      max_devices: createForm.value.max_devices,
    })
    if (res.code === 200 && res.data) {
      createdLicenseKey.value = res.data.license_key
      showCreateModal.value = false
      showCreatedKeyModal.value = true
      createForm.value = { customer: '', app_id: '', months: 1, available_count: 500, max_devices: 1 }
      loadList()
      message.success('颁发成功')
    } else {
      message.error(res.message || '颁发失败')
    }
  } catch (e: any) {
    message.error(e?.message || '颁发失败')
  } finally {
    createLoading.value = false
  }
}

function copyCreatedKey() {
  navigator.clipboard.writeText(createdLicenseKey.value).then(
    () => message.success('已复制'),
    () => message.warning('复制失败，请手动复制')
  )
}

async function toggleStatus(row: GptRtLicense) {
  const newStatus = row.status === 1 ? 0 : 1
  try {
    const res = await updateGptRtLicense(row.id, { status: newStatus })
    if (res.code === 200) {
      message.success(newStatus === 1 ? '已启用' : '已禁用')
      loadList()
    } else {
      message.error(res.message || '操作失败')
    }
  } catch (e: any) {
    message.error(e?.message || '操作失败')
  }
}

const showRenewModal = ref(false)
const renewLoading = ref(false)
const renewMonths = ref(1)
const renewTargetId = ref(0)

const showAddCountModal = ref(false)
const addCountLoading = ref(false)
const addCountValue = ref(500)
const addCountCurrent = ref(0)
const addCountTargetId = ref(0)

function openAddCount(row: GptRtLicense) {
  addCountTargetId.value = row.id
  addCountCurrent.value = row.available_count ?? 0
  addCountValue.value = 500
  showAddCountModal.value = true
}

async function handleAddCount() {
  if (!addCountValue.value || addCountValue.value < 1) {
    message.warning('增加数量必须为正整数')
    return
  }
  addCountLoading.value = true
  try {
    const res = await updateGptRtLicense(addCountTargetId.value, {
      add_available_count: addCountValue.value,
    })
    if (res.code === 200) {
      message.success('已增加可用数量')
      showAddCountModal.value = false
      loadList()
    } else {
      message.error(res.message || '操作失败')
    }
  } catch (e: any) {
    message.error(e?.message || '操作失败')
  } finally {
    addCountLoading.value = false
  }
}

const showMaxDevicesModal = ref(false)
const maxDevicesLoading = ref(false)
const maxDevicesValue = ref(1)
const maxDevicesBound = ref(0)
const maxDevicesTargetId = ref(0)

function openMaxDevices(row: GptRtLicense) {
  maxDevicesTargetId.value = row.id
  maxDevicesBound.value = row.bound_device_count ?? 0
  maxDevicesValue.value = row.max_devices ?? 1
  showMaxDevicesModal.value = true
}

async function handleMaxDevices() {
  if (!maxDevicesValue.value || maxDevicesValue.value < 1) {
    message.warning('绑定数量必须为正整数')
    return
  }
  maxDevicesLoading.value = true
  try {
    const res = await updateGptRtLicense(maxDevicesTargetId.value, {
      max_devices: maxDevicesValue.value,
    })
    if (res.code === 200) {
      message.success('已修改绑定数量')
      showMaxDevicesModal.value = false
      loadList()
    } else {
      message.error(res.message || '操作失败')
    }
  } catch (e: any) {
    message.error(e?.message || '操作失败')
  } finally {
    maxDevicesLoading.value = false
  }
}

function openRenew(row: GptRtLicense) {
  renewTargetId.value = row.id
  renewMonths.value = 1
  showRenewModal.value = true
}

async function handleRenew() {
  renewLoading.value = true
  try {
    const res = await updateGptRtLicense(renewTargetId.value, {
      months: renewMonths.value,
      available_count: 500,
    })
    if (res.code === 200) {
      message.success('续期成功')
      showRenewModal.value = false
      loadList()
    } else {
      message.error(res.message || '续期失败')
    }
  } catch (e: any) {
    message.error(e?.message || '续期失败')
  } finally {
    renewLoading.value = false
  }
}

function handleDelete(row: GptRtLicense) {
  dialog.warning({
    title: '确认删除',
    content: `确定删除激活码 ${row.license_key}？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await deleteGptRtLicense(row.id)
        if (res.code === 200) {
          message.success('已删除')
          loadList()
        } else {
          message.error(res.message || '删除失败')
        }
      } catch (e: any) {
        message.error(e?.message || '删除失败')
      }
    },
  })
}

onMounted(() => loadList())
</script>
