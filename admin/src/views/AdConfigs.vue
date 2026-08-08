<template>
  <div class="ad-configs">
    <n-card title="广告配置">
      <template #header-extra>
        <n-button type="primary" @click="handleAdd">
          <template #icon><span>➕</span></template>
          新增广告
        </n-button>
      </template>

      <n-space vertical :size="16">
        <n-space align="center">
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索标题/购买人/链接"
            clearable
            style="width: 260px"
          />
          <n-select
            v-model:value="searchPosition"
            :options="positionOptions"
            clearable
            placeholder="广告位置"
            style="width: 140px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <n-alert type="info" :bordered="false">
          左右侧广告图片固定 160×100；「顶部」为全宽走马灯通知，只展示最新一条生效记录的标题文案，无需图片。
        </n-alert>

        <n-data-table
          remote
          :columns="columns"
          :data="list"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: AdConfig) => row.id"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
          :scroll-x="1400"
        />
      </n-space>
    </n-card>

    <n-modal
      v-model:show="showModal"
      :title="isEdit ? '编辑广告' : '新增广告'"
      preset="card"
      :style="{ width: '640px' }"
      :mask-closable="false"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="110px"
        style="margin-top: 8px"
      >
        <n-form-item label="标题" path="title">
          <n-input
            v-model:value="formData.title"
            type="textarea"
            :rows="formData.position === 'top' ? 3 : 1"
            :placeholder="formData.position === 'top' ? '顶部走马灯通知文案' : '广告标题（显示在图片下方）'"
          />
        </n-form-item>
        <n-form-item label="位置" path="position">
          <n-select v-model:value="formData.position" :options="positionOptions" />
        </n-form-item>
        <n-form-item label="购买人" path="buyer">
          <n-input v-model:value="formData.buyer" placeholder="广告位购买人" />
        </n-form-item>
        <n-form-item label="跳转链接" path="link">
          <n-input
            v-model:value="formData.link"
            :placeholder="formData.position === 'top' ? '可选，点击通知时跳转' : 'https://...'"
          />
        </n-form-item>
        <n-form-item v-if="formData.position !== 'top'" label="广告图片" path="image">
          <n-space vertical>
            <n-upload
              :key="uploadInputKey"
              v-model:file-list="uploadFileList"
              :max="1"
              accept="image/png,image/jpeg,image/gif,image/jpg"
              :default-upload="false"
              :show-file-list="false"
              @change="handleUploadChange"
            >
              <n-button :loading="uploading">
                {{ formData.image ? '重新上传 (160×100)' : '上传图片 (160×100)' }}
              </n-button>
            </n-upload>
            <div v-if="formData.image" class="ad-preview">
              <img :src="imagePreviewUrl" alt="preview" />
              <div class="ad-preview-title">{{ formData.title || '标题预览' }}</div>
            </div>
          </n-space>
        </n-form-item>
        <n-form-item label="生效时间" path="start_time">
          <n-date-picker
            v-model:value="formData.start_time"
            type="datetime"
            clearable
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="失效时间" path="end_time">
          <n-date-picker
            v-model:value="formData.end_time"
            type="datetime"
            clearable
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="排序" path="sort">
          <n-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
        </n-form-item>
        <n-form-item label="状态" path="status">
          <n-switch v-model:value="formData.status" :checked-value="1" :unchecked-value="0">
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitLoading" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  NUpload,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
  type PaginationProps,
  type UploadFileInfo,
} from 'naive-ui'
import {
  createAdConfig,
  deleteAdConfig,
  getAdConfigList,
  updateAdConfig,
  uploadAdImage,
  type AdConfig,
} from '@/api/ad-config'

interface FormData {
  title: string
  image: string
  link: string
  position: string
  buyer: string
  sort: number
  status: number
  start_time: number | null
  end_time: number | null
}

const message = useMessage()
const loading = ref(false)
const submitLoading = ref(false)
const uploading = ref(false)
const uploadFileList = ref<UploadFileInfo[]>([])
const uploadInputKey = ref(0)
const list = ref<AdConfig[]>([])
const searchKeyword = ref('')
const searchPosition = ref<string | null>(null)
const showModal = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInst | null>(null)
const currentEditId = ref(0)

const positionOptions = [
  { label: '左侧', value: 'left' },
  { label: '右侧', value: 'right' },
  { label: '顶部', value: 'top' },
]

const isTopPosition = computed(() => formData.value.position === 'top')

const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  prefix: ({ itemCount }: { itemCount?: number }) => `共 ${itemCount ?? 0} 条`,
})

const formData = ref<FormData>({
  title: '',
  image: '',
  link: '',
  position: 'left',
  buyer: '',
  sort: 0,
  status: 1,
  start_time: null,
  end_time: null,
})

const rules = computed<FormRules>(() => ({
  title: [{ required: true, message: isTopPosition.value ? '请输入通知文案' : '请输入标题', trigger: 'blur' }],
  position: [{ required: true, message: '请选择位置', trigger: 'change' }],
  link: isTopPosition.value
    ? []
    : [{ required: true, message: '请输入跳转链接', trigger: 'blur' }],
  image: isTopPosition.value
    ? []
    : [{ required: true, message: '请上传 160×100 图片', trigger: 'change' }],
  start_time: [
    {
      required: true,
      type: 'number',
      message: '请选择生效时间',
      trigger: 'change',
      validator: (_rule, value) => {
        if (value === null || value === undefined) return new Error('请选择生效时间')
        return true
      },
    },
  ],
  end_time: [
    {
      required: true,
      type: 'number',
      message: '请选择失效时间',
      trigger: 'change',
      validator: (_rule, value) => {
        if (value === null || value === undefined) return new Error('请选择失效时间')
        return true
      },
    },
  ],
}))

const resolveAssetUrl = (img?: string) => {
  if (!img) return ''
  if (img.startsWith('http://') || img.startsWith('https://')) return img
  if (import.meta.env.DEV) return img
  const base = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/api\/?$/, '')
  return base + img
}

const imagePreviewUrl = computed(() => resolveAssetUrl(formData.value.image))

const formatTime = (v?: string) => {
  if (!v) return '-'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const columns: DataTableColumns<AdConfig> = [
  {
    title: '预览',
    key: 'image',
    width: 90,
    render(row) {
      if (row.position === 'top' || !row.image) {
        return h('span', { style: 'color:#999;font-size:12px' }, '走马灯')
      }
      return h('img', {
        src: resolveAssetUrl(row.image),
        style: 'width:80px;height:50px;object-fit:cover;border-radius:6px;background:#111',
      })
    },
  },
  { title: '标题', key: 'title', width: 140, ellipsis: { tooltip: true } },
  {
    title: '位置',
    key: 'position',
    width: 80,
    render(row) {
      const map: Record<string, { label: string; type: 'info' | 'warning' | 'success' }> = {
        left: { label: '左侧', type: 'info' },
        right: { label: '右侧', type: 'warning' },
        top: { label: '顶部', type: 'success' },
      }
      const item = map[row.position] || { label: row.position, type: 'info' as const }
      return h(
        NTag,
        { size: 'small', type: item.type, bordered: false },
        { default: () => item.label }
      )
    },
  },
  { title: '购买人', key: 'buyer', width: 120, ellipsis: { tooltip: true }, render: (row) => row.buyer || '-' },
  { title: '点击', key: 'click_count', width: 80 },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      return h(
        NTag,
        { size: 'small', type: row.status === 1 ? 'success' : 'default', bordered: false },
        { default: () => (row.status === 1 ? '启用' : '禁用') }
      )
    },
  },
  {
    title: '生效时间',
    key: 'start_time',
    width: 160,
    render: (row) => formatTime(row.start_time),
  },
  {
    title: '失效时间',
    key: 'end_time',
    width: 160,
    render: (row) => formatTime(row.end_time),
  },
  { title: '排序', key: 'sort', width: 70 },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right',
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row.id) },
            {
              default: () => '确定删除该广告吗？',
              trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
            }
          ),
        ],
      })
    },
  },
]

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getAdConfigList({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      keyword: searchKeyword.value,
      position: searchPosition.value || '',
    })
    if (res.code === 200) {
      list.value = res.data.list || []
      pagination.value.itemCount = res.data.total
    } else {
      message.error(res.message || '获取失败')
    }
  } catch (e: any) {
    message.error(e?.message || '获取失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  fetchList()
}

const handleReset = () => {
  searchKeyword.value = ''
  searchPosition.value = null
  pagination.value.page = 1
  fetchList()
}

const handlePageChange = (page: number) => {
  pagination.value.page = page
  fetchList()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  fetchList()
}

const resetUploadInput = () => {
  uploadFileList.value = []
  uploadInputKey.value += 1
}

const resetForm = () => {
  formData.value = {
    title: '',
    image: '',
    link: '',
    position: 'left',
    buyer: '',
    sort: 0,
    status: 1,
    start_time: Date.now(),
    end_time: Date.now() + 30 * 24 * 3600 * 1000,
  }
  currentEditId.value = 0
  resetUploadInput()
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  showModal.value = true
}

const handleEdit = (row: AdConfig) => {
  isEdit.value = true
  currentEditId.value = row.id
  formData.value = {
    title: row.title,
    image: row.image,
    link: row.link,
    position: row.position,
    buyer: row.buyer || '',
    sort: row.sort || 0,
    status: row.status ?? 1,
    start_time: row.start_time ? new Date(row.start_time).getTime() : null,
    end_time: row.end_time ? new Date(row.end_time).getTime() : null,
  }
  resetUploadInput()
  showModal.value = true
}

const handleUploadChange = async (options: { file: UploadFileInfo; fileList: UploadFileInfo[] }) => {
  const fileInfo = options.file
  const raw = fileInfo.file
  if (!raw || fileInfo.status === 'removed') return
  if (uploading.value) return
  uploading.value = true
  try {
    const res = await uploadAdImage(raw)
    if (res.code === 200 && res.data?.url) {
      formData.value.image = res.data.url
      message.success('上传成功')
    } else {
      message.error(res.message || '上传失败')
    }
  } catch (e: any) {
    message.error(e?.message || '上传失败')
  } finally {
    uploading.value = false
    // 清空上传列表并重建组件，允许同一弹窗内再次选择图片
    resetUploadInput()
  }
}

const toISOString = (ms: number | null) => {
  if (ms === null || ms === undefined) return ''
  return new Date(ms).toISOString()
}

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }
  if (formData.value.position !== 'top' && !formData.value.image) {
    message.error('请上传 160×100 图片')
    return
  }
  if (
    formData.value.start_time !== null &&
    formData.value.end_time !== null &&
    formData.value.end_time < formData.value.start_time
  ) {
    message.error('失效时间不能早于生效时间')
    return
  }

  submitLoading.value = true
  const payload = {
    title: formData.value.title,
    image: formData.value.position === 'top' ? '' : formData.value.image,
    link: formData.value.link,
    position: formData.value.position,
    buyer: formData.value.buyer,
    sort: formData.value.sort,
    status: formData.value.status,
    start_time: toISOString(formData.value.start_time),
    end_time: toISOString(formData.value.end_time),
  }
  try {
    const res = isEdit.value
      ? await updateAdConfig(currentEditId.value, payload)
      : await createAdConfig(payload)
    if (res.code === 200) {
      message.success(isEdit.value ? '保存成功' : '创建成功')
      showModal.value = false
      fetchList()
    } else {
      message.error(res.message || '操作失败')
    }
  } catch (e: any) {
    message.error(e?.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleDelete = async (id: number) => {
  try {
    const res = await deleteAdConfig(id)
    if (res.code === 200) {
      message.success('删除成功')
      fetchList()
    } else {
      message.error(res.message || '删除失败')
    }
  } catch (e: any) {
    message.error(e?.message || '删除失败')
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.ad-preview {
  width: 160px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  background: #0a0e1a;
}
.ad-preview img {
  display: block;
  width: 160px;
  height: 100px;
  object-fit: cover;
}
.ad-preview-title {
  padding: 8px 10px;
  font-size: 13px;
  color: #e8ecf4;
  text-align: center;
  word-break: break-all;
}
</style>
