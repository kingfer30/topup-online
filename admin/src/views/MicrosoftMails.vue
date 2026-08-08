<template>
  <div class="ms-mails-management">
    <n-card :title="pageTitle">
      <template #header-extra>
        <n-space>
          <n-button v-if="listType === 'unsold'" type="success" @click="handlePickup">
            <template #icon><span>📦</span></template>
            我要取货
          </n-button>
          <n-button
            v-if="listType === 'unsold'"
            type="warning"
            :disabled="checkedRowKeys.length === 0"
            @click="showBatchPickupModal = true"
          >
            <template #icon><span>🚚</span></template>
            批量取货 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button
            type="info"
            :disabled="checkedRowKeys.length === 0"
            :loading="batchCheckLoading"
            @click="handleBatchCheck"
          >
            <template #icon><span>🔍</span></template>
            批量检查 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button type="error" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            <template #icon><span>🗑️</span></template>
            批量删除 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button type="primary" @click="handleAdd">
            <template #icon><span>➕</span></template>
            新增邮箱
          </n-button>
          <n-button type="info" @click="handleBatchImport">
            <template #icon><span>📥</span></template>
            批量导入
          </n-button>
          <n-button @click="showExportModal = true">
            <template #icon><span>📤</span></template>
            导出
          </n-button>
        </n-space>
      </template>

      <n-space vertical :size="16">
        <n-space>
          <n-input
            v-model:value="searchAccountsText"
            type="textarea"
            placeholder="搜索账号，每行一个"
            clearable
            :rows="3"
            style="width: 300px"
          />
          <n-input
            v-if="listType === 'sold'"
            v-model:value="searchSellTo"
            placeholder="出售对方"
            clearable
            style="width: 160px"
          />
          <n-input
            v-if="listType === 'sold'"
            v-model:value="searchPurchaseBy"
            placeholder="卖家"
            clearable
            style="width: 140px"
          />
          <n-select
            v-model:value="searchIsCheck"
            :options="isCheckFilterOptions"
            placeholder="检查状态"
            style="width: 150px"
          />
          <n-date-picker
            v-model:value="searchPurchaseDate"
            type="date"
            clearable
            style="width: 180px"
            :placeholder="listType === 'sold' ? '已售日期' : '购买日期'"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <n-data-table
          remote
          :columns="columns"
          :data="mailList"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: MicrosoftMail) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>

    <!-- 新增/编辑 -->
    <n-modal
      v-model:show="showModal"
      :title="isEdit ? '编辑邮箱' : '新增邮箱'"
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
        style="margin-top: 20px"
      >
        <n-grid :cols="2" :x-gap="24">
          <n-form-item-gi label="账号" path="account">
            <n-input v-model:value="formData.account" placeholder="邮箱账号" />
          </n-form-item-gi>
          <n-form-item-gi label="邮箱密码" path="password">
            <n-input v-model:value="formData.password" placeholder="邮箱密码" />
          </n-form-item-gi>
          <n-form-item-gi label="邮箱地址" path="mail_url">
            <n-input v-model:value="formData.mail_url" placeholder="mail_url" />
          </n-form-item-gi>
          <n-form-item-gi label="Client ID" path="client_id">
            <n-input v-model:value="formData.client_id" placeholder="client_id" />
          </n-form-item-gi>
          <n-form-item-gi label="Refresh Token" path="token" :span="2">
            <n-input v-model:value="formData.token" type="textarea" :rows="3" placeholder="refresh_token" />
          </n-form-item-gi>
          <n-form-item-gi label="2FA" path="two_fa">
            <n-input v-model:value="formData.two_fa" placeholder="2fa" />
          </n-form-item-gi>
          <n-form-item-gi label="出售状态" path="sell_status">
            <n-select v-model:value="formData.sell_status" :options="sellStatusOptions" />
          </n-form-item-gi>
          <n-form-item-gi label="购买价格" path="purchase_price">
            <n-input-number
              v-model:value="formData.purchase_price"
              :min="0"
              :precision="2"
              style="width: 100%"
            />
          </n-form-item-gi>
          <n-form-item-gi label="购买平台" path="purchase_from">
            <n-input v-model:value="formData.purchase_from" />
          </n-form-item-gi>
          <n-form-item-gi label="卖家名称" path="purchase_by">
            <n-input v-model:value="formData.purchase_by" />
          </n-form-item-gi>
          <n-form-item-gi label="出售价格" path="sell_price">
            <n-input-number
              v-model:value="formData.sell_price"
              :min="0"
              :precision="2"
              style="width: 100%"
            />
          </n-form-item-gi>
          <n-form-item-gi label="出售对方" path="sell_to">
            <n-input v-model:value="formData.sell_to" />
          </n-form-item-gi>
          <n-form-item-gi label="所属卡密ID" path="account_card_id">
            <n-input-number v-model:value="formData.account_card_id" :min="0" style="width: 100%" clearable />
          </n-form-item-gi>
          <n-form-item-gi label="所属卡密表" path="account_card_table">
            <n-input v-model:value="formData.account_card_table" placeholder="如 cards_cursor" />
          </n-form-item-gi>
          <n-form-item-gi label="备注" path="remark" :span="2">
            <n-input v-model:value="formData.remark" />
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-modal>

    <!-- 批量导入 -->
    <n-modal
      v-model:show="showBatchModal"
      title="批量导入微软邮箱"
      preset="dialog"
      positive-text="导入"
      negative-text="取消"
      :mask-closable="!batchImportLoading"
      :close-on-esc="!batchImportLoading"
      :closable="!batchImportLoading"
      :positive-button-props="{ loading: batchImportLoading, disabled: batchImportLoading }"
      :negative-button-props="{ disabled: batchImportLoading }"
      @positive-click="handleBatchSubmit"
      @negative-click="handleBatchCancel"
      style="width: 1200px; max-width: 95vw"
    >
      <n-spin :show="batchImportLoading" description="正在导入，请稍候...">
        <n-space vertical style="margin-top: 20px" :size="16">
          <n-card title="公共字段配置" size="small">
            <n-grid :cols="6" :x-gap="12" :y-gap="12">
              <n-form-item-gi label="购买价格">
                <n-input-number
                  v-model:value="batchConfig.purchase_price"
                  :min="0"
                  :precision="2"
                  placeholder="购买价格"
                  style="width: 100%"
                  :disabled="batchImportLoading"
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
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="卖家名称">
                <n-input
                  v-model:value="batchConfig.purchase_by"
                  placeholder="卖家名称"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="购买时间">
                <n-date-picker
                  v-model:value="batchConfig.purchase_date"
                  type="datetime"
                  placeholder="默认为当前时间"
                  style="width: 100%"
                  clearable
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="邮箱地址">
                <n-select
                  v-model:value="batchConfig.mail_url"
                  :options="mailUrlOptions"
                  placeholder="选择或输入邮箱地址"
                  filterable
                  tag
                  clearable
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="出售状态">
                <n-select
                  v-model:value="batchConfig.sell_status"
                  :options="sellStatusOptions"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="所属卡密表">
                <n-input
                  v-model:value="batchConfig.account_card_table"
                  placeholder="如 cards_cursor"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="所属卡密ID">
                <n-input-number
                  v-model:value="batchConfig.account_card_id"
                  :min="0"
                  placeholder="选填"
                  style="width: 100%"
                  clearable
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="备注">
                <n-input
                  v-model:value="batchConfig.remark"
                  placeholder="批量备注（选填）"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
            </n-grid>
          </n-card>

          <n-card title="字段映射配置" size="small">
            <n-alert type="info" style="margin-bottom: 12px">
              导入数据使用分隔符"----"分割，请配置各字段的顺序位置（从1开始，0表示不导入该字段）。
              Outlook 默认行：邮箱----密码----refresh_token----client_id
            </n-alert>
            <n-grid :cols="8" :x-gap="12" :y-gap="12">
              <n-form-item-gi label="账号">
                <n-input-number
                  v-model:value="batchConfig.field_mapping.account"
                  :min="1"
                  placeholder="必填"
                  style="width: 100%"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="邮箱密码">
                <n-input-number
                  v-model:value="batchConfig.field_mapping.password"
                  :min="0"
                  placeholder="0=不导入"
                  style="width: 100%"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="Token">
                <n-input-number
                  v-model:value="batchConfig.field_mapping.token"
                  :min="0"
                  placeholder="0=不导入"
                  style="width: 100%"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="Client ID">
                <n-input-number
                  v-model:value="batchConfig.field_mapping.client_id"
                  :min="0"
                  placeholder="0=不导入"
                  style="width: 100%"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
              <n-form-item-gi label="2FA">
                <n-input-number
                  v-model:value="batchConfig.field_mapping['2fa']"
                  :min="0"
                  placeholder="0=不导入"
                  style="width: 100%"
                  :disabled="batchImportLoading"
                />
              </n-form-item-gi>
            </n-grid>
          </n-card>

          <n-alert type="warning">
            每行一条数据，字段之间使用"----"分隔。根据上方字段映射配置，按顺序填写对应字段数据。
          </n-alert>
          <n-input
            v-model:value="batchImportText"
            type="textarea"
            :rows="8"
            :disabled="batchImportLoading"
            placeholder="示例（默认映射：账号=1，密码=2，Token=3，Client ID=4）：&#10;email@outlook.com----password----M.C5_BAY...----9e5f94bc-e8a4-4e73-b8be-63364c29d753&#10;email2@hotmail.com----password2----M.C5_BAY...----9e5f94bc-e8a4-4e73-b8be-63364c29d753"
          />
        </n-space>
      </n-spin>
    </n-modal>

    <!-- 导出 -->
    <n-modal
      v-model:show="showExportModal"
      title="导出邮箱"
      preset="dialog"
      positive-text="导出"
      negative-text="取消"
      :loading="exportLoading"
      @positive-click="handleExport"
      style="width: 560px"
    >
      <n-space vertical :size="12" style="margin-top: 12px">
        <n-radio-group v-model:value="exportMode">
          <n-space>
            <n-radio value="filter">按当前筛选导出</n-radio>
            <n-radio value="selected" :disabled="checkedRowKeys.length === 0">
              导出已勾选 ({{ checkedRowKeys.length }})
            </n-radio>
          </n-space>
        </n-radio-group>
        <n-checkbox-group v-model:value="exportSelectedFields">
          <n-space>
            <n-checkbox v-for="opt in exportFieldOptions" :key="opt.value" :value="opt.value" :label="opt.label" />
          </n-space>
        </n-checkbox-group>
        <div class="text-sm text-gray-500">预览：{{ exportPreview }}</div>
      </n-space>
    </n-modal>

    <!-- 取货 -->
    <n-modal
      v-model:show="showPickupModal"
      :title="pickupStep === 1 ? '我要取货 - 选择格式' : '我要取货 - 预览确认'"
      preset="dialog"
      :positive-text="pickupStep === 1 ? '下一步' : '完成取货'"
      negative-text="取消"
      @positive-click="handlePickupSubmit"
      @negative-click="handlePickupCancel"
      style="width: 640px"
    >
      <div v-if="pickupStep === 1" style="margin-top: 12px">
        <n-form label-placement="left" label-width="100px">
          <n-form-item label="取货格式">
            <n-radio-group v-model:value="pickupFormat">
              <n-space>
                <n-radio value="normal">标准格式</n-radio>
                <n-radio value="digiseller">Digiseller</n-radio>
                <n-radio value="reverse">逆向(账号----token)</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>
        </n-form>
      </div>
      <div v-else style="margin-top: 12px">
        <n-alert type="success" :bordered="false" style="margin-bottom: 12px">
          已选出一条邮箱，确认后完成取货
        </n-alert>
        <pre class="card-info-display" @click="selectPickupText">{{ pickupInfo }}</pre>
        <n-form label-placement="left" label-width="100px" style="margin-top: 12px">
          <n-form-item label="出售价格">
            <n-input-number v-model:value="completeSellPrice" :min="0" :precision="2" clearable />
          </n-form-item>
          <n-form-item label="出售对方">
            <n-input v-model:value="completeSellTo" clearable />
          </n-form-item>
        </n-form>
      </div>
    </n-modal>

    <!-- 批量取货 -->
    <n-modal
      v-model:show="showBatchPickupModal"
      title="批量取货"
      preset="dialog"
      positive-text="确认取货"
      negative-text="取消"
      @positive-click="handleBatchPickupSubmit"
      style="width: 480px"
    >
      <n-form label-placement="left" label-width="100px" style="margin-top: 12px">
        <n-form-item label="出售价格">
          <n-input-number v-model:value="batchPickupSellPrice" :min="0" :precision="2" clearable />
        </n-form-item>
        <n-form-item label="出售对方">
          <n-input v-model:value="batchPickupSellTo" clearable />
        </n-form-item>
      </n-form>
    </n-modal>
    <!-- 已发货确认 -->
    <n-modal
      v-model:show="showShippedModal"
      title="确认已发货"
      preset="dialog"
      positive-text="确认"
      negative-text="取消"
      @positive-click="handleShippedSubmit"
      style="width: 420px"
    >
      <n-space vertical style="margin-top: 20px" :size="16">
        <n-alert type="info">确认后将把该邮箱状态标记为已出售</n-alert>
        <n-form :model="shippedForm" label-placement="left" label-width="100px">
          <n-form-item label="售出价格">
            <n-input-number
              v-model:value="shippedForm.sell_price"
              :min="0"
              :precision="2"
              placeholder="非必填"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="售出对方">
            <n-input v-model:value="shippedForm.sell_to" placeholder="非必填" />
          </n-form-item>
        </n-form>
      </n-space>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
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
  NCheckbox,
  NCheckboxGroup,
  NDropdown,
  NSpin,
  useMessage,
  useDialog,
  type DataTableColumns,
  type FormInst,
  type FormRules,
} from 'naive-ui'
import {
  getMicrosoftMailList,
  createMicrosoftMail,
  updateMicrosoftMail,
  deleteMicrosoftMail,
  batchImportMicrosoftMails,
  exportMicrosoftMails,
  pickupMicrosoftMail,
  completeMicrosoftMailPickup,
  rollbackMicrosoftMailPickup,
  rollbackMicrosoftMailSold,
  batchPickupMicrosoftMails,
  batchCheckMicrosoftMails,
  batchDeleteMicrosoftMails,
  updateMicrosoftMailRemark,
  type MicrosoftMail,
  type MicrosoftMailRequest,
} from '@/api/microsoft-mail'

const route = useRoute()
const message = useMessage()
const dialog = useDialog()

const listType = computed(() => {
  const t = String(route.query.type || 'unsold')
  return t === 'sold' ? 'sold' : 'unsold'
})
const pageTitle = computed(() => (listType.value === 'sold' ? '微软邮箱已售' : '微软邮箱未售'))

const loading = ref(false)
const mailList = ref<MicrosoftMail[]>([])
const checkedRowKeys = ref<(string | number)[]>([])
const searchAccountsText = ref('')
const searchSellTo = ref('')
const searchPurchaseBy = ref('')
const searchIsCheck = ref(0)
const searchPurchaseDate = ref<number | null>(null)
const batchCheckLoading = ref(false)
const exportLoading = ref(false)

const pagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [20, 50, 100, 200],
})

const isCheckFilterOptions = [
  { label: '全部检查状态', value: 0 },
  { label: '未检查', value: -1 },
  { label: '检查成功', value: 1 },
  { label: '检查失败', value: 2 },
]
const sellStatusOptions = [
  { label: '未出售', value: 1 },
  { label: '发货中', value: 2 },
  { label: '已出售', value: 3 },
]

const purchasePlatformOptions = [
  { label: '支付宝', value: '支付宝' },
  { label: '微信', value: '微信' },
  { label: 'Telegram', value: 'Telegram' },
  { label: '闲鱼', value: '闲鱼' },
  { label: '淘宝', value: '淘宝' },
  { label: '卡充', value: '卡充' },
]

const mailUrlOptions = [
  { label: 'https://login.live.com', value: 'https://login.live.com' },
  { label: 'https://mail.com', value: 'https://mail.com' },
  { label: 'https://gmx.us', value: 'https://gmx.us' },
  { label: 'https://outlook.live.com', value: 'https://outlook.live.com' },
]

const formatTimestamp = (ts?: number | null): string => {
  if (!ts) return '-'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const sellStatusTag = (status: number) => {
  if (status === 2) return h(NTag, { type: 'warning', size: 'small', bordered: false }, { default: () => '发货中' })
  if (status === 3) return h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '已出售' })
  return h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => '未出售' })
}

const checkStatusTag = (status?: number) => {
  if (status === 1) return h(NTag, { type: 'success', size: 'small' }, { default: () => '检查成功' })
  if (status === 2) return h(NTag, { type: 'error', size: 'small' }, { default: () => '检查失败' })
  return h(NTag, { type: 'default', size: 'small' }, { default: () => '未检查' })
}

const handleCopy = async (mail: MicrosoftMail, format: string) => {
  let text = ''
  if (format === 'digiseller') {
    text = `account: ${mail.account}\npass: ${mail.password || ''}\n\nmail-login: ${mail.mail_url || ''}`
  } else if (format === 'digiseller_auto') {
    text = `account: ${mail.account}<br>pass: ${mail.password || ''}<br>mail-login: ${mail.mail_url || ''}`
  } else if (format === 'reverse') {
    text = `${mail.account}----${mail.token || ''}`
  } else if (format === 'outlook') {
    text = `${mail.account}|${mail.password || ''}|${mail.token || ''}|${mail.client_id || ''}`
  } else {
    text = `${mail.account}----${mail.password || ''}----${mail.token || ''}----${mail.client_id || ''}`
  }
  try {
    await navigator.clipboard.writeText(text)
    message.success('已复制')
  } catch {
    message.error('复制失败，请手动复制')
  }
}

const showShippedModal = ref(false)
const shippedMail = ref<MicrosoftMail | null>(null)
const shippedForm = ref({
  sell_price: undefined as number | undefined,
  sell_to: 'Digiseller',
})

const handleShipped = (row: MicrosoftMail) => {
  shippedMail.value = row
  shippedForm.value = {
    sell_price: row.sell_price ?? undefined,
    sell_to: row.sell_to || 'Digiseller',
  }
  showShippedModal.value = true
}

const handleShippedSubmit = async () => {
  if (!shippedMail.value) return false
  const res = await completeMicrosoftMailPickup({
    id: shippedMail.value.id,
    sell_price: shippedForm.value.sell_price,
    sell_to: shippedForm.value.sell_to || undefined,
  })
  if (res.code === 200) {
    message.success('已标记为已出售')
    showShippedModal.value = false
    shippedMail.value = null
    loadData()
    return true
  }
  message.error(res.message || '操作失败')
  return false
}

const columns = computed<DataTableColumns<MicrosoftMail>>(() => {
  const isSold = listType.value === 'sold'
  const isUnsold = listType.value === 'unsold'

  const baseColumns: DataTableColumns<MicrosoftMail> = [
    { type: 'selection' as const },
    { title: 'ID', key: 'id', width: 60 },
    {
      title: '账号',
      key: 'account',
      width: 220,
      render: (row) => {
        const tags: ReturnType<typeof h>[] = []
        if (isUnsold && row.sell_status === 2) {
          tags.push(sellStatusTag(2))
        }
        if (tags.length > 0) {
          return h(
            'div',
            { style: { display: 'flex', flexDirection: 'column', gap: '6px', lineHeight: '1.2' } },
            [
              h('div', { style: { fontWeight: 500 } }, row.account),
              h(NSpace, { size: 6, wrap: true }, { default: () => tags }),
            ]
          )
        }
        return h('div', { style: { fontWeight: 500 } }, row.account)
      },
    },
    {
      title: '卖家',
      key: 'purchase_by',
      width: 80,
      render: (row) => row.purchase_by || '—',
    },
    {
      title: '价格',
      key: isSold ? 'sell_price' : 'purchase_price',
      width: 80,
      render: (row) => {
        const price = isSold ? row.sell_price : row.purchase_price
        return price != null ? String(price) : '—'
      },
    },
    {
      title: isSold ? '出售时间' : '购买时间',
      key: isSold ? 'sell_date' : 'purchase_date',
      width: 170,
      render: (row) => formatTimestamp(isSold ? row.sell_date : row.purchase_date),
    },
    {
      title: '检查状态',
      key: 'is_check',
      width: 100,
      render: (row) => checkStatusTag(row.is_check),
    },
    {
      title: '下次检查',
      key: 'next_check_time',
      width: 170,
      render: (row) => formatTimestamp(row.next_check_time),
    },
  ]

  if (isSold) {
    baseColumns.splice(5, 0, {
      title: '出售对方',
      key: 'sell_to',
      width: 120,
      render: (row) => row.sell_to || '—',
    })
  }

  const copyOptions = [
    { label: 'Digiseller格式', key: 'digiseller' },
    { label: 'Digiseller自动发货', key: 'digiseller_auto' },
    { label: '国内格式', key: 'domestic' },
    { label: '逆向格式', key: 'reverse' },
    { label: 'Outlook行格式', key: 'outlook' },
  ]

  baseColumns.push({
    title: '操作',
    key: 'actions',
    width: 350,
    fixed: 'right',
    render: (row) => {
      const buttons = [
        h(
          NDropdown,
          {
            trigger: 'click',
            options: copyOptions,
            onSelect: (key: string) => handleCopy(row, key),
          },
          {
            default: () =>
              h(NButton, { size: 'small', type: 'default' }, { default: () => '复制' }),
          }
        ),
        h(
          NButton,
          { size: 'small', type: 'primary', onClick: () => handleEdit(row) },
          { default: () => '编辑' }
        ),
      ]

      if (isSold) {
        buttons.splice(
          1,
          0,
          h(
            NButton,
            { size: 'small', type: 'warning', onClick: () => handleRollbackSold(row) },
            { default: () => '回滚' }
          )
        )
      }

      if (isUnsold && row.sell_status === 2) {
        buttons.splice(
          1,
          0,
          h(
            NButton,
            { size: 'small', type: 'success', onClick: () => handleShipped(row) },
            { default: () => '已发货' }
          ),
          h(
            NButton,
            { size: 'small', type: 'warning', onClick: () => handleRollbackPickup(row) },
            { default: () => '回滚' }
          )
        )
      }

      buttons.push(
        h(
          NButton,
          { size: 'small', quaternary: true, onClick: () => handleQuickRemark(row) },
          { default: () => '备注' }
        ),
        h(
          NButton,
          { size: 'small', type: 'error', onClick: () => handleDelete(row) },
          { default: () => '删除' }
        )
      )

      return h(NSpace, { size: 8 }, { default: () => buttons })
    },
  })

  return baseColumns
})

const buildListParams = () => {
  const params: Record<string, unknown> = {
    type: listType.value,
    page: pagination.value.page,
    page_size: pagination.value.pageSize,
  }
  if (searchAccountsText.value.trim()) params.accounts = searchAccountsText.value.trim()
  if (searchIsCheck.value !== 0) params.is_check = searchIsCheck.value
  if (searchSellTo.value.trim()) params.sell_to = searchSellTo.value.trim()
  if (searchPurchaseBy.value.trim()) params.purchase_by = searchPurchaseBy.value.trim()
  if (searchPurchaseDate.value) {
    const d = new Date(searchPurchaseDate.value)
    const pad = (n: number) => String(n).padStart(2, '0')
    params.purchase_date = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
  }
  return params
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMicrosoftMailList(buildListParams() as any)
    if (res.code === 200 && res.data) {
      mailList.value = res.data.list || []
      pagination.value.itemCount = res.data.total || 0
    } else {
      message.error(res.message || '加载失败')
    }
  } catch (e: any) {
    message.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  loadData()
}
const handleReset = () => {
  searchAccountsText.value = ''
  searchSellTo.value = ''
  searchPurchaseBy.value = ''
  searchIsCheck.value = 0
  searchPurchaseDate.value = null
  pagination.value.page = 1
  loadData()
}
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadData()
}
const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadData()
}

// ---- 表单 ----
const showModal = ref(false)
const isEdit = ref(false)
const editingId = ref(0)
const formRef = ref<FormInst | null>(null)
const formData = ref({
  account: '',
  password: '',
  mail_url: '',
  client_id: '',
  token: '',
  two_fa: '',
  sell_status: 1,
  purchase_price: null as number | null,
  purchase_from: '',
  purchase_by: '',
  sell_price: null as number | null,
  sell_to: '',
  account_card_id: null as number | null,
  account_card_table: '',
  remark: '',
})
const rules: FormRules = {
  account: [{ required: true, message: '请输入账号', trigger: 'blur' }],
}

const resetForm = () => {
  formData.value = {
    account: '',
    password: '',
    mail_url: '',
    client_id: '',
    token: '',
    two_fa: '',
    sell_status: 1,
    purchase_price: null,
    purchase_from: '',
    purchase_by: '',
    sell_price: null,
    sell_to: '',
    account_card_id: null,
    account_card_table: '',
    remark: '',
  }
}

const handleAdd = () => {
  isEdit.value = false
  editingId.value = 0
  resetForm()
  showModal.value = true
}

const handleEdit = (row: MicrosoftMail) => {
  isEdit.value = true
  editingId.value = row.id
  formData.value = {
    account: row.account || '',
    password: row.password || '',
    mail_url: row.mail_url || '',
    client_id: row.client_id || '',
    token: row.token || '',
    two_fa: row['2fa'] || '',
    sell_status: row.sell_status || 1,
    purchase_price: row.purchase_price ?? null,
    purchase_from: row.purchase_from || '',
    purchase_by: row.purchase_by || '',
    sell_price: row.sell_price ?? null,
    sell_to: row.sell_to || '',
    account_card_id: row.account_card_id ?? null,
    account_card_table: row.account_card_table || '',
    remark: row.remark || '',
  }
  showModal.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return false
  }
  const payload: MicrosoftMailRequest = {
    account: formData.value.account,
    password: formData.value.password,
    mail_url: formData.value.mail_url,
    client_id: formData.value.client_id,
    token: formData.value.token,
    '2fa': formData.value.two_fa,
    sell_status: formData.value.sell_status,
    purchase_price: formData.value.purchase_price ?? undefined,
    purchase_from: formData.value.purchase_from,
    purchase_by: formData.value.purchase_by,
    sell_price: formData.value.sell_price ?? undefined,
    sell_to: formData.value.sell_to,
    account_card_id: formData.value.account_card_id,
    account_card_table: formData.value.account_card_table,
    remark: formData.value.remark,
  }
  try {
    const res = isEdit.value
      ? await updateMicrosoftMail(editingId.value, payload)
      : await createMicrosoftMail(payload)
    if (res.code === 200) {
      message.success(isEdit.value ? '保存成功' : '创建成功')
      showModal.value = false
      loadData()
      return true
    }
    message.error(res.message || '操作失败')
    return false
  } catch (e: any) {
    message.error(e?.message || '操作失败')
    return false
  }
}

const handleDelete = (row: MicrosoftMail) => {
  dialog.warning({
    title: '确认删除',
    content: `确定删除账号 ${row.account} 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await deleteMicrosoftMail(row.id)
      if (res.code === 200) {
        message.success('删除成功')
        loadData()
      } else {
        message.error(res.message || '删除失败')
      }
    },
  })
}

const handleBatchDelete = () => {
  dialog.warning({
    title: '批量删除',
    content: `确定删除选中的 ${checkedRowKeys.value.length} 条记录吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await batchDeleteMicrosoftMails(checkedRowKeys.value.map(Number))
      if (res.code === 200) {
        message.success(res.message || '删除成功')
        checkedRowKeys.value = []
        loadData()
      } else {
        message.error(res.message || '删除失败')
      }
    },
  })
}

const handleBatchCheck = async () => {
  batchCheckLoading.value = true
  try {
    const res = await batchCheckMicrosoftMails(checkedRowKeys.value.map(Number))
    if (res.code === 200) {
      message.success(res.message || '已提交检查任务')
      setTimeout(() => loadData(), 2000)
    } else {
      message.error(res.message || '提交失败')
    }
  } finally {
    batchCheckLoading.value = false
  }
}

const handleRollbackPickup = (row: MicrosoftMail) => {
  dialog.warning({
    title: '确认回滚',
    content: `确定要将邮箱"${row.account}"从发货中回滚为未出售吗？`,
    positiveText: '确认回滚',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await rollbackMicrosoftMailPickup(row.id)
      if (res.code === 200) {
        message.success('已回滚为未出售')
        loadData()
      } else {
        message.error(res.message || '回滚失败')
      }
    },
  })
}

const handleRollbackSold = (row: MicrosoftMail) => {
  dialog.error({
    title: '确认回滚',
    content: `确定要将邮箱"${row.account}"从已出售回滚为未出售吗？售出记录将被清空。`,
    positiveText: '确认回滚',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await rollbackMicrosoftMailSold(row.id)
      if (res.code === 200) {
        message.success('已回滚为未出售')
        loadData()
      } else {
        message.error(res.message || '回滚失败')
      }
    },
  })
}

const handleQuickRemark = (row: MicrosoftMail) => {
  let remark = row.remark || ''
  dialog.create({
    title: '更新备注',
    content: () =>
      h('input', {
        value: remark,
        style: 'width:100%;padding:8px;border:1px solid #ddd;border-radius:4px',
        onInput: (e: Event) => {
          remark = (e.target as HTMLInputElement).value
        },
      }),
    positiveText: '保存',
    negativeText: '取消',
    onPositiveClick: async () => {
      const res = await updateMicrosoftMailRemark(row.id, remark)
      if (res.code === 200) {
        message.success('备注已更新')
        loadData()
      } else {
        message.error(res.message || '更新失败')
      }
    },
  })
}

// ---- 导入 ----
const showBatchModal = ref(false)
const batchImportLoading = ref(false)
const batchImportText = ref('')
const batchConfig = ref({
  purchase_price: undefined as number | undefined,
  purchase_from: '支付宝',
  purchase_by: '',
  purchase_date: undefined as number | undefined,
  mail_url: 'https://login.live.com',
  sell_status: 1,
  account_card_id: null as number | null,
  account_card_table: '',
  remark: '',
  field_mapping: {
    account: 1,
    password: 2,
    token: 3,
    client_id: 4,
    '2fa': 0,
  },
})

const handleBatchImport = () => {
  if (batchImportLoading.value) return
  batchImportText.value = ''
  const currentMapping = { ...batchConfig.value.field_mapping }
  batchConfig.value = {
    purchase_price: undefined,
    purchase_from: '支付宝',
    purchase_by: '',
    purchase_date: undefined,
    mail_url: 'https://login.live.com',
    sell_status: 1,
    account_card_id: null,
    account_card_table: '',
    remark: '',
    field_mapping: currentMapping,
  }
  showBatchModal.value = true
}

const handleBatchCancel = () => !batchImportLoading.value

const handleBatchSubmit = async () => {
  if (batchImportLoading.value) return false
  if (!batchImportText.value.trim()) {
    message.error('请输入要导入的数据')
    return false
  }

  try {
    const lines = batchImportText.value.trim().split('\n')
    const mails: MicrosoftMailRequest[] = []
    const currentTimestamp = Math.floor(Date.now() / 1000)
    const purchaseDate = batchConfig.value.purchase_date
      ? Math.floor(batchConfig.value.purchase_date / 1000)
      : currentTimestamp
    const mapping = batchConfig.value.field_mapping

    for (const line of lines) {
      const trimmedLine = line.trim()
      if (!trimmedLine) continue

      // 兼容 ---- 与 | 分隔
      const parts = trimmedLine.includes('----')
        ? trimmedLine.split('----').map((p) => p.trim())
        : trimmedLine.split('|').map((p) => p.trim())
      if (parts.length === 0) continue

      const account = mapping.account > 0 && parts[mapping.account - 1] ? parts[mapping.account - 1] : ''
      if (!account) continue

      const password =
        mapping.password > 0 && parts[mapping.password - 1] ? parts[mapping.password - 1] : ''
      const token = mapping.token > 0 && parts[mapping.token - 1] ? parts[mapping.token - 1] : ''
      const clientId =
        mapping.client_id > 0 && parts[mapping.client_id - 1] ? parts[mapping.client_id - 1] : ''
      const twoFA = mapping['2fa'] > 0 && parts[mapping['2fa'] - 1] ? parts[mapping['2fa'] - 1] : ''

      mails.push({
        account,
        password: password || undefined,
        token: token || undefined,
        client_id: clientId || undefined,
        '2fa': twoFA || undefined,
        purchase_date: purchaseDate,
        purchase_price: batchConfig.value.purchase_price,
        purchase_from: batchConfig.value.purchase_from || undefined,
        purchase_by: batchConfig.value.purchase_by || undefined,
        sell_status: batchConfig.value.sell_status,
        status: 1,
        mail_url: batchConfig.value.mail_url || undefined,
        remark: batchConfig.value.remark || undefined,
        account_card_id: batchConfig.value.account_card_id,
        account_card_table: batchConfig.value.account_card_table || undefined,
      })
    }

    if (mails.length === 0) {
      message.error('没有有效的数据')
      return false
    }

    batchImportLoading.value = true
    const res = await batchImportMicrosoftMails(mails)
    if (res.code === 200) {
      message.success(`成功导入 ${mails.length} 条数据`)
      batchImportText.value = ''
      showBatchModal.value = false
      loadData()
      return true
    }
    message.error(res.message || '导入失败')
    return false
  } catch (error: any) {
    message.error(error?.response?.data?.message || error?.message || '导入失败')
    return false
  } finally {
    batchImportLoading.value = false
  }
}

// ---- 导出 ----
const showExportModal = ref(false)
const exportMode = ref<'filter' | 'selected'>('filter')
const exportSelectedFields = ref(['account', 'password', 'token', 'client_id'])
const exportFieldOptions = [
  { label: '账号', value: 'account' },
  { label: '密码', value: 'password' },
  { label: 'Token', value: 'token' },
  { label: 'Client ID', value: 'client_id' },
  { label: '邮箱地址', value: 'mail_url' },
  { label: '2FA', value: '2fa' },
  { label: '购买价格', value: 'purchase_price' },
  { label: '购买平台', value: 'purchase_from' },
  { label: '卖家', value: 'purchase_by' },
  { label: '出售价格', value: 'sell_price' },
  { label: '出售对方', value: 'sell_to' },
  { label: '所属卡密ID', value: 'account_card_id' },
  { label: '所属卡密表', value: 'account_card_table' },
  { label: '备注', value: 'remark' },
]
const exportPreview = computed(() =>
  exportSelectedFields.value
    .map((v) => exportFieldOptions.find((o) => o.value === v)?.label ?? v)
    .join('----')
)

const fieldValue = (mail: MicrosoftMail, field: string): string => {
  if (field === '2fa') return mail['2fa'] || ''
  const val = (mail as any)[field]
  if (val === null || val === undefined) return ''
  return String(val)
}

const handleExport = async () => {
  exportLoading.value = true
  try {
    let rows: MicrosoftMail[] = []
    if (exportMode.value === 'selected') {
      const idSet = new Set(checkedRowKeys.value.map(Number))
      rows = mailList.value.filter((m) => idSet.has(m.id))
    } else {
      const res = await exportMicrosoftMails(buildListParams() as any)
      if (res.code !== 200) {
        message.error(res.message || '导出失败')
        return false
      }
      rows = res.data || []
    }
    if (rows.length === 0) {
      message.warning('没有可导出的数据')
      return false
    }
    const lines = rows.map((m) =>
      exportSelectedFields.value.map((f) => fieldValue(m, f)).join('----')
    )
    const blob = new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `microsoft-mails-${listType.value}-${Date.now()}.txt`
    a.click()
    URL.revokeObjectURL(url)
    message.success(`已导出 ${rows.length} 条`)
    showExportModal.value = false
    return true
  } finally {
    exportLoading.value = false
  }
}

// ---- 取货 ----
const showPickupModal = ref(false)
const pickupStep = ref(1)
const pickupFormat = ref('normal')
const pickedMail = ref<MicrosoftMail | null>(null)
const completeSellPrice = ref<number | null>(null)
const completeSellTo = ref('')
const showBatchPickupModal = ref(false)
const batchPickupSellPrice = ref<number | null>(null)
const batchPickupSellTo = ref('')

const pickupInfo = computed(() => {
  const mail = pickedMail.value
  if (!mail) return ''
  if (pickupFormat.value === 'digiseller') {
    return `account: ${mail.account}
pass: ${mail.password || ''}

mail-login: ${mail.mail_url || ''}`
  }
  if (pickupFormat.value === 'reverse') {
    return `${mail.account}----${mail.token || ''}`
  }
  return `账号----密码----token----client_id
${mail.account}----${mail.password || ''}----${mail.token || ''}----${mail.client_id || ''}`
})

const selectPickupText = () => {
  const text = pickupInfo.value
  if (text) navigator.clipboard?.writeText(text).then(() => message.success('已复制'))
}

const handlePickup = () => {
  pickupStep.value = 1
  pickedMail.value = null
  completeSellPrice.value = null
  completeSellTo.value = ''
  showPickupModal.value = true
}

const handlePickupCancel = () => {
  if (pickedMail.value && pickupStep.value === 2) {
    rollbackMicrosoftMailPickup(pickedMail.value.id).finally(() => {
      pickedMail.value = null
      pickupStep.value = 1
    })
  }
  return true
}

const handlePickupSubmit = async () => {
  if (pickupStep.value === 1) {
    const res = await pickupMicrosoftMail(pickupFormat.value)
    if (res.code !== 200 || !res.data) {
      message.error(res.message || '取货失败')
      return false
    }
    pickedMail.value = res.data
    pickupStep.value = 2
    return false
  }
  if (!pickedMail.value) return false
  const res = await completeMicrosoftMailPickup({
    id: pickedMail.value.id,
    sell_price: completeSellPrice.value ?? undefined,
    sell_to: completeSellTo.value || undefined,
  })
  if (res.code === 200) {
    message.success('取货完成')
    showPickupModal.value = false
    pickupStep.value = 1
    pickedMail.value = null
    loadData()
    return true
  }
  message.error(res.message || '完成取货失败')
  return false
}

const handleBatchPickupSubmit = async () => {
  const res = await batchPickupMicrosoftMails({
    ids: checkedRowKeys.value.map(Number),
    sell_price: batchPickupSellPrice.value ?? undefined,
    sell_to: batchPickupSellTo.value || undefined,
  })
  if (res.code === 200) {
    message.success(res.message || '批量取货成功')
    checkedRowKeys.value = []
    showBatchPickupModal.value = false
    loadData()
    return true
  }
  message.error(res.message || '批量取货失败')
  return false
}

watch(
  () => route.query.type,
  () => {
    checkedRowKeys.value = []
    pagination.value.page = 1
    loadData()
  }
)

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.card-info-display {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.5;
}
</style>
