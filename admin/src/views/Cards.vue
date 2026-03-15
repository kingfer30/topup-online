<template>
  <div class="cards-management">
    <n-card :title="pageTitle">
      <template #header-extra>
        <n-space>
          <n-button
            v-if="cardType === 'all'"
            type="warning"
            :disabled="checkedRowKeys.length === 0"
            @click="showUpgradeModal = true"
          >
            <template #icon><span>⬆️</span></template>
            更新为成品 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'unsold'" type="success" @click="handlePickup">
            <template #icon><span>📦</span></template>
            我要取货
          </n-button>
          <n-button
            v-if="cardType === 'unsold'"
            type="warning"
            :disabled="checkedRowKeys.length === 0"
            @click="showBatchPickupModal = true"
          >
            <template #icon><span>🚚</span></template>
            批量取货 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button
            v-if="cardType === 'unsold' || cardType === 'sold'"
            type="info"
            :disabled="checkedRowKeys.length === 0"
            :loading="batchCheckLoading"
            @click="handleBatchCheck"
          >
            <template #icon><span>🔍</span></template>
            批量检查 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'sold'" type="error" @click="handleOpenRecharge">
            <template #icon><span>💳</span></template>
            新增代充
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
          <n-button @click="showExportModal = true">
            <template #icon><span>📤</span></template>
            导出
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
          <n-select
            v-if="cardType === 'unsold' || cardType === 'sold'"
            v-model:value="searchSubscriptionType"
            :options="[{ label: '全部类型', value: '' }, ...subscriptionTypeOptions]"
            placeholder="订阅类型"
            clearable
            style="width: 150px"
          />
          <n-select
            v-if="cardType === 'sold'"
            v-model:value="searchSubscriptionStatus"
            :options="[{ label: '全部订阅状态', value: 0 }, ...subscriptionStatusOptions]"
            placeholder="订阅状态"
            style="width: 150px"
          />
          <n-select
            v-if="cardType === 'sold' || cardType === 'unsold'"
            v-model:value="searchIsCheck"
            :options="[
              { label: '全部检查状态', value: 0 },
              { label: '未检查', value: -1 },
              { label: '检查成功', value: 1 },
              { label: '检查失败', value: 2 },
            ]"
            placeholder="检查状态"
            style="width: 150px"
          />
          <n-input
            v-if="cardType === 'sold' || cardType === 'unsold'"
            v-model:value="searchPurchaseDate"
            placeholder="购买时间 如: 2026-03-06 22:25:36"
            clearable
            style="width: 220px"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <!-- 卡密表格 -->
        <n-data-table
          remote
          :columns="columns"
          :data="cardList"
          :pagination="pagination"
          :loading="loading"
          :bordered="false"
          :single-line="false"
          :row-key="(row: Card) => row.id"
          v-model:checked-row-keys="checkedRowKeys"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
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
              placeholder="请输入密码"
            />
          </n-form-item-gi>

          <n-form-item-gi label="邮箱密码" path="mail_password">
            <n-input
              v-model:value="formData.mail_password"
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

          <n-form-item-gi label="卖家名称" path="purchase_by">
            <n-input v-model:value="formData.purchase_by" placeholder="请输入卖家名称" />
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

          <n-form-item-gi label="出售对方" path="sell_to">
            <n-input v-model:value="formData.sell_to" placeholder="请输入出售对方" />
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

    <!-- 批量升级为成品对话框 -->
    <n-modal
      v-model:show="showUpgradeModal"
      title="批量更新为成品"
      preset="dialog"
      positive-text="确认更新"
      negative-text="取消"
      @positive-click="handleBatchUpgrade"
      style="width: 600px"
    >
      <n-form label-placement="left" label-width="120px" style="margin-top: 20px">
        <n-grid :cols="2" :x-gap="24" :y-gap="12">
          <n-form-item-gi label="订阅类型">
            <n-select
              v-model:value="upgradeForm.subscription_type"
              :options="subscriptionTypeOptions"
              placeholder="选择或输入订阅类型"
              filterable
              tag
              clearable
            />
          </n-form-item-gi>

          <n-form-item-gi label="订阅时间">
            <n-date-picker
              v-model:value="upgradeForm.subscription_time"
              type="datetime"
              placeholder="默认为当前时间"
              style="width: 100%"
              clearable
            />
          </n-form-item-gi>

          <n-form-item-gi label="购买价格(追加)">
            <n-input-number
              v-model:value="upgradeForm.purchase_price"
              :min="0"
              :precision="2"
              placeholder="追加到已有价格"
              style="width: 100%"
            />
          </n-form-item-gi>

          <n-form-item-gi label="购买平台">
            <n-select
              v-model:value="upgradeForm.purchase_from"
              :options="purchasePlatformOptions"
              placeholder="选择或输入购买平台"
              filterable
              tag
              clearable
            />
          </n-form-item-gi>

          <n-form-item-gi label="购买时间" :span="2">
            <n-date-picker
              v-model:value="upgradeForm.purchase_date"
              type="datetime"
              placeholder="默认为当前时间"
              style="width: 100%"
              clearable
            />
          </n-form-item-gi>
        </n-grid>

        <n-alert type="info" style="margin-top: 12px">
          已选择 <strong>{{ checkedRowKeys.length }}</strong> 条记录，更新后将自动设置订阅状态为"已订阅"、账号类型为"成品"。
        </n-alert>
      </n-form>
    </n-modal>

    <!-- 导出对话框 -->
    <n-modal
      v-model:show="showExportModal"
      title="导出数据"
      preset="dialog"
      :positive-text="exportLoading ? '导出中...' : '确认导出'"
      negative-text="取消"
      @positive-click="handleExport"
      style="width: 680px"
    >
      <n-space vertical :size="16" style="margin-top: 16px">
        <!-- 导出模式 -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">导出范围</div>
          <n-radio-group v-model:value="exportMode">
            <n-space>
              <n-radio value="filter">按当前筛选条件全量导出</n-radio>
              <n-radio value="selected">
                导出勾选记录
                <n-tag
                  v-if="checkedRowKeys.length > 0"
                  type="success"
                  size="small"
                  style="margin-left: 6px"
                >
                  已选 {{ checkedRowKeys.length }} 条
                </n-tag>
                <n-tag v-else type="warning" size="small" style="margin-left: 6px">未勾选</n-tag>
              </n-radio>
            </n-space>
          </n-radio-group>
        </div>

        <n-divider style="margin: 0" />

        <!-- 快速预设 -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">快速预设</div>
          <n-space>
            <n-button size="small" @click="applyExportPreset('digiseller')">Digiseller格式</n-button>
            <n-button size="small" @click="applyExportPreset('domestic')">国内格式</n-button>
            <n-button size="small" @click="applyExportPreset('reverse')">逆向格式</n-button>
            <n-button size="small" @click="applyExportPreset('all')">全部字段</n-button>
            <n-button size="small" type="error" @click="applyExportPreset('clear')">清空</n-button>
          </n-space>
        </div>

        <n-divider style="margin: 0" />

        <!-- 字段选择 -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">
            选择导出字段（顺序即为列顺序，分隔符固定为 ----）
          </div>
          <n-checkbox-group v-model:value="exportSelectedFields">
            <n-grid :cols="5" :x-gap="8" :y-gap="8">
              <n-gi v-for="field in exportFieldOptions" :key="field.value">
                <n-checkbox :value="field.value" :label="field.label" />
              </n-gi>
            </n-grid>
          </n-checkbox-group>
        </div>

        <!-- 格式预览 -->
        <div v-if="exportSelectedFields.length > 0">
          <div style="font-size: 12px; color: #888; margin-bottom: 6px">格式预览</div>
          <n-tag type="info" style="white-space: normal; word-break: break-all">
            {{ exportPreview }}
          </n-tag>
        </div>
      </n-space>
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
          <n-grid :cols="6" :x-gap="12" :y-gap="12">
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

            <n-form-item-gi label="卖家名称">
              <n-input v-model:value="batchConfig.purchase_by" placeholder="卖家名称" />
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

            <n-form-item-gi label="邮箱地址">
              <n-select
                v-model:value="batchConfig.mail_url"
                :options="mailUrlOptions"
                placeholder="选择或输入邮箱地址"
                filterable
                tag
                clearable
              />
            </n-form-item-gi>

            <n-form-item-gi label="出售状态">
              <n-select
                v-model:value="batchConfig.sell_status"
                :options="sellStatusOptions"
              />
            </n-form-item-gi>

            <n-form-item-gi label="订阅状态">
              <n-select
                v-model:value="batchConfig.subscription_status"
                :options="subscriptionStatusOptions"
              />
            </n-form-item-gi>

            <n-form-item-gi label="账号类型">
              <n-select
                v-model:value="batchConfig.account_type"
                :options="accountTypeOptions"
              />
            </n-form-item-gi>

            <n-form-item-gi label="备注">
              <n-input v-model:value="batchConfig.remark" placeholder="批量备注（选填）" />
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

    <!-- 新增代充对话框 -->
    <n-modal
      v-model:show="showRechargeModal"
      title="新增代充"
      preset="dialog"
      positive-text="确认提交"
      negative-text="取消"
      @positive-click="handleRechargeSubmit"
      style="width: 480px"
    >
      <n-form
        :model="rechargeForm"
        label-placement="left"
        label-width="100px"
        style="margin-top: 20px"
      >
        <n-form-item label="订阅类型">
          <n-select
            v-model:value="rechargeForm.subscription_type"
            :options="subscriptionTypeOptions"
            placeholder="选择或输入订阅类型"
            filterable
            tag
            clearable
          />
        </n-form-item>

        <n-form-item label="账号">
          <n-input v-model:value="rechargeForm.account" placeholder="请输入账号" />
        </n-form-item>

        <n-form-item label="购买价格">
          <n-input-number
            v-model:value="rechargeForm.purchase_price"
            :min="0"
            :precision="2"
            placeholder="代充成本"
            style="width: 100%"
          />
        </n-form-item>

        <n-form-item label="售出价格">
          <n-input-number
            v-model:value="rechargeForm.sell_price"
            :min="0"
            :precision="2"
            placeholder="售出金额"
            style="width: 100%"
          />
        </n-form-item>

        <n-form-item label="售出对方">
          <n-input v-model:value="rechargeForm.sell_to" placeholder="默认 Digiseller" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 批量取货对话框 -->
    <n-modal
      v-model:show="showBatchPickupModal"
      title="批量取货"
      preset="dialog"
      positive-text="确认取货"
      negative-text="取消"
      @positive-click="handleBatchPickup"
      style="width: 500px"
    >
      <n-space vertical :size="16" style="margin-top: 16px">
        <n-alert type="info">
          已勾选 <strong>{{ checkedRowKeys.length }}</strong> 条记录，确认后将全部标记为已出售，并复制卡密信息到剪贴板。
        </n-alert>

        <n-form
          :model="batchPickupForm"
          label-placement="left"
          label-width="100px"
        >
          <n-form-item label="复制格式">
            <n-radio-group v-model:value="batchPickupForm.format">
              <n-space>
                <n-radio value="digiseller">Digiseller</n-radio>
                <n-radio value="domestic">国内格式</n-radio>
                <n-radio value="reverse">逆向格式</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>

          <n-form-item label="售出价格">
            <n-input-number
              v-model:value="batchPickupForm.sell_price"
              :min="0"
              :precision="2"
              placeholder="非必填"
              style="width: 100%"
            />
          </n-form-item>

          <n-form-item label="售出对方">
            <n-input v-model:value="batchPickupForm.sell_to" placeholder="非必填" />
          </n-form-item>
        </n-form>
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
      @negative-click="handlePickupCancel"
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
                  <n-radio value="reverse">逆向格式</n-radio>
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

    <!-- 已发货确认弹窗 -->
    <n-modal
      v-model:show="showShippedModal"
      title="确认已发货"
      preset="dialog"
      positive-text="确认发货"
      negative-text="取消"
      @positive-click="handleShippedSubmit"
      style="width: 420px"
    >
      <n-space vertical style="margin-top: 20px" :size="16">
        <n-alert type="info">
          确认后将把该卡密状态标记为已出售
        </n-alert>
        <n-form
          :model="shippedForm"
          label-placement="left"
          label-width="100px"
        >
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
            <n-input
              v-model:value="shippedForm.sell_to"
              placeholder="非必填"
            />
          </n-form-item>
        </n-form>
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
  NDropdown,
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
  NGi,
  NAlert,
  NDatePicker,
  NRadio,
  NRadioGroup,
  NCheckbox,
  NCheckboxGroup,
  NDivider,
  useMessage,
  useDialog,
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
  batchUpgradeToProduct,
  batchPickup,
  batchCheckCards,
  exportCards,
  getUnsoldSubscriptionTypes,
  pickupCard,
  completePickup,
  rollbackPickup,
  rollbackSoldCard,
  enableOnDemandSpend,
  type Card,
  type CardRequest,
} from '@/api/card'

const route = useRoute()
const message = useMessage()
const dialog = useDialog()

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
    // 密码和邮箱密码均为空时，使用邮箱验证码登录格式
    if (!card.password && !card.mail_password) {
      return `Пожалуйста, войдите в систему, используя код подтверждения, отправленный на электронную почту:

${card.account}

mail-login: ${card.mail_url || ''}

Пожалуйста, выполните следующие шаги заново:
1. Введите аккаунт: ${card.account}
2. Нажмите «Далее»
3. Нажмите кнопку: «Email sign-in code»`
    }
    // 常规 digiseller 订阅格式
    return `account: ${card.account}
pass: ${card.password || ''}
mail-pass: ${card.mail_password || ''}

mail-login: ${card.mail_url || ''}`
  } else if (pickupForm.value.format === 'reverse') {
    // 逆向格式
    return `${card.account}----${card.token || ''}`
  } else {
    // 国内订阅格式
    return `账号----密码----邮箱密码
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
const searchSubscriptionType = ref('')
const searchSubscriptionStatus = ref(0)
const searchIsCheck = ref(0)
// 购买时间精确查询，格式 "2026-03-06 22:25:36"
const searchPurchaseDate = ref('')
const batchCheckLoading = ref(false)
const batchImportText = ref('')

// 普号列表批量勾选
const checkedRowKeys = ref<number[]>([])

// 导出相关
const showExportModal = ref(false)
const exportLoading = ref(false)
const exportMode = ref<'selected' | 'filter'>('filter')
const exportSelectedFields = ref<string[]>(['account', 'password', 'mail_password'])

// 跨页勾选数据积累：key=id，val=Card 完整数据
const selectedCardsMap = ref<Map<number, Card>>(new Map())

// 监听勾选变化，同步更新 selectedCardsMap
watch(checkedRowKeys, (newKeys) => {
  // 把当前页已勾选的行存入 Map
  for (const row of cardList.value) {
    if (newKeys.includes(row.id)) {
      selectedCardsMap.value.set(row.id, row)
    }
  }
  // 移除取消勾选的行
  const keySet = new Set(newKeys)
  for (const id of selectedCardsMap.value.keys()) {
    if (!keySet.has(id)) selectedCardsMap.value.delete(id)
  }
})

// 当前已积累的勾选卡密列表
const selectedCardsForExport = computed(() =>
  checkedRowKeys.value
    .map(id => selectedCardsMap.value.get(id as number))
    .filter((c): c is Card => !!c)
)

// 所有可导出字段定义（有序）
const exportFieldOptions = [
  { label: '账号', value: 'account' },
  { label: '密码', value: 'password' },
  { label: '邮箱密码', value: 'mail_password' },
  { label: '邮箱地址', value: 'mail_url' },
  { label: '订阅类型', value: 'subscription_type' },
  { label: '订阅状态', value: 'subscription_status' },
  { label: '订阅时间', value: 'subscription_time' },
  { label: '订阅过期时间', value: 'subscription_expired_time' },
  { label: '购买时间', value: 'purchase_date' },
  { label: '购买价格', value: 'purchase_price' },
  { label: '购买平台', value: 'purchase_from' },
  { label: '卖家名称', value: 'purchase_by' },
  { label: '出售价格', value: 'sell_price' },
  { label: '出售时间', value: 'sell_date' },
  { label: '售出对方', value: 'sell_to' },
  { label: '出售订单号', value: 'sell_order_no' },
  { label: 'API Key', value: 'api_key' },
  { label: 'Token', value: 'token' },
  { label: '2FA', value: '2fa' },
  { label: '备注', value: 'remark' },
]

// 格式预览：将选中字段的中文标签用 ---- 拼接
const exportPreview = computed(() =>
  exportSelectedFields.value
    .map(v => exportFieldOptions.find(o => o.value === v)?.label ?? v)
    .join('----')
)

// 快速预设
const applyExportPreset = (preset: string) => {
  switch (preset) {
    case 'digiseller':
      exportSelectedFields.value = ['account', 'password', 'mail_password', 'mail_url']
      break
    case 'domestic':
      exportSelectedFields.value = ['account', 'password', 'mail_password']
      break
    case 'reverse':
      exportSelectedFields.value = ['account', 'token']
      break
    case 'all':
      exportSelectedFields.value = exportFieldOptions.map(o => o.value)
      break
    case 'clear':
      exportSelectedFields.value = []
      break
  }
}

// 批量升级为成品弹窗
const showUpgradeModal = ref(false)
const upgradeForm = ref({
  subscription_type: '',
  subscription_time: Date.now() as number | undefined,   // 毫秒（n-date-picker）
  purchase_price: undefined as number | undefined,
  purchase_from: '微信',
  purchase_date: Date.now() as number | undefined,       // 毫秒（n-date-picker）
})

// 新增代充弹窗
const showRechargeModal = ref(false)
const rechargeForm = ref({
  subscription_type: '',
  account: '',
  purchase_price: undefined as number | undefined,
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 批量取货弹窗
const showBatchPickupModal = ref(false)
const batchPickupForm = ref({
  format: 'digiseller' as 'digiseller' | 'domestic' | 'reverse',
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 取货相关状态
const showPickupModal = ref(false)
const pickupStep = ref(1) // 1: 选择条件, 2: 预览确认
const unsoldSubscriptionTypes = ref<{ label: string; value: string }[]>([])
const pickedCard = ref<Card | null>(null)
const pickupCardInfoRef = ref<HTMLPreElement | null>(null)

// 取货表单
const pickupForm = ref({
  subscription_type: '',
  format: 'digiseller' as 'digiseller' | 'domestic' | 'reverse',
})

// 完成取货表单
const completeForm = ref({
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 已发货弹窗
const showShippedModal = ref(false)
const shippedCard = ref<Card | null>(null)
const shippedForm = ref({
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 批量导入配置
const batchConfig = ref({
  subscription_type: 'pro',
  subscription_remaining_days: 30 as number | undefined,
  purchase_price: undefined as number | undefined,
  purchase_from: '微信',
  purchase_by: '',
  purchase_date: undefined as number | undefined,
  mail_url: 'https://login.live.com',
  sell_status: 1,
  subscription_status: 1,
  account_type: 2,
  remark: '',
  field_mapping: {
    account: 1,
    password: 2,
    mail_password: 3,
    subscription_time: 0,
    token: 0,
    api_key: 0,
    '2fa': 0,
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
  purchase_from: '',
  purchase_by: '',
  sell_to: '',
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
  { label: '掉订阅', value: -1 },
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

// 格式化 Unix 时间戳为可读时间
const formatTimestamp = (ts?: number | null): string => {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 计算剩余天数：未售列表用当前时间，已售列表用出售时间，均与购买时间相差取 30 - 已用天数
const calcRemainingDays = (row: Card, isSold: boolean): number | null => {
  if (!row.purchase_date) return null
  const refTime = isSold && row.sell_date ? row.sell_date : Math.floor(Date.now() / 1000)
  const usedDays = Math.floor((refTime - row.purchase_date) / 86400)
  return 30 - usedDays
}

// 表格列定义（computed，根据列表类型展示不同列）
const columns = computed<DataTableColumns<Card>>(() => {
  const isSold = cardType.value === 'sold'
  const isAll = cardType.value === 'all'

  const baseColumns: DataTableColumns<Card> = [
  // 所有列表均支持批量勾选
  { type: 'selection' as const },
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
  // 普号列表显示密码和邮箱密码
  ...(isAll ? [
    {
      title: '密码',
      key: 'password',
      width: 130,
      render: (row: Card) => row.password || '—',
    },
    {
      title: '邮箱密码',
      key: 'mail_password',
      width: 130,
      render: (row: Card) => row.mail_password || '—',
    },
  ] as DataTableColumns<Card> : []),
  // 订阅类型：普号列表不显示
  ...(!isAll ? [{
    title: '订阅类型',
    key: 'subscription_type',
    width: 100,
  }] as DataTableColumns<Card> : []),
  // 订阅状态：非普号列表显示（未售+已售均显示，含掉订阅-1）
  ...(!isAll ? [
    {
      title: '订阅状态',
      key: 'subscription_status',
      width: 100,
      render: (row: Card) => {
        const map: Record<number, { label: string; type: 'success' | 'warning' | 'error' }> = {
          1:  { label: '已订阅', type: 'success' },
          2:  { label: '未订阅', type: 'warning' },
          [-1]: { label: '掉订阅', type: 'error' },
        }
        const info = map[row.subscription_status] ?? { label: String(row.subscription_status), type: 'warning' as const }
        return h(NTag, { type: info.type, size: 'small' }, { default: () => info.label })
      },
    },
  ] as DataTableColumns<Card> : []),
  ...(isSold ? [{
    title: '卖家',
    key: 'purchase_by',
    width: 80,
    render: (row: Card) => row.purchase_by || '—',
  }] as DataTableColumns<Card> : []),
  {
    title: '价格',
    key: 'purchase_price',
    width: 80,
  },
  {
    title: '订阅时间',
    key: 'subscription_time',
    width: 170,
    render: (row: Card) => formatTimestamp(row.subscription_time ?? row.purchase_date),
  },
  ...(!isSold ? [{
    title: '剩余天数',
    key: 'remaining_days',
    width: 90,
    render: (row: Card) => {
      const days = calcRemainingDays(row, isSold)
      if (days === null) return '—'
      const type = days > 7 ? 'success' : days > 0 ? 'warning' : 'error'
      return h(NTag, { type, size: 'small' }, { default: () => `${days}天` })
    },
  }] as DataTableColumns<Card> : []),
  // 未售/已售列表显示订阅额度
  ...(!isAll ? [{
    title: '额度',
    key: 'subscription_credits',
    width: 80,
    render: (row: Card) => row.subscription_credits != null ? `$${row.subscription_credits.toFixed(2)}` : '—',
  }] as DataTableColumns<Card> : []),
  // 未售/已售列表显示检查状态
  ...(!isAll ? [{
    title: '检查状态',
    key: 'is_check',
    width: 100,
    render: (row: Card) => {
      const map: Record<number, { label: string; type: 'default' | 'success' | 'error' }> = {
        [-1]: { label: '未检查', type: 'default' },
        1:    { label: '检查成功', type: 'success' },
        2:    { label: '检查失败', type: 'error' },
      }
      const val = row.is_check ?? -1
      const info = map[val] ?? { label: String(val), type: 'default' as const }
      return h(NTag, { type: info.type, size: 'small' }, { default: () => info.label })
    },
  }] as DataTableColumns<Card> : []),
  // 已售列表显示售出对方和出售时间
  ...(isSold ? [
    {
      title: '出售时间',
      key: 'sell_date',
      width: 170,
      render: (row: Card) => formatTimestamp(row.sell_date),
    },
  ] as DataTableColumns<Card> : []),
 
  ]

  // 复制格式选项
  const copyOptions = [
    { label: 'Digiseller格式', key: 'digiseller' },
    { label: '国内格式', key: 'domestic' },
    { label: '逆向格式', key: 'reverse' },
  ]

  // 操作列始终放最后
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
            default: () => h(
              NButton,
              { size: 'small', type: 'default' },
              { default: () => '复制' }
            ),
          }
        ),
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
      ]

      // 已售列表中，显示回滚按钮（已出售→未出售）
      if (cardType.value === 'sold') {
        buttons.splice(1, 0,
          h(
            NButton,
            {
              size: 'small',
              type: 'warning',
              onClick: () => handleRollbackSold(row),
            },
            { default: () => '回滚' }
          )
        )
      }

      // 未售列表中，发货中(sell_status=2)的卡密显示"已发货"和"回滚"按钮
      if (cardType.value === 'unsold' && row.sell_status === 2) {
        buttons.splice(1, 0,
          h(
            NButton,
            {
              size: 'small',
              type: 'success',
              onClick: () => handleShipped(row),
            },
            { default: () => '已发货' }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'warning',
              onClick: () => handleRollback(row),
            },
            { default: () => '回滚' }
          )
        )
      }

      // 未售/已售列表中，均显示"开启按需付费"按钮
      buttons.push(
        h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => handleEnableOnDemand(row),
          },
          { default: () => '后付费' }
        )
      )

      return h(NSpace, {}, { default: () => buttons })
    },
  })

  return baseColumns
})

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
      ...(searchSubscriptionType.value ? { subscription_type: searchSubscriptionType.value } : {}),
      ...(searchSubscriptionStatus.value !== 0 ? { subscription_status: searchSubscriptionStatus.value } : {}),
      ...(searchIsCheck.value !== 0 ? { is_check: searchIsCheck.value } : {}),
      ...(searchPurchaseDate.value ? { purchase_date: searchPurchaseDate.value } : {}),
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

// 批量检查订阅状态
const handleBatchCheck = async () => {
  if (checkedRowKeys.value.length === 0) return
  batchCheckLoading.value = true
  try {
    const response = await batchCheckCards({
      category: category.value,
      ids: checkedRowKeys.value as number[],
    })
    if (response.code === 200) {
      message.success(response.message || '检查任务已提交，正在后台执行')
      checkedRowKeys.value = []
      selectedCardsMap.value.clear()
      await loadCards()
    } else {
      message.error(response.message || '提交失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '提交失败')
  } finally {
    batchCheckLoading.value = false
  }
}

const handleEnableOnDemand = async (row: Card) => {
  try {
    const response = await enableOnDemandSpend(category.value, row.id)
    if (response.code === 200) {
      message.success(response.message || '按需付费已开启')
    } else {
      message.error(response.message || '开启失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '开启失败')
  }
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  searchSubscriptionType.value = ''
  searchSubscriptionStatus.value = 0
  searchIsCheck.value = 0
  searchPurchaseDate.value = ''
  pagination.value.page = 1
  loadCards()
}

// 分页变化
const handlePageChange = (page: number) => {
  pagination.value.page = page
  loadCards()
}

// 每页条数变化
const handlePageSizeChange = (pageSize: number) => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
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
    purchase_from: '',
    purchase_by: '',
    sell_to: '',
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
    purchase_by: card.purchase_by || '',
    sell_to: card.sell_to || '',
    api_key: card.api_key || '',
    '2fa': card['2fa'] || '',
    token: card.token || '',
    mail_url: card.mail_url || '',
    remark: card.remark || '',
  }
  showModal.value = true
}

// 删除卡密
const handleDelete = (card: Card) => {
  dialog.error({
    title: '确认删除',
    content: `确定要删除卡密"${card.account}"吗？删除后不可恢复。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
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
    },
  })
}

// 取消取货弹窗：若第一步已完成（卡密已被标记为发货中），刷新列表
const handlePickupCancel = async () => {
  if (pickedCard.value) {
    // 第一步已取货，刷新列表以显示最新状态
    await loadCards()
  }
  showPickupModal.value = false
}

// 回滚已售（已出售 -> 未出售）
const handleRollbackSold = (card: Card) => {
  dialog.error({
    title: '确认回滚',
    content: `确定要将卡密"${card.account}"从已出售回滚为未出售吗？售出记录将被清空。`,
    positiveText: '确认回滚',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response = await rollbackSoldCard({ category: category.value, id: card.id })
        if (response.code === 200) {
          message.success('回滚成功')
          await loadCards()
        } else {
          message.error(response.message || '回滚失败')
        }
      } catch (error: any) {
        console.error('回滚失败', error)
        message.error(error.response?.data?.message || '回滚失败')
      }
    },
  })
}

// 打开已发货弹窗
const handleShipped = (card: Card) => {
  shippedCard.value = card
  shippedForm.value = {
    sell_price: 20,
    sell_to: 'Digiseller',
  }
  showShippedModal.value = true
}

// 提交已发货
const handleShippedSubmit = async () => {
  if (!shippedCard.value) return

  try {
    const response = await completePickup({
      category: category.value,
      id: shippedCard.value.id,
      sell_price: shippedForm.value.sell_price,
      sell_to: shippedForm.value.sell_to || undefined,
    })
    if (response.code === 200) {
      message.success('已标记为已出售')
      showShippedModal.value = false
      await loadCards()
    } else {
      message.error(response.message || '操作失败')
      return false
    }
  } catch (error: any) {
    console.error('已发货操作失败', error)
    message.error(error.response?.data?.message || '操作失败')
    return false
  }
}

// 回滚取货（发货中 -> 未出售）
const handleRollback = (card: Card) => {
  dialog.error({
    title: '确认回滚',
    content: `确定要将卡密"${card.account}"的状态回滚为未出售吗？`,
    positiveText: '确认回滚',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response = await rollbackPickup({ category: category.value, id: card.id })
        if (response.code === 200) {
          message.success('回滚成功')
          await loadCards()
        } else {
          message.error(response.message || '回滚失败')
        }
      } catch (error: any) {
        console.error('回滚失败', error)
        message.error(error.response?.data?.message || '回滚失败')
      }
    },
  })
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

// 复制卡密信息到剪贴板
const handleCopy = async (card: Card, format: string) => {
  let text = ''
  if (format === 'digiseller') {
    text = `account: ${card.account}\npass: ${card.password || ''}\nmail-pass: ${card.mail_password || ''}\n\nmail-login: ${card.mail_url || ''}`
  } else if (format === 'reverse') {
    text = `${card.account}----${card.token || ''}`
  } else {
    // 国内格式
    text = `${card.account}----${card.password || ''}----${card.mail_password || ''}`
  }
  try {
    await navigator.clipboard.writeText(text)
    message.success('已复制')
  } catch {
    message.error('复制失败，请手动复制')
  }
}

// 获取卡密某字段的可读文本值
const getFieldValue = (card: Card, field: string): string => {
  switch (field) {
    case 'subscription_time':
    case 'subscription_expired_time':
    case 'purchase_date':
    case 'sell_date':
      return formatTimestamp((card as unknown as Record<string, unknown>)[field] as number | undefined)
    case 'subscription_status':
      return card.subscription_status === 1 ? '已订阅' : '未订阅'
    case 'sell_status':
      return (['', '未出售', '发货中', '已出售'] as string[])[card.sell_status] || ''
    case 'account_type':
      return card.account_type === 1 ? '普号' : '成品'
    case 'status':
      return card.status === 1 ? '正常' : '禁用'
    default:
      return String((card as unknown as Record<string, unknown>)[field] ?? '')
  }
}

// 生成并下载导出文件
const doDownload = (cards: Card[]) => {
  const lines = cards.map(card =>
    exportSelectedFields.value.map(f => getFieldValue(card, f)).join('----')
  )
  const content = lines.join('\n')
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${category.value}_${cardType.value}_${new Date().toISOString().slice(0, 10)}.txt`
  a.click()
  URL.revokeObjectURL(url)
}

// 执行导出
const handleExport = async () => {
  if (exportSelectedFields.value.length === 0) {
    message.warning('请至少选择一个导出字段')
    return false
  }

  // 勾选导出：直接使用已积累的勾选数据
  if (exportMode.value === 'selected') {
    const cards = selectedCardsForExport.value
    if (cards.length === 0) {
      message.warning('请先勾选要导出的记录')
      return false
    }
    doDownload(cards)
    message.success(`成功导出 ${cards.length} 条数据`)
    showExportModal.value = false
    return true
  }

  // 筛选条件全量导出：请求后端
  exportLoading.value = true
  try {
    const response = await exportCards({
      category: category.value,
      type: cardType.value,
      keyword: searchKeyword.value,
      ...(searchSubscriptionType.value ? { subscription_type: searchSubscriptionType.value } : {}),
    })

    if (response.code === 200) {
      doDownload(response.data)
      message.success(`成功导出 ${response.data.length} 条数据`)
      showExportModal.value = false
    } else {
      message.error(response.message || '导出失败')
      return false
    }
  } catch (error: unknown) {
    message.error('导出失败: ' + (error instanceof Error ? error.message : '未知错误'))
    return false
  } finally {
    exportLoading.value = false
  }
}

// 批量升级为成品
const handleBatchUpgrade = async () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning('请先勾选要升级的记录')
    return false
  }

  try {
    const now = Math.floor(Date.now() / 1000)
    const subscriptionTime = upgradeForm.value.subscription_time
      ? Math.floor(upgradeForm.value.subscription_time / 1000)
      : now
    const purchaseDate = upgradeForm.value.purchase_date
      ? Math.floor(upgradeForm.value.purchase_date / 1000)
      : now

    const response = await batchUpgradeToProduct({
      category: category.value,
      ids: checkedRowKeys.value as number[],
      subscription_type: upgradeForm.value.subscription_type || undefined,
      subscription_time: subscriptionTime,
      purchase_price: upgradeForm.value.purchase_price,
      purchase_from: upgradeForm.value.purchase_from || undefined,
      purchase_date: purchaseDate,
    })

    if (response.code === 200) {
      message.success(response.message || '更新成功')
      showUpgradeModal.value = false
      checkedRowKeys.value = []
      // 重置表单
      upgradeForm.value = {
        subscription_type: '',
        subscription_time: Date.now(),
        purchase_price: undefined,
        purchase_from: '微信',
        purchase_date: Date.now(),
      }
      await loadCards()
    } else {
      message.error(response.message || '更新失败')
      return false
    }
  } catch (error: any) {
    message.error('更新失败: ' + (error.message || '未知错误'))
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
    purchase_by: '',
    purchase_date: undefined,
    mail_url: 'https://login.live.com',
    sell_status: 1,
    subscription_status: 1,
    account_type: 2,
    remark: '',
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

      const cardData: CardRequest = {
        account,
        password: password || undefined,
        mail_password: mailPassword || undefined,
        subscription_status: batchConfig.value.subscription_status,
        subscription_type: batchConfig.value.subscription_type || undefined,
        subscription_time: subscriptionTime,
        subscription_expired_time: subscriptionExpiredTime,
        purchase_date: purchaseDate,
        purchase_price: batchConfig.value.purchase_price,
        purchase_from: batchConfig.value.purchase_from || undefined,
        purchase_by: batchConfig.value.purchase_by || undefined,
        sell_status: batchConfig.value.sell_status,
        account_type: batchConfig.value.account_type,
        status: 1,
        token: token || undefined,
        api_key: apiKey || undefined,
        '2fa': twoFA || undefined,
        mail_url: batchConfig.value.mail_url || undefined,
        remark: batchConfig.value.remark || undefined,
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
// 打开代充弹窗（重置表单）
const handleOpenRecharge = () => {
  rechargeForm.value = {
    subscription_type: '',
    account: '',
    purchase_price: undefined,
    sell_price: 20,
    sell_to: 'Digiseller',
  }
  showRechargeModal.value = true
}

// 提交代充：直接创建一条已售出的记录
const handleRechargeSubmit = async () => {
  if (!rechargeForm.value.account.trim()) {
    message.error('请输入账号')
    return false
  }

  const now = Math.floor(Date.now() / 1000)
  try {
    const response = await createCard(category.value, {
      account: rechargeForm.value.account.trim(),
      subscription_type: rechargeForm.value.subscription_type || undefined,
      subscription_status: 1,
      purchase_price: rechargeForm.value.purchase_price,
      purchase_date: now,
      sell_price: rechargeForm.value.sell_price,
      sell_to: rechargeForm.value.sell_to || undefined,
      sell_status: 3,
      sell_date: now,
      account_type: 2,
      status: 1,
    })

    if (response.code === 200) {
      message.success('代充记录已提交')
      showRechargeModal.value = false
      await loadCards()
    } else {
      message.error(response.message || '提交失败')
      return false
    }
  } catch (error: unknown) {
    message.error('提交失败: ' + (error instanceof Error ? error.message : '未知错误'))
    return false
  }
}

// 批量取货处理
const handleBatchPickup = async () => {
  const ids = checkedRowKeys.value as number[]
  if (ids.length === 0) {
    message.warning('请先勾选要取货的记录')
    return false
  }

  try {
    const response = await batchPickup({
      category: category.value,
      ids,
      sell_price: batchPickupForm.value.sell_price,
      sell_to: batchPickupForm.value.sell_to || undefined,
    })

    if (response.code !== 200) {
      message.error(response.message || '批量取货失败')
      return false
    }

    // 取出勾选卡密的数据，按格式拼接后复制到剪贴板
    const cards = ids
      .map(id => selectedCardsMap.value.get(id))
      .filter((c): c is Card => !!c)

    const fmt = batchPickupForm.value.format
    const lines = cards.map(card => {
      if (fmt === 'digiseller') {
        return `account: ${card.account}\npass: ${card.password || ''}\nmail-pass: ${card.mail_password || ''}\n\nmail-login: ${card.mail_url || ''}`
      } else if (fmt === 'reverse') {
        return `${card.account}----${card.token || ''}`
      } else {
        return `${card.account}----${card.password || ''}----${card.mail_password || ''}`
      }
    })

    try {
      await navigator.clipboard.writeText(lines.join('\n\n'))
      message.success(`批量取货完成，已复制 ${ids.length} 条卡密信息到剪贴板`)
    } catch {
      message.success(`批量取货完成（${response.data} 条），剪贴板复制失败请手动复制`)
    }

    showBatchPickupModal.value = false
    checkedRowKeys.value = []
    selectedCardsMap.value.clear()
    await loadCards()
  } catch (error: unknown) {
    message.error('批量取货失败: ' + (error instanceof Error ? error.message : '未知错误'))
    return false
  }
}

const handlePickup = async () => {
  pickupStep.value = 1
  pickedCard.value = null
  pickupForm.value = {
    subscription_type: '',
    format: 'digiseller' as 'digiseller' | 'domestic' | 'reverse',
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
        format: pickupForm.value.format,
      })
      
      if (response.code === 200) {
        pickedCard.value = response.data
        pickupStep.value = 2

        // 刷新列表，使 sell_status=2(发货中) 状态即时生效
        await loadCards()
        
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

// 监听取货格式变化，同步更新售出对方默认值
watch(
  () => pickupForm.value.format,
  (format) => {
    completeForm.value.sell_to = format === 'reverse' ? '自用逆向' : 'Digiseller'
  }
)

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
      searchSubscriptionType.value = ''
      
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

:deep(.n-card) {
  border-radius: 16px !important;
  border: 1px solid rgba(0, 0, 0, 0.04) !important;
}

:deep(.n-card-header__main) {
  font-size: 17px;
  font-weight: 600;
  color: #1d1d1f;
}

.card-info-display {
  margin: 0;
  padding: 16px;
  white-space: pre-wrap;
  word-wrap: break-word;
  background: linear-gradient(135deg, #007AFF 0%, #5856D6 100%);
  color: #ffffff;
  border-radius: 12px;
  font-family: 'SF Mono', 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  box-shadow: 0 4px 16px rgba(0, 122, 255, 0.2);
  user-select: text;
  -webkit-user-select: text;
  -moz-user-select: text;
  -ms-user-select: text;
}

.card-info-display:hover {
  background: linear-gradient(135deg, #5856D6 0%, #007AFF 100%);
  box-shadow: 0 8px 24px rgba(0, 122, 255, 0.3);
  transform: translateY(-2px);
}

.card-info-display:active {
  transform: translateY(0);
}
</style>

