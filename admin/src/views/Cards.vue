<template>
  <div class="cards-management">
    <n-card :title="pageTitle">
      <template #header-extra>
        <n-space>
          <n-button v-if="cardType === 'unsold'" type="success" @click="handlePickup">
            <template #icon>
              <span>📦</span>
            </template>
            我要取货
          </n-button>
          <n-button type="primary" @click="handleAdd">
            <template #icon>
              <span>➕</span>
            </template>
            新增卡密
          </n-button>
          <n-button type="info" @click="handleBatchImport">
            <template #icon>
              <span>📥</span>
            </template>
            批量导入
          </n-button>
        </n-space>
      </template>

      <!-- 搜索栏 -->
      <n-space vertical :size="16">
        <n-space>
          <n-input
            v-model:value="searchKeyword"
            placeholder="搜索账号/邮箱"
            clearable
            style="width: 300px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <!-- 卡密表格 -->
        <n-data-table
          :columns="columns"
          :data="cardList"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: Card) => row.id"
          @update:page="handlePageChange"
        />
      </n-space>
    </n-card>

    <!-- 新增/编辑对话框 -->
    <n-modal
      v-model:show="showModal"
      :title="isEdit ? '编辑卡密' : '新增卡密'"
      preset="dialog"
      :positive-text="isEdit ? '保存' : '创建'"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 800px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="140px"
        require-mark-placement="left"
        style="margin-top: 20px"
      >
        <n-grid :cols="2" :x-gap="24">
          <n-form-item-gi label="账号" path="account">
            <n-input v-model:value="formData.account" placeholder="请输入账号" />
          </n-form-item-gi>

          <n-form-item-gi label="密码" path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              placeholder="请输入密码"
            />
          </n-form-item-gi>

          <n-form-item-gi label="邮箱密码" path="mail_password">
            <n-input
              v-model:value="formData.mail_password"
              type="password"
              placeholder="请输入邮箱密码"
            />
          </n-form-item-gi>

          <n-form-item-gi label="邮箱地址" path="mail_url">
            <n-input v-model:value="formData.mail_url" placeholder="请输入邮箱地址" />
          </n-form-item-gi>

          <n-form-item-gi label="订阅类型" path="subscription_type">
            <n-select
              v-model:value="formData.subscription_type"
              :options="subscriptionTypeOptions"
              placeholder="选择或输入订阅类型"
              filterable
              tag
              clearable
            />
          </n-form-item-gi>

          <n-form-item-gi label="订阅状态" path="subscription_status">
            <n-select
              v-model:value="formData.subscription_status"
              :options="subscriptionStatusOptions"
            />
          </n-form-item-gi>

          <n-form-item-gi label="购买价格" path="purchase_price">
            <n-input-number
              v-model:value="formData.purchase_price"
              :min="0"
              :precision="2"
              placeholder="购买价格"
              style="width: 100%"
            />
          </n-form-item-gi>

          <n-form-item-gi label="购买平台" path="purchase_from">
            <n-select
              v-model:value="formData.purchase_from"
              :options="purchasePlatformOptions"
              placeholder="选择或输入购买平台"
              filterable
              tag
              clearable
            />
          </n-form-item-gi>

          <n-form-item-gi label="出售价格" path="sell_price">
            <n-input-number
              v-model:value="formData.sell_price"
              :min="0"
              :precision="2"
              placeholder="出售价格"
              style="width: 100%"
            />
          </n-form-item-gi>

          <n-form-item-gi label="出售状态" path="sell_status">
            <n-select v-model:value="formData.sell_status" :options="sellStatusOptions" />
          </n-form-item-gi>

          <n-form-item-gi label="账号类型" path="account_type">
            <n-select
              v-model:value="formData.account_type"
              :options="accountTypeOptions"
            />
          </n-form-item-gi>

          <n-form-item-gi label="状态" path="status">
            <n-select v-model:value="formData.status" :options="statusOptions" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="API Key" path="api_key">
            <n-input v-model:value="formData.api_key" placeholder="请输入API Key" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="2FA" path="2fa">
            <n-input v-model:value="formData['2fa']" placeholder="请输入2FA" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="Token" path="token">
            <n-input
              v-model:value="formData.token"
              type="textarea"
              placeholder="请输入Token"
              :rows="3"
            />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="备注" path="remark">
            <n-input
              v-model:value="formData.remark"
              type="textarea"
              placeholder="请输入备注"
              :rows="2"
            />
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-modal>

    <!-- 批量导入对话框 -->
    <n-modal
      v-model:show="showBatchModal"
      title="批量导入卡密"
      preset="dialog"
      positive-text="导入"
      negative-text="取消"
      @positive-click="handleBatchSubmit"
      style="width: 1200px; max-width: 95vw"
    >
      <n-space vertical style="margin-top: 20px" :size="16">
        <!-- 公共字段配置 -->
        <n-card title="公共字段配置" size="small">
          <n-grid :cols="4" :x-gap="12" :y-gap="12">
            <n-form-item-gi label="订阅类型">
              <n-select
                v-model:value="batchConfig.subscription_type"
                :options="subscriptionTypeOptions"
                placeholder="选择或输入订阅类型"
                filterable
                tag
                clearable
              />
            </n-form-item-gi>

            <n-form-item-gi label="订阅剩余天数">
              <n-input-number
                v-model:value="batchConfig.subscription_remaining_days"
                :min="0"
                placeholder="默认30天"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="购买价格">
              <n-input-number
                v-model:value="batchConfig.purchase_price"
                :min="0"
                :precision="2"
                placeholder="购买价格"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="购买平台">
              <n-select
                v-model:value="batchConfig.purchase_from"
                :options="purchasePlatformOptions"
                placeholder="选择或输入购买平台"
                filterable
                tag
                clearable
              />
            </n-form-item-gi>

            <n-form-item-gi label="购买订单号">
              <n-input v-model:value="batchConfig.purchase_order_no" placeholder="购买订单号" />
            </n-form-item-gi>

            <n-form-item-gi label="购买时间">
              <n-date-picker
                v-model:value="batchConfig.purchase_date"
                type="datetime"
                placeholder="默认为当前时间"
                style="width: 100%"
                clearable
              />
            </n-form-item-gi>

            <n-form-item-gi label="邮箱地址" :span="2">
              <n-select
                v-model:value="batchConfig.mail_url"
                :options="mailUrlOptions"
                placeholder="选择或输入邮箱地址"
                filterable
                tag
                clearable
              />
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <!-- 字段映射配置 -->
        <n-card title="字段映射配置" size="small">
          <n-alert type="info" style="margin-bottom: 12px">
            导入数据使用分隔符"----"分割，请配置各字段的顺序位置（从1开始，0表示不导入该字段）
          </n-alert>
          <n-grid :cols="8" :x-gap="12" :y-gap="12">
            <n-form-item-gi label="账号">
              <n-input-number
                v-model:value="batchConfig.field_mapping.account"
                :min="1"
                placeholder="必填"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="密码">
              <n-input-number
                v-model:value="batchConfig.field_mapping.password"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="邮箱密码">
              <n-input-number
                v-model:value="batchConfig.field_mapping.mail_password"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="订阅时间">
              <n-input-number
                v-model:value="batchConfig.field_mapping.subscription_time"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="Token">
              <n-input-number
                v-model:value="batchConfig.field_mapping.token"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="API Key">
              <n-input-number
                v-model:value="batchConfig.field_mapping.api_key"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="2FA">
              <n-input-number
                v-model:value="batchConfig.field_mapping['2fa']"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>

            <n-form-item-gi label="备注">
              <n-input-number
                v-model:value="batchConfig.field_mapping.remark"
                :min="0"
                placeholder="0=不导入"
                style="width: 100%"
              />
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <!-- 导入数据输入 -->
        <n-alert type="warning">
          每行一条数据，字段之间使用"----"分隔。根据上方字段映射配置，按顺序填写对应字段数据。
        </n-alert>
        <n-input
          v-model:value="batchImportText"
          type="textarea"
          placeholder="示例（假设配置：账号=1，密码=2，邮箱密码=3）：&#10;account1@example.com----password1----mailpass1&#10;account2@example.com----password2----mailpass2&#10;&#10;包含Token示例（假设配置：账号=1，密码=2，Token=3）：&#10;account3@example.com----password3----token123"
          :rows="8"
        />
      </n-space>
    </n-modal>

    <!-- 取货对话框 -->
    <n-modal
      v-model:show="showPickupModal"
      :title="pickupStep === 1 ? '我要取货 - 选择条件' : '我要取货 - 预览确认'"
      preset="dialog"
      :positive-text="pickupStep === 1 ? '下一步' : '完成取货'"
      negative-text="取消"
      @positive-click="handlePickupSubmit"
      style="width: 700px"
    >
      <n-space vertical style="margin-top: 20px" :size="16">
        <!-- 第一步：选择条件 -->
        <div v-if="pickupStep === 1">
          <n-form
            :model="pickupForm"
            label-placement="left"
            label-width="100px"
          >
            <n-form-item label="订阅类型" path="subscription_type">
              <n-select
                v-model:value="pickupForm.subscription_type"
                :options="unsoldSubscriptionTypes"
                placeholder="请选择订阅类型"
              />
            </n-form-item>

            <n-form-item label="取货格式" path="format">
              <n-radio-group v-model:value="pickupForm.format">
                <n-space>
                  <n-radio value="digiseller">Digiseller订阅</n-radio>
                  <n-radio value="domestic">国内订阅</n-radio>
                </n-space>
              </n-radio-group>
            </n-form-item>
          </n-form>
        </div>

        <!-- 第二步：预览确认 -->
        <div v-if="pickupStep === 2">
          <n-alert type="success" style="margin-bottom: 16px">
            已为您选出一条卡密，请确认信息后完成取货
          </n-alert>

          <!-- 卡密信息预览 -->
          <n-card title="卡密信息" size="small" style="margin-bottom: 16px">
            <div style="position: relative">
              <n-button
                size="small"
                style="position: absolute; top: -40px; right: 0"
                @click="handleCopyPickupInfo"
              >
                复制
              </n-button>
              <pre 
                ref="pickupCardInfoRef"
                class="card-info-display"
                @click="handleSelectCardInfo"
              >{{ pickupCardInfo }}</pre>
            </div>
          </n-card>

          <!-- 售出信息 -->
          <n-form
            :model="completeForm"
            label-placement="left"
            label-width="100px"
          >
            <n-form-item label="售出价格">
              <n-input-number
                v-model:value="completeForm.sell_price"
                :min="0"
                :precision="2"
                placeholder="非必填"
                style="width: 100%"
              />
            </n-form-item>

            <n-form-item label="售出对方">
              <n-input
                v-model:value="completeForm.sell_to"
                placeholder="非必填"
              />
            </n-form-item>
          </n-form>
        </div>
      </n-space>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  NCard,
  NButton,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NFormItemGi,
  NInput,
  NInputNumber,
  NSelect,
  NSpace,
  NTag,
  NGrid,
  NAlert,
  NDatePicker,
  NRadio,
  NRadioGroup,
  useMessage,
  type DataTableColumns,
  type FormInst,
  type FormRules,
  type PaginationProps,
} from 'naive-ui'
import {
  getCardList,
  createCard,
  updateCard,
  deleteCard,
  batchImportCards,
  getUnsoldSubscriptionTypes,
  pickupCard,
  completePickup,
  type Card,
  type CardRequest,
} from '@/api/card'

const route = useRoute()
const message = useMessage()

// 从路由参数获取卡密类别和类型
const category = computed(() => {
  const cat = (route.query.category as string) || ''
  console.log('📦 Cards.vue - category:', cat)
  return cat
})
const cardType = computed(() => {
  const type = (route.query.type as string) || 'all'
  console.log('📦 Cards.vue - cardType:', type)
  return type
})

// 页面标题
const pageTitle = computed(() => {
  const typeMap: Record<string, string> = {
    all: '普号列表',
    unsold: '未售列表',
    sold: '已售列表',
  }
  const title = typeMap[cardType.value] || '卡密列表'
  console.log('📦 Cards.vue - pageTitle:', title)
  return title
})

// 取货卡密信息格式化
const pickupCardInfo = computed(() => {
  if (!pickedCard.value) return ''
  
  const card = pickedCard.value
  if (pickupForm.value.format === 'digiseller') {
    // digiseller订阅格式
    return `account: ${card.account}
pass: ${card.password || ''}
mail-pass: ${card.mail_password || ''}

mail-login: ${card.mail_url || ''}`
  } else {
    // 国内订阅格式
    return `账号----密码----邮箱密码|
${card.account}----${card.password || ''}----${card.mail_password || ''}`
  }
})

// 状态
const loading = ref(false)
const showModal = ref(false)
const showBatchModal = ref(false)
const isEdit = ref(false)
const cardList = ref<Card[]>([])
const formRef = ref<FormInst | null>(null)
const searchKeyword = ref('')
const batchImportText = ref('')

// 取货相关状态
const showPickupModal = ref(false)
const pickupStep = ref(1) // 1: 选择条件, 2: 预览确认
const unsoldSubscriptionTypes = ref<{ label: string; value: string }[]>([])
const pickedCard = ref<Card | null>(null)
const pickupCardInfoRef = ref<HTMLPreElement | null>(null)

// 取货表单
const pickupForm = ref({
  subscription_type: '',
  format: 'digiseller' as 'digiseller' | 'domestic',
})

// 完成取货表单
const completeForm = ref({
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 批量导入配置
const batchConfig = ref({
  subscription_type: 'pro',
  subscription_remaining_days: 30 as number | undefined,
  purchase_price: undefined as number | undefined,
  purchase_from: '微信',
  purchase_order_no: '',
  purchase_date: undefined as number | undefined,
  mail_url: 'https://login.live.com',
  field_mapping: {
    account: 1,
    password: 2,
    mail_password: 3,
    subscription_time: 0,
    token: 0,
    api_key: 0,
    '2fa': 0,
    remark: 0,
  },
})

// 分页
const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

// 表单数据
const formData = ref<CardRequest>({
  account: '',
  password: '',
  mail_password: '',
  subscription_status: 1,
  subscription_type: '',
  sell_status: 1,
  account_type: 1,
  status: 1,
  purchase_price: 0,
  sell_price: 0,
  api_key: '',
  '2fa': '',
  token: '',
  mail_url: '',
  remark: '',
})

// 当前编辑的卡密ID
const currentEditId = ref<number>(0)

// 表单验证规则
const rules: FormRules = {
  account: [{ required: true, message: '请输入账号', trigger: 'blur' }],
}

// 下拉选项
const subscriptionStatusOptions = [
  { label: '已订阅', value: 1 },
  { label: '未订阅', value: 2 },
]

const sellStatusOptions = [
  { label: '未出售', value: 1 },
  { label: '发货中', value: 2 },
  { label: '已出售', value: 3 },
]

const accountTypeOptions = [
  { label: '普号', value: 1 },
  { label: '成品', value: 2 },
]

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 2 },
]

// 订阅类型选项（支持手动输入）
const subscriptionTypeOptions = [
  { label: 'Pro', value: 'pro' },
  { label: 'Pro+', value: 'pro+' },
  { label: 'Ultra', value: 'ultra' },
  { label: 'Go', value: 'go' },
  { label: 'Plus', value: 'plus' },
  { label: 'Team', value: 'team' },
]

// 购买平台选项（支持手动输入）
const purchasePlatformOptions = [
  { label: '微信', value: '微信' },
  { label: 'Telegram', value: 'Telegram' },
  { label: '闲鱼', value: '闲鱼' },
  { label: '淘宝', value: '淘宝' },
]

// 邮箱地址选项（支持手动输入）
const mailUrlOptions = [
  { label: 'https://login.live.com', value: 'https://login.live.com' },
  { label: 'https://mail.com', value: 'https://mail.com' },
  { label: 'https://gmx.us', value: 'https://gmx.us' },
  { label: 'https://gmail.com', value: 'https://gmail.com' },
]

// 表格列定义
const columns: DataTableColumns<Card> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: '账号',
    key: 'account',
    width: 200,
  },
  {
    title: '订阅类型',
    key: 'subscription_type',
    width: 100,
  },
  {
    title: '订阅状态',
    key: 'subscription_status',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.subscription_status === 1 ? 'success' : 'warning',
          size: 'small',
        },
        { default: () => (row.subscription_status === 1 ? '已订阅' : '未订阅') }
      )
    },
  },
  {
    title: '购买价格',
    key: 'purchase_price',
    width: 100,
  },
  {
    title: '出售价格',
    key: 'sell_price',
    width: 100,
  },
  {
    title: '出售状态',
    key: 'sell_status',
    width: 100,
    render: (row) => {
      const typeMap: Record<number, { text: string; type: 'default' | 'info' | 'success' }> = {
        1: { text: '未出售', type: 'default' },
        2: { text: '发货中', type: 'info' },
        3: { text: '已出售', type: 'success' },
      }
      const config = typeMap[row.sell_status] || { text: '未知', type: 'default' }
      return h(NTag, { type: config.type, size: 'small' }, { default: () => config.text })
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.status === 1 ? 'success' : 'error',
          size: 'small',
        },
        { default: () => (row.status === 1 ? '正常' : '禁用') }
      )
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right',
    render: (row) => {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                type: 'primary',
                onClick: () => handleEdit(row),
              },
              { default: () => '编辑' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row),
              },
              { default: () => '删除' }
            ),
          ],
        }
      )
    },
  },
]

// 加载卡密列表
const loadCards = async () => {
  console.log('📡 loadCards 调用:')
  console.log('  category:', category.value)
  console.log('  cardType:', cardType.value)
  
  if (!category.value) {
    console.log('  ❌ 缺少卡密类别参数')
    message.error('缺少卡密类别参数')
    return
  }

  loading.value = true
  try {
    const params = {
      category: category.value,
      type: cardType.value,
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      keyword: searchKeyword.value,
    }
    console.log('  📤 请求参数:', params)
    
    const response = await getCardList(params)

    if (response.code === 200) {
      cardList.value = response.data.list || []
      pagination.value.itemCount = response.data.total
      console.log('  ✅ 加载成功，数据量:', cardList.value.length)
    } else {
      console.log('  ❌ 加载失败:', response.message)
      message.error(response.message || '加载失败')
    }
  } catch (error) {
    console.error('  ❌ 请求异常:', error)
    message.error('加载卡密列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.value.page = 1
  loadCards()
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  pagination.value.page = 1
  loadCards()
}

// 分页变化
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadCards()
}

// 新增卡密
const handleAdd = () => {
  isEdit.value = false
  currentEditId.value = 0
  formData.value = {
    account: '',
    password: '',
    mail_password: '',
    subscription_status: 1,
    subscription_type: '',
    sell_status: 1,
    account_type: 1,
    status: 1,
    purchase_price: 0,
    sell_price: 0,
    api_key: '',
    '2fa': '',
    token: '',
    mail_url: '',
    remark: '',
  }
  showModal.value = true
}

// 编辑卡密
const handleEdit = (card: Card) => {
  isEdit.value = true
  currentEditId.value = card.id
  formData.value = {
    account: card.account,
    password: card.password || '',
    mail_password: card.mail_password || '',
    subscription_status: card.subscription_status,
    subscription_type: card.subscription_type || '',
    sell_status: card.sell_status,
    account_type: card.account_type,
    status: card.status,
    purchase_price: card.purchase_price || 0,
    sell_price: card.sell_price || 0,
    purchase_from: card.purchase_from || '',
    api_key: card.api_key || '',
    '2fa': card['2fa'] || '',
    token: card.token || '',
    mail_url: card.mail_url || '',
    remark: card.remark || '',
  }
  showModal.value = true
}

// 删除卡密
const handleDelete = async (card: Card) => {
  const confirmed = await new Promise((resolve) => {
    const dialog = window.confirm(`确定要删除卡密"${card.account}"吗？`)
    resolve(dialog)
  })

  if (!confirmed) {
    return
  }

  try {
    const response = await deleteCard(category.value, card.id)
    if (response.code === 200) {
      message.success('删除成功')
      await loadCards()
    } else {
      message.error(response.message || '删除失败')
    }
  } catch (error: any) {
    console.error('删除卡密失败', error)
    message.error(error.response?.data?.message || '删除失败')
  }
}

// 提交表单
const handleSubmit = async () => {
  // 验证表单
  await formRef.value?.validate()

  try {
    if (isEdit.value) {
      // 更新卡密
      const response = await updateCard(category.value, currentEditId.value, formData.value)
      if (response.code === 200) {
        message.success('更新成功')
        showModal.value = false
        await loadCards()
      } else {
        message.error(response.message || '更新失败')
        return false
      }
    } else {
      // 创建卡密
      const response = await createCard(category.value, formData.value)
      if (response.code === 200) {
        message.success('创建成功')
        showModal.value = false
        await loadCards()
      } else {
        message.error(response.message || '创建失败')
        return false
      }
    }
  } catch (error: any) {
    console.error('提交表单失败', error)
    message.error(error.response?.data?.message || '操作失败')
    return false
  }
}

// 打开批量导入对话框
const handleBatchImport = () => {
  batchImportText.value = ''
  // 重置批量导入配置（保留字段映射）
  const currentMapping = { ...batchConfig.value.field_mapping }
  batchConfig.value = {
    subscription_type: 'pro',
    subscription_remaining_days: 30,
    purchase_price: undefined,
    purchase_from: '微信',
    purchase_order_no: '',
    purchase_date: undefined,
    mail_url: 'https://login.live.com',
    field_mapping: currentMapping,
  }
  showBatchModal.value = true
}

// 提交批量导入
const handleBatchSubmit = async () => {
  if (!batchImportText.value.trim()) {
    message.error('请输入要导入的数据')
    return false
  }

  try {
    // 解析导入数据
    const lines = batchImportText.value.trim().split('\n')
    const cards: CardRequest[] = []

    // 获取当前时间戳（秒）
    const currentTimestamp = Math.floor(Date.now() / 1000)
    
    // 计算购买时间
    const purchaseDate = batchConfig.value.purchase_date
      ? Math.floor(batchConfig.value.purchase_date / 1000)
      : currentTimestamp

    // 计算订阅过期时间（如果配置了订阅剩余天数）
    let subscriptionExpiredTime: number | undefined = undefined
    if (batchConfig.value.subscription_remaining_days) {
      subscriptionExpiredTime = currentTimestamp + batchConfig.value.subscription_remaining_days * 24 * 60 * 60
    }

    for (const line of lines) {
      const trimmedLine = line.trim()
      if (!trimmedLine) continue

      // 使用"----"分割
      const parts = trimmedLine.split('----').map((p) => p.trim())
      
      if (parts.length === 0) continue

      // 根据字段映射提取数据
      const mapping = batchConfig.value.field_mapping
      const account = mapping.account > 0 && parts[mapping.account - 1] ? parts[mapping.account - 1] : ''
      
      if (!account) {
        continue // 跳过没有账号的行
      }

      const password = mapping.password > 0 && parts[mapping.password - 1] ? parts[mapping.password - 1] : ''
      const mailPassword = mapping.mail_password > 0 && parts[mapping.mail_password - 1] ? parts[mapping.mail_password - 1] : ''
      
      // 订阅时间：如果配置了位置则从数据中读取，为0则不设置
      let subscriptionTime: number | undefined = undefined
      if (mapping.subscription_time > 0) {
        if (parts[mapping.subscription_time - 1]) {
          const timeValue = parts[mapping.subscription_time - 1]
          subscriptionTime = parseInt(timeValue) || currentTimestamp
        } else {
          subscriptionTime = currentTimestamp
        }
      }

      // Token：从数据中读取（如果配置了位置）
      const token = mapping.token > 0 && parts[mapping.token - 1] ? parts[mapping.token - 1] : ''

      // API Key：从数据中读取（如果配置了位置）
      const apiKey = mapping.api_key > 0 && parts[mapping.api_key - 1] ? parts[mapping.api_key - 1] : ''

      // 2FA：从数据中读取（如果配置了位置）
      const twoFA = mapping['2fa'] > 0 && parts[mapping['2fa'] - 1] ? parts[mapping['2fa'] - 1] : ''

      // 备注：从数据中读取（如果配置了位置）
      const remark = mapping.remark > 0 && parts[mapping.remark - 1] ? parts[mapping.remark - 1] : ''

      // 根据当前列表类型设置账号类型
      // all（普号列表）-> account_type = 1（普号）
      // unsold（未售列表）-> account_type = 2（成品）
      // sold（已售列表）-> account_type = 2（成品）
      const accountType = cardType.value === 'all' ? 1 : 2

      const cardData: CardRequest = {
        account,
        password: password || undefined,
        mail_password: mailPassword || undefined,
        subscription_status: 1,
        subscription_type: batchConfig.value.subscription_type || undefined,
        subscription_time: subscriptionTime,
        subscription_expired_time: subscriptionExpiredTime,
        purchase_date: purchaseDate,
        purchase_price: batchConfig.value.purchase_price,
        purchase_from: batchConfig.value.purchase_from || undefined,
        purchase_order_no: batchConfig.value.purchase_order_no || undefined,
        sell_status: 1,
        account_type: accountType,
        status: 1,
        token: token || undefined,
        api_key: apiKey || undefined,
        '2fa': twoFA || undefined,
        mail_url: batchConfig.value.mail_url || undefined,
        remark: remark || undefined,
      }

      cards.push(cardData)
    }

    if (cards.length === 0) {
      message.error('没有有效的数据')
      return false
    }

    const response = await batchImportCards({
      category: category.value,
      cards,
    })

    if (response.code === 200) {
      message.success(`成功导入 ${cards.length} 条数据`)
      showBatchModal.value = false
      await loadCards()
    } else {
      message.error(response.message || '导入失败')
      return false
    }
  } catch (error: any) {
    console.error('批量导入失败', error)
    message.error(error.response?.data?.message || '导入失败')
    return false
  }
}

// 打开取货对话框
const handlePickup = async () => {
  pickupStep.value = 1
  pickedCard.value = null
  pickupForm.value = {
    subscription_type: '',
    format: 'digiseller',
  }
  completeForm.value = {
    sell_price: 20,
    sell_to: 'Digiseller',
  }
  
  // 加载未售订阅类型
  try {
    const response = await getUnsoldSubscriptionTypes(category.value)
    if (response.code === 200) {
      unsoldSubscriptionTypes.value = (response.data || []).map((type) => ({
        label: type,
        value: type,
      }))
      if (unsoldSubscriptionTypes.value.length === 0) {
        message.warning('暂无未售卡密')
        return
      }
      // 默认选中第一个订阅类型
      if (unsoldSubscriptionTypes.value.length > 0) {
        pickupForm.value.subscription_type = unsoldSubscriptionTypes.value[0].value
      }
    } else {
      message.error(response.message || '获取订阅类型失败')
      return
    }
  } catch (error: any) {
    console.error('获取订阅类型失败', error)
    message.error('获取订阅类型失败')
    return
  }
  
  showPickupModal.value = true
}

// 提交取货（根据步骤处理）
const handlePickupSubmit = async () => {
  if (pickupStep.value === 1) {
    // 第一步：验证并执行取货
    if (!pickupForm.value.subscription_type) {
      message.error('请选择订阅类型')
      return false
    }
    
    try {
      const response = await pickupCard({
        category: category.value,
        subscription_type: pickupForm.value.subscription_type,
      })
      
      if (response.code === 200) {
        pickedCard.value = response.data
        pickupStep.value = 2
        
        // 自动复制卡密信息到剪贴板
        try {
          await navigator.clipboard.writeText(pickupCardInfo.value)
          message.success('取货成功，已自动复制到剪贴板')
        } catch (error) {
          console.error('自动复制失败', error)
          message.success('取货成功')
        }
        
        return false // 阻止关闭对话框
      } else {
        message.error(response.message || '取货失败')
        return false
      }
    } catch (error: any) {
      console.error('取货失败', error)
      message.error(error.response?.data?.message || '取货失败')
      return false
    }
  } else {
    // 第二步：完成取货
    if (!pickedCard.value) {
      message.error('未找到已取货的卡密')
      return false
    }
    
    try {
      const response = await completePickup({
        category: category.value,
        id: pickedCard.value.id,
        sell_price: completeForm.value.sell_price,
        sell_to: completeForm.value.sell_to || undefined,
      })
      
      if (response.code === 200) {
        // 复制默认文本到剪贴板
        const defaultText = `Ваш заказ выполнен !

Скорость нашей доставки быстра, как Молния Маккуин; сервис точен, как периодическая таблица Менделеева; — если вы согласны с этим, пожалуйста, оставьте положительный отзыв в заказе, и вы сразу же получите подарочную карту на сумму, равную 5% от общей суммы заказа.💰️

Подписывайтесь на наш канал, чтобы получать больше выгодных предложений: https://t.me/AI_GUO_GUO

хорошего дня )`
        
        try {
          await navigator.clipboard.writeText(defaultText)
          message.success('取货完成，已复制默认消息到剪贴板')
        } catch (error) {
          console.error('复制失败', error)
          message.success('取货完成')
        }
        
        showPickupModal.value = false
        await loadCards() // 刷新列表
      } else {
        message.error(response.message || '完成取货失败')
        return false
      }
    } catch (error: any) {
      console.error('完成取货失败', error)
      message.error(error.response?.data?.message || '完成取货失败')
      return false
    }
  }
}

// 复制取货信息
const handleCopyPickupInfo = async () => {
  try {
    await navigator.clipboard.writeText(pickupCardInfo.value)
    message.success('已复制到剪贴板')
  } catch (error) {
    console.error('复制失败', error)
    message.error('复制失败')
  }
}

// 点击卡密信息区域自动选中所有文本
const handleSelectCardInfo = () => {
  if (pickupCardInfoRef.value) {
    const range = document.createRange()
    range.selectNodeContents(pickupCardInfoRef.value)
    const selection = window.getSelection()
    if (selection) {
      selection.removeAllRanges()
      selection.addRange(range)
    }
  }
}

// 监听路由参数变化，当切换不同类型的列表时重新加载数据
watch(
  () => [route.query.category, route.query.type],
  ([newCategory, newType], [oldCategory, oldType]) => {
    // 只有当参数真正改变时才重新加载
    if (newCategory !== oldCategory || newType !== oldType) {
      console.log('🔄 路由参数变化，重新加载数据')
      console.log('  category:', oldCategory, '->', newCategory)
      console.log('  type:', oldType, '->', newType)
      
      // 重置分页和搜索条件
      pagination.value.page = 1
      searchKeyword.value = ''
      
      // 重新加载数据
      loadCards()
    }
  }
)

// 初始化
onMounted(() => {
  loadCards()
})
</script>

<style scoped>
.cards-management {
  padding: 0;
}

.card-info-display {
  margin: 0;
  padding: 16px;
  white-space: pre-wrap;
  word-wrap: break-word;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  border-radius: 8px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.8;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
  user-select: text;
  -webkit-user-select: text;
  -moz-user-select: text;
  -ms-user-select: text;
}

.card-info-display:hover {
  background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
  box-shadow: 0 4px 16px rgba(102, 126, 234, 0.5);
  transform: translateY(-2px);
}

.card-info-display:active {
  transform: translateY(0);
}
</style>

