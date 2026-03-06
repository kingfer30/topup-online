<template>
  <div class="sales-talks-management">
    <n-card title="话术管理">
      <template #header-extra>
        <n-space>
          <!-- 批量设置标签（有选中行时显示） -->
          <n-button
            v-if="checkedRowKeys.length > 0"
            type="warning"
            @click="showBatchTagModal = true"
          >
            🏷️ 批量设置标签（已选 {{ checkedRowKeys.length }} 条）
          </n-button>
          <n-button type="primary" @click="handleAdd">
            <template #icon><span>➕</span></template>
            新增话术
          </n-button>
        </n-space>
      </template>

      <n-space vertical :size="16">
        <!-- 搜索栏 -->
        <n-space align="center" :wrap="false">
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索标题/内容"
            clearable
            style="width: 260px"
          />
          <n-select
            v-model:value="searchTag"
            :options="TAG_OPTIONS"
            filterable
            tag
            clearable
            placeholder="按标签筛选"
            style="width: 180px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <!-- 话术列表 -->
        <n-data-table
          remote
          :columns="columns"
          :data="list"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: SalesTalk) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>

    <!-- 新增/编辑对话框 -->
    <n-modal
      v-model:show="showModal"
      :title="isEdit ? '编辑话术' : '新增话术'"
      preset="dialog"
      :positive-text="isEdit ? '保存' : '创建'"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 760px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="100px"
        require-mark-placement="left"
        style="margin-top: 20px"
      >
        <n-form-item label="标题" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入话术标题" />
        </n-form-item>

        <n-form-item label="标签" path="tag">
          <n-select
            v-model:value="formData.tag"
            :options="TAG_OPTIONS"
            filterable
            tag
            clearable
            placeholder="选择或输入标签"
          />
        </n-form-item>

        <n-form-item label="排序" path="sort">
          <n-input-number
            v-model:value="formData.sort"
            :min="0"
            placeholder="数值越小越靠前"
            style="width: 100%"
          />
        </n-form-item>

        <n-form-item label="中文内容" path="zh_content">
          <n-input
            v-model:value="formData.zh_content"
            type="textarea"
            placeholder="请输入中文话术内容"
            :rows="5"
            show-count
          />
        </n-form-item>

        <n-form-item label="英文内容" path="en_content">
          <n-input
            v-model:value="formData.en_content"
            type="textarea"
            placeholder="请输入英文话术内容"
            :rows="5"
            show-count
          />
        </n-form-item>

        <n-form-item label="俄文内容" path="ru_content">
          <n-input
            v-model:value="formData.ru_content"
            type="textarea"
            placeholder="请输入俄文话术内容"
            :rows="5"
            show-count
          />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 批量设置标签对话框 -->
    <n-modal
      v-model:show="showBatchTagModal"
      title="批量设置标签"
      preset="dialog"
      positive-text="确认设置"
      negative-text="取消"
      @positive-click="handleBatchTag"
      style="width: 440px"
    >
      <div style="margin-top: 16px">
        <p style="margin-bottom: 12px; color: #666;">
          已选中 <strong>{{ checkedRowKeys.length }}</strong> 条记录，将统一设置标签为：
        </p>
        <n-select
          v-model:value="batchTag"
          :options="TAG_OPTIONS"
          filterable
          tag
          clearable
          placeholder="选择或输入标签"
        />
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  NCard,
  NButton,
  NSpace,
  NInput,
  NInputNumber,
  NSelect,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NDropdown,
  NTag,
  NPopconfirm,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type PaginationProps,
} from 'naive-ui'
import { http } from '@/utils/http'

// 标签预设选项
const TAG_OPTIONS = [
  { label: '通用', value: '通用' },
  { label: 'API', value: 'API' },
  { label: 'Cursor', value: 'Cursor' },
  { label: 'GPT', value: 'GPT' },
  { label: 'Suno', value: 'Suno' },
  { label: 'DeepSeek', value: 'DeepSeek' },
]

// 标签颜色映射
const TAG_TYPE_MAP: Record<string, 'default' | "primary" |  'info' | 'success' | 'warning' | 'error'> = {
  通用: 'default',
  API: 'info',
  Cursor: 'warning',
  GPT: 'success',
  Suno: 'primary',
  DeepSeek: 'warning',
}

// 类型定义
interface SalesTalk {
  id: number
  title: string
  tag: string
  sort: number
  zh_content: string
  en_content: string
  ru_content: string
  created_at: string | null
}

interface FormData {
  title: string
  tag: string | null
  sort: number
  zh_content: string
  en_content: string
  ru_content: string
}

// 状态
const message = useMessage()
const loading = ref(false)
const list = ref<SalesTalk[]>([])
const searchKeyword = ref('')
const searchTag = ref<string | null>(null)
const showModal = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInst | null>(null)
const currentEditId = ref<number>(0)
const checkedRowKeys = ref<number[]>([])
const showBatchTagModal = ref(false)
const batchTag = ref<string | null>(null)

const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: ({ itemCount }: { itemCount: number | undefined }) => `共 ${itemCount ?? 0} 条`,
})

const formData = ref<FormData>({
  title: '',
  tag: null,
  sort: 0,
  zh_content: '',
  en_content: '',
  ru_content: '',
})

const rules = {
  title: [{ required: true, message: '请输入话术标题', trigger: 'blur' }],
  zh_content: [{ required: true, message: '请输入中文内容', trigger: 'blur' }],
  en_content: [{ required: true, message: '请输入英文内容', trigger: 'blur' }],
  ru_content: [{ required: true, message: '请输入俄文内容', trigger: 'blur' }],
}

// 复制下拉选项
const COPY_OPTIONS = [
  { label: '🇨🇳 中文', key: 'zh' },
  { label: '🇺🇸 English', key: 'en' },
  { label: '🇷🇺 Русский', key: 'ru' },
]

// 表格列定义
const columns: DataTableColumns<SalesTalk> = [
  {
    type: 'selection',
  },
  {
    title: '标题',
    key: 'title',
    width: 220,
    ellipsis: { tooltip: true },
  },
  {
    title: '标签',
    key: 'tag',
    width: 90,
    render(row) {
      if (!row.tag) return h('span', { style: 'color:#ccc' }, '-')
      return h(NTag, { type: TAG_TYPE_MAP[row.tag] ?? 'default', size: 'small', bordered: false }, { default: () => row.tag })
    },
  },
  {
    title: '中文预览',
    key: 'zh_content',
    ellipsis: { tooltip: true },
    render(row) {
      const preview = row.zh_content?.length > 60 ? row.zh_content.slice(0, 60) + '…' : row.zh_content
      return h('span', { style: 'color:#666' }, preview || '-')
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right',
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NDropdown,
            {
              options: COPY_OPTIONS,
              onSelect: (key: string) => {
                const contentMap: Record<string, string> = {
                  zh: row.zh_content,
                  en: row.en_content,
                  ru: row.ru_content,
                }
                const langMap: Record<string, string> = { zh: '中文', en: '英文', ru: '俄文' }
                copyContent(contentMap[key] || '', langMap[key] || '')
              },
            },
            { default: () => h(NButton, { size: 'small', type: 'info' }, { default: () => '复制文案 ▾' }) }
          ),
          h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
          h(NPopconfirm, { onPositiveClick: () => handleDelete(row.id) }, {
            default: () => '确定要删除这条话术吗？',
            trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
          }),
        ],
      })
    },
  },
]

// 获取话术列表
const fetchList = async () => {
  loading.value = true
  try {
    const res = await http.get('/admin/sales-talks', {
      params: {
        page: pagination.value.page,
        page_size: pagination.value.pageSize,
        keyword: searchKeyword.value,
        tag: searchTag.value ?? '',
      },
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

// 搜索
const handleSearch = () => {
  pagination.value.page = 1
  checkedRowKeys.value = []
  fetchList()
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  searchTag.value = null
  pagination.value.page = 1
  checkedRowKeys.value = []
  fetchList()
}

// 分页切换
const handlePageChange = (page: number) => {
  pagination.value.page = page
  checkedRowKeys.value = []
  fetchList()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  checkedRowKeys.value = []
  fetchList()
}

// 重置表单
const resetForm = () => {
  formData.value = { title: '', tag: null, sort: 0, zh_content: '', en_content: '', ru_content: '' }
  currentEditId.value = 0
}

// 新增
const handleAdd = () => {
  isEdit.value = false
  resetForm()
  showModal.value = true
}

// 编辑
const handleEdit = (row: SalesTalk) => {
  isEdit.value = true
  currentEditId.value = row.id
  formData.value = {
    title: row.title,
    tag: row.tag || null,
    sort: row.sort,
    zh_content: row.zh_content,
    en_content: row.en_content,
    ru_content: row.ru_content,
  }
  showModal.value = true
}

// 复制内容
const copyContent = async (content: string, lang: string) => {
  try {
    await navigator.clipboard.writeText(content)
    message.success(`${lang}话术已复制到剪贴板`)
  } catch {
    message.error('复制失败，请手动选中文本复制')
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return false
  }

  try {
    let res
    const payload = { ...formData.value, tag: formData.value.tag ?? '' }
    if (isEdit.value) {
      res = await http.put(`/admin/sales-talks/${currentEditId.value}`, payload)
    } else {
      res = await http.post('/admin/sales-talks', payload)
    }

    if (res.code === 200) {
      message.success(isEdit.value ? '更新成功' : '创建成功')
      showModal.value = false
      fetchList()
    } else {
      message.error(res.message || '操作失败')
      return false
    }
  } catch (e: any) {
    message.error(e?.message || '操作失败')
    return false
  }
}

// 删除
const handleDelete = async (id: number) => {
  try {
    const res = await http.delete(`/admin/sales-talks/${id}`)
    if (res.code === 200) {
      message.success('删除成功')
      checkedRowKeys.value = checkedRowKeys.value.filter(k => k !== id)
      fetchList()
    } else {
      message.error(res.message || '删除失败')
    }
  } catch (e: any) {
    message.error(e?.message || '删除失败')
  }
}

// 批量设置标签
const handleBatchTag = async () => {
  if (batchTag.value === null) {
    message.warning('请选择或输入标签')
    return false
  }
  try {
    const res = await http.post('/admin/sales-talks/batch-tag', {
      ids: checkedRowKeys.value,
      tag: batchTag.value,
    })
    if (res.code === 200) {
      message.success('批量设置成功')
      showBatchTagModal.value = false
      batchTag.value = null
      checkedRowKeys.value = []
      fetchList()
    } else {
      message.error(res.message || '批量设置失败')
      return false
    }
  } catch (e: any) {
    message.error(e?.message || '批量设置失败')
    return false
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.talk-content-box {
  padding: 8px 0;
}
</style>
