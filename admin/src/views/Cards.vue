<template>
  <div class="cards-management">
    <n-card :title="pageTitle">
      <template #header-extra>
        <n-space>
          <n-button v-if="cardType === 'all'" type="warning" :disabled="checkedRowKeys.length === 0"
            @click="handleOpenUpgradeModal">
            <template #icon><span>⬆️</span></template>
            更新为成品 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'all'" type="error" :disabled="checkedRowKeys.length === 0"
            @click="handleOpenFreezeModal(1)">
            <template #icon><span>🔒</span></template>
            批量冻结 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'all'" type="default" :disabled="checkedRowKeys.length === 0"
            @click="handleOpenFreezeModal(-1)">
            <template #icon><span>🔓</span></template>
            批量解冻 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'unsold'" type="success" @click="handlePickup">
            <template #icon><span>📦</span></template>
            我要取货
          </n-button>
          <n-button v-if="cardType === 'unsold' || cardType === 'all'" type="warning"
            :disabled="checkedRowKeys.length === 0" @click="handleOpenBatchPickup">
            <template #icon><span>🚚</span></template>
            批量取货 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'all' || cardType === 'unsold' || cardType === 'sold'" type="info"
            :disabled="checkedRowKeys.length === 0" :loading="batchCheckLoading" @click="handleBatchCheck">
            <template #icon><span>🔍</span></template>
            批量检查 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'unsold' || cardType === 'sold'" type="warning"
            :disabled="checkedRowKeys.length === 0" :loading="batchOnDemandLoading" @click="handleBatchEnableOnDemand">
            <template #icon><span>💸</span></template>
            批量后付费 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'sold'" type="warning" :disabled="checkedRowKeys.length === 0"
            :loading="batchGotoProLoading" @click="handleBatchGotoPro">
            <template #icon><span>🔗</span></template>
            批量提链 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button v-if="cardType === 'sold'" type="error" @click="handleOpenRecharge">
            <template #icon><span>💳</span></template>
            新增代充
          </n-button>
          <n-button type="error" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
            <template #icon><span>🗑️</span></template>
            批量删除 {{ checkedRowKeys.length > 0 ? `(${checkedRowKeys.length})` : '' }}
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
          <n-button @click="handleOpenExportModal">
            <template #icon><span>📤</span></template>
            导出
          </n-button>
        </n-space>
      </template>

      <!-- 搜索栏 -->
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
          <n-select v-if="cardType === 'unsold' || cardType === 'sold'" v-model:value="searchSubscriptionType"
            :options="[{ label: '全部类型', value: '' }, ...subscriptionTypeOptions]" placeholder="订阅类型" clearable
            style="width: 150px" />
          <n-select v-if="cardType === 'sold'" v-model:value="searchSubscriptionStatus"
            :options="[{ label: '全部订阅状态', value: 0 }, ...subscriptionStatusOptions]" placeholder="订阅状态"
            style="width: 150px" />
          <n-input
            v-if="cardType === 'sold'"
            v-model:value="searchSellTo"
            placeholder="出售对方"
            clearable
            style="width: 160px"
          />
          <n-input
            v-if="cardType === 'sold'"
            v-model:value="searchPurchaseBy"
            placeholder="卖家"
            clearable
            style="width: 140px"
          />
          <n-select v-if="cardType === 'sold' || cardType === 'unsold'" v-model:value="searchIsCheck" :options="[
            { label: '全部检查状态', value: 0 },
            { label: '未检查', value: -1 },
            { label: '检查成功', value: 1 },
            { label: '检查失败', value: 2 },
          ]" placeholder="检查状态" style="width: 150px" />
          <n-select v-if="cardType === 'all'" v-model:value="searchFreezeStatus" :options="[
            { label: '全部冻结状态', value: 0 },
            { label: '未冻结', value: -1 },
            { label: '已冻结', value: 1 },
          ]" placeholder="冻结状态" style="width: 150px" />
          <n-date-picker
            v-if="cardType === 'all'"
            v-model:value="searchFreezeTime"
            type="date"
            clearable
            style="width: 180px"
            placeholder="冻结时间"
          />
          <n-date-picker
            v-if="cardType === 'sold' || cardType === 'unsold'"
            v-model:value="searchPurchaseDate"
            type="date"
            clearable
            style="width: 180px"
            :placeholder="cardType === 'sold' ? '已售日期' : '购买日期'"
          />
          <n-button type="primary" @click="handleSearch">搜索</n-button>
          <n-button @click="handleReset">重置</n-button>
        </n-space>

        <!-- 卡密表格 -->
        <n-data-table remote :columns="columns" :data="cardList" :pagination="pagination" :loading="loading"
          :bordered="false" :single-line="false" :row-key="(row: Card) => row.id"
          v-model:checked-row-keys="checkedRowKeys" @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange" />
      </n-space>
    </n-card>

    <!-- 新增/编辑对话框 -->
    <n-modal v-model:show="showModal" :title="isEdit ? '编辑卡密' : '新增卡密'" preset="dialog"
      :positive-text="isEdit ? '保存' : '创建'" negative-text="取消" @positive-click="handleSubmit" style="width: 800px">
      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="140px"
        require-mark-placement="left" style="margin-top: 20px">
        <n-grid :cols="2" :x-gap="24">
          <n-form-item-gi label="账号" path="account">
            <n-input v-model:value="formData.account" placeholder="请输入账号" />
          </n-form-item-gi>

          <n-form-item-gi label="密码" path="password">
            <n-input v-model:value="formData.password" placeholder="请输入密码" />
          </n-form-item-gi>

          <n-form-item-gi label="邮箱密码" path="mail_password">
            <n-input v-model:value="formData.mail_password" placeholder="请输入邮箱密码" />
          </n-form-item-gi>

          <n-form-item-gi label="邮箱地址" path="mail_url">
            <n-input v-model:value="formData.mail_url" placeholder="请输入邮箱地址" />
          </n-form-item-gi>

          <n-form-item-gi label="订阅类型" path="subscription_type">
            <n-select v-model:value="formData.subscription_type" :options="subscriptionTypeOptions"
              placeholder="选择或输入订阅类型" filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅状态" path="subscription_status">
            <n-select v-model:value="formData.subscription_status" :options="subscriptionStatusOptions" />
          </n-form-item-gi>

          <n-form-item-gi label="购买价格" path="purchase_price">
            <n-input-number v-model:value="formData.purchase_price" :min="0" :precision="2" placeholder="购买价格"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买平台" path="purchase_from">
            <n-select v-model:value="formData.purchase_from" :options="purchasePlatformOptions" placeholder="选择或输入购买平台"
              filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi label="卖家名称" path="purchase_by">
            <n-input v-model:value="formData.purchase_by" placeholder="请输入卖家名称" />
          </n-form-item-gi>

          <n-form-item-gi label="出售价格" path="sell_price">
            <n-input-number v-model:value="formData.sell_price" :min="0" :precision="2" placeholder="出售价格"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="出售状态" path="sell_status">
            <n-select v-model:value="formData.sell_status" :options="sellStatusOptions" />
          </n-form-item-gi>

          <n-form-item-gi label="出售对方" path="sell_to">
            <n-input v-model:value="formData.sell_to" placeholder="请输入出售对方" />
          </n-form-item-gi>

          <n-form-item-gi label="出售订单号" path="sell_order_no">
            <n-input v-model:value="formData.sell_order_no" placeholder="请输入出售订单号" />
          </n-form-item-gi>

          <n-form-item-gi label="订阅时间" path="subscription_time">
            <n-date-picker v-model:value="formData.subscription_time" type="datetime" placeholder="订阅时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅过期时间" path="subscription_expired_time">
            <n-date-picker v-model:value="formData.subscription_expired_time" type="datetime" placeholder="订阅过期时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅额度" path="subscription_credits">
            <n-input-number v-model:value="formData.subscription_credits" :min="0" :precision="2" placeholder="订阅额度"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买时间" path="purchase_date">
            <n-date-picker v-model:value="formData.purchase_date" type="datetime" placeholder="购买时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="出售时间" path="sell_date">
            <n-date-picker v-model:value="formData.sell_date" type="datetime" placeholder="出售时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="账号类型" path="account_type">
            <n-select v-model:value="formData.account_type" :options="accountTypeOptions" />
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

          <n-form-item-gi :span="2" label="Client ID" path="client_id">
            <n-input v-model:value="formData.client_id" placeholder="请输入 Client ID" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="Token" path="token">
            <n-input v-model:value="formData.token" type="textarea" placeholder="请输入Token" :rows="3" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="接码链接" path="code_link">
            <n-select v-model:value="formData.code_link" :options="codeLinkOptions" placeholder="选择或输入接码链接"
              filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="手机号" path="phone">
            <n-input v-model:value="formData.phone" placeholder="请输入手机号" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="短信接码地址" path="phone_link">
            <n-input v-model:value="formData.phone_link" placeholder="请输入手机号接码地址" />
          </n-form-item-gi>

          <n-form-item-gi :span="2" label="备注" path="remark">
            <n-input v-model:value="formData.remark" type="textarea" placeholder="请输入备注" :rows="2" />
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-modal>

    <!-- 批量升级为成品对话框 -->
    <n-modal v-model:show="showUpgradeModal" title="批量更新为成品" preset="dialog" positive-text="确认更新" negative-text="取消"
      @positive-click="handleBatchUpgrade" style="width: 600px">
      <n-form label-placement="left" label-width="120px" style="margin-top: 20px">
        <n-grid :cols="2" :x-gap="24" :y-gap="12">
          <n-form-item-gi label="订阅类型" required>
            <n-select v-model:value="upgradeForm.subscription_type" :options="subscriptionTypeOptions"
              placeholder="请选择订阅类型" filterable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅时间">
            <n-date-picker v-model:value="upgradeForm.subscription_time" type="datetime" placeholder="默认为当前时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅剩余天数">
            <n-input-number v-model:value="upgradeForm.subscription_remaining_days" :min="0" placeholder="默认30天"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买价格(追加)">
            <n-input-number v-model:value="upgradeForm.purchase_price" :min="0" :precision="2" placeholder="追加到已有价格"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买平台">
            <n-select v-model:value="upgradeForm.purchase_from" :options="purchasePlatformOptions"
              placeholder="选择或输入购买平台" filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi label="购买时间" :span="2">
            <n-date-picker v-model:value="upgradeForm.purchase_date" type="datetime" placeholder="默认为当前时间"
              style="width: 100%" clearable />
          </n-form-item-gi>
        </n-grid>

        <n-alert type="info" style="margin-top: 12px">
          已选择 <strong>{{ checkedRowKeys.length }}</strong> 条记录，更新后将自动设置订阅状态为"已订阅"、账号类型为"成品"。
        </n-alert>
      </n-form>
    </n-modal>

    <!-- 导出对话框 -->
    <n-modal v-model:show="showExportModal" title="导出数据" preset="dialog"
      :positive-text="exportLoading ? '导出中...' : '确认导出'" negative-text="取消" @positive-click="handleExport"
      style="width: 680px">
      <n-space vertical :size="16" style="margin-top: 16px">
        <!-- 导出模式 -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">导出范围</div>
          <n-radio-group v-model:value="exportMode">
            <n-space>
              <n-radio value="filter">按当前筛选条件全量导出</n-radio>
              <n-radio value="selected">
                导出勾选记录
                <n-tag v-if="checkedRowKeys.length > 0" type="success" size="small" style="margin-left: 6px">
                  已选 {{ checkedRowKeys.length }} 条
                </n-tag>
                <n-tag v-else type="warning" size="small" style="margin-left: 6px">未勾选</n-tag>
              </n-radio>
            </n-space>
          </n-radio-group>
        </div>

        <n-divider style="margin: 0" />

        <!-- 快速预设（与复制/批量取货格式对应） -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">快速预设</div>
          <n-space>
            <n-button size="small" @click="applyExportPreset('digiseller')">Digiseller格式</n-button>
            <n-button size="small" @click="applyExportPreset('digiseller_auto')">Digiseller自动发货</n-button>
            <n-button size="small" @click="applyExportPreset('domestic')">国内格式</n-button>
            <n-button size="small" @click="applyExportPreset('reverse')">逆向格式</n-button>
            <n-button size="small" @click="applyExportPreset('code_method')">接码方式</n-button>
            <n-button size="small" @click="applyExportPreset('all')">全部字段</n-button>
            <n-button size="small" type="error" @click="applyExportPreset('clear')">清空</n-button>
          </n-space>
        </div>

        <n-divider style="margin: 0" />

        <!-- 字段选择 -->
        <div>
          <div style="font-size: 12px; color: #888; margin-bottom: 8px">
            {{ exportFormatHint }}
          </div>
          <n-checkbox-group :value="exportSelectedFields" @update:value="handleExportFieldsUpdate">
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
    <n-modal v-model:show="showBatchModal" title="批量导入卡密" preset="dialog" positive-text="导入" negative-text="取消"
      @positive-click="handleBatchSubmit" style="width: 1200px; max-width: 95vw">
      <n-space vertical style="margin-top: 20px" :size="16">
        <!-- 公共字段配置 -->
        <n-card title="公共字段配置" size="small">
          <n-grid :cols="6" :x-gap="12" :y-gap="12">
            <n-form-item-gi label="订阅类型">
              <n-select v-model:value="batchConfig.subscription_type" :options="subscriptionTypeOptions"
                placeholder="选择或输入订阅类型" filterable tag clearable />
            </n-form-item-gi>

            <n-form-item-gi label="订阅剩余天数">
              <n-input-number v-model:value="batchConfig.subscription_remaining_days" :min="0" placeholder="默认30天"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="购买价格">
              <n-input-number v-model:value="batchConfig.purchase_price" :min="0" :precision="2" placeholder="购买价格"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="购买平台">
              <n-select v-model:value="batchConfig.purchase_from" :options="purchasePlatformOptions"
                placeholder="选择或输入购买平台" filterable tag clearable />
            </n-form-item-gi>

            <n-form-item-gi label="卖家名称">
              <n-input v-model:value="batchConfig.purchase_by" placeholder="卖家名称" />
            </n-form-item-gi>

            <n-form-item-gi label="购买时间">
              <n-date-picker v-model:value="batchConfig.purchase_date" type="datetime" placeholder="默认为当前时间"
                style="width: 100%" clearable />
            </n-form-item-gi>

            <n-form-item-gi label="邮箱地址">
              <n-select v-model:value="batchConfig.mail_url" :options="mailUrlOptions" placeholder="选择或输入邮箱地址"
                filterable tag clearable />
            </n-form-item-gi>

            <n-form-item-gi label="出售状态">
              <n-select v-model:value="batchConfig.sell_status" :options="sellStatusOptions" />
            </n-form-item-gi>

            <n-form-item-gi label="订阅状态">
              <n-select v-model:value="batchConfig.subscription_status" :options="subscriptionStatusOptions" />
            </n-form-item-gi>

            <n-form-item-gi label="账号类型">
              <n-select v-model:value="batchConfig.account_type" :options="accountTypeOptions" />
            </n-form-item-gi>

            <n-form-item-gi label="接码链接">
              <n-select v-model:value="batchConfig.code_link" :options="codeLinkOptions" placeholder="选择或输入接码链接"
                filterable tag clearable />
            </n-form-item-gi>

            <n-form-item-gi label="短信接码地址">
              <n-input v-model:value="batchConfig.phone_link" placeholder="批量默认手机号接码地址（选填）" />
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
              <n-input-number v-model:value="batchConfig.field_mapping.account" :min="1" placeholder="必填"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="密码">
              <n-input-number v-model:value="batchConfig.field_mapping.password" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="邮箱密码">
              <n-input-number v-model:value="batchConfig.field_mapping.mail_password" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="订阅类型">
              <n-input-number v-model:value="batchConfig.field_mapping.subscription_type" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="订阅时间">
              <n-input-number v-model:value="batchConfig.field_mapping.subscription_time" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="订阅过期时间">
              <n-input-number v-model:value="batchConfig.field_mapping.subscription_expired_time" :min="0"
                placeholder="0=不导入" style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="Token">
              <n-input-number v-model:value="batchConfig.field_mapping.token" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="Client ID">
              <n-input-number v-model:value="batchConfig.field_mapping.client_id" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="API Key">
              <n-input-number v-model:value="batchConfig.field_mapping.api_key" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="2FA">
              <n-input-number v-model:value="batchConfig.field_mapping['2fa']" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="手机号">
              <n-input-number v-model:value="batchConfig.field_mapping.phone" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>

            <n-form-item-gi label="短信接码地址">
              <n-input-number v-model:value="batchConfig.field_mapping.phone_link" :min="0" placeholder="0=不导入"
                style="width: 100%" />
            </n-form-item-gi>
          </n-grid>
        </n-card>

        <!-- 导入数据输入 -->
        <n-alert type="warning">
          每行一条数据，字段之间使用"----"分隔。根据上方字段映射配置，按顺序填写对应字段数据。
        </n-alert>
        <n-input v-model:value="batchImportText" type="textarea"
          placeholder="示例（假设配置：账号=1，密码=2，邮箱密码=3）：&#10;account1@example.com----password1----mailpass1&#10;account2@example.com----password2----mailpass2&#10;&#10;包含Token示例（假设配置：账号=1，密码=2，Token=3）：&#10;account3@example.com----password3----token123"
          :rows="8" />
      </n-space>
    </n-modal>

    <!-- 新增代充对话框 -->
    <n-modal v-model:show="showRechargeModal" title="新增代充" preset="dialog" positive-text="确认提交" negative-text="取消"
      @positive-click="handleRechargeSubmit" style="width: 480px">
      <n-form :model="rechargeForm" label-placement="left" label-width="100px" style="margin-top: 20px">
        <n-form-item label="订阅类型">
          <n-select v-model:value="rechargeForm.subscription_type" :options="subscriptionTypeOptions"
            placeholder="选择或输入订阅类型" filterable tag clearable />
        </n-form-item>

        <n-form-item label="账号">
          <n-input v-model:value="rechargeForm.account" placeholder="请输入账号" />
        </n-form-item>

        <n-form-item label="购买价格">
          <n-input-number v-model:value="rechargeForm.purchase_price" :min="0" :precision="2" placeholder="代充成本"
            style="width: 100%" />
        </n-form-item>

        <n-form-item label="售出价格">
          <n-input-number v-model:value="rechargeForm.sell_price" :min="0" :precision="2" placeholder="售出金额"
            style="width: 100%" />
        </n-form-item>

        <n-form-item label="售出对方">
          <n-input v-model:value="rechargeForm.sell_to" placeholder="默认 Digiseller" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 批量取货对话框 -->
    <n-modal v-model:show="showBatchPickupModal" title="批量取货" preset="dialog" positive-text="确认取货" negative-text="取消"
      @positive-click="handleBatchPickup" style="width: 500px">
      <n-space vertical :size="16" style="margin-top: 16px">
        <n-alert type="info">
          已勾选 <strong>{{ checkedRowKeys.length }}</strong> 条记录，确认后将全部标记为已出售，并复制卡密信息到剪贴板。
        </n-alert>

        <n-form :model="batchPickupForm" label-placement="left" label-width="100px">
          <n-form-item label="复制格式">
            <n-radio-group v-model:value="batchPickupForm.format">
              <n-space>
                <n-radio value="digiseller">Digiseller</n-radio>
                <n-radio value="digiseller_auto">Digiseller自动发货</n-radio>
                <n-radio value="domestic">国内格式</n-radio>
                <n-radio value="reverse">逆向格式</n-radio>
                <n-radio value="code_method">接码方式</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>

          <n-form-item label="售出价格">
            <n-input-number v-model:value="batchPickupForm.sell_price" :min="0" :precision="2" placeholder="非必填"
              style="width: 100%" />
          </n-form-item>

          <n-form-item label="售出对方">
            <n-input v-model:value="batchPickupForm.sell_to" placeholder="非必填" />
          </n-form-item>
        </n-form>
      </n-space>
    </n-modal>

    <!-- 取货对话框 -->
    <n-modal v-model:show="showPickupModal" :title="pickupStep === 1 ? '我要取货 - 选择条件' : '我要取货 - 预览确认'" preset="dialog"
      :positive-text="pickupStep === 1 ? '下一步' : '完成取货'" negative-text="取消" @positive-click="handlePickupSubmit"
      @negative-click="handlePickupCancel" style="width: 700px">
      <n-space vertical style="margin-top: 20px" :size="16">
        <!-- 第一步：选择条件 -->
        <div v-if="pickupStep === 1">
          <n-form :model="pickupForm" label-placement="left" label-width="100px">
            <n-form-item label="订阅类型" path="subscription_type">
              <n-select v-model:value="pickupForm.subscription_type" :options="unsoldSubscriptionTypes"
                placeholder="请选择订阅类型" />
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
              <n-button size="small" style="position: absolute; top: -40px; right: 0" @click="handleCopyPickupInfo">
                复制
              </n-button>
              <pre ref="pickupCardInfoRef" class="card-info-display" @click="handleSelectCardInfo">{{ pickupCardInfo }}
          </pre>
            </div>
          </n-card>

          <!-- 售出信息 -->
          <n-form :model="completeForm" label-placement="left" label-width="100px">
            <n-form-item label="售出价格">
              <n-input-number v-model:value="completeForm.sell_price" :min="0" :precision="2" placeholder="非必填"
                style="width: 100%" />
            </n-form-item>

            <n-form-item label="售出对方">
              <n-input v-model:value="completeForm.sell_to" placeholder="非必填" />
            </n-form-item>
          </n-form>
        </div>
      </n-space>
    </n-modal>

    <!-- 已发货确认弹窗 -->
    <n-modal v-model:show="showShippedModal" title="确认已发货" preset="dialog" positive-text="确认发货" negative-text="取消"
      @positive-click="handleShippedSubmit" style="width: 420px">
      <n-space vertical style="margin-top: 20px" :size="16">
        <n-alert type="info">
          确认后将把该卡密状态标记为已出售
        </n-alert>
        <n-form :model="shippedForm" label-placement="left" label-width="100px">
          <n-form-item label="售出价格">
            <n-input-number v-model:value="shippedForm.sell_price" :min="0" :precision="2" placeholder="非必填"
              style="width: 100%" />
          </n-form-item>
          <n-form-item label="售出对方">
            <n-input v-model:value="shippedForm.sell_to" placeholder="非必填" />
          </n-form-item>
        </n-form>
      </n-space>
    </n-modal>

    <!-- 提链结果弹窗 -->
    <n-modal v-model:show="showGotoProModal" :title="`提链成功 - ${gotoProAccount || '付款链接'}`" preset="card" style="width: 640px">
      <n-space vertical :size="12">
        <n-alert type="success">付款链接已生成，正在自动提交 USD + Alipay 账单</n-alert>
        <n-input v-model:value="gotoProLink" type="textarea" :rows="3" placeholder="粘贴或覆盖 Stripe 结账链接"
          style="font-size: 12px; word-break: break-all" />
        <n-button size="small" @click="handleCopyGotoProLink">复制结账链接</n-button>

        <n-alert v-if="stripeAlipayLoading" type="info">正在填写账单并生成支付宝付款页…</n-alert>
        <n-alert v-else-if="stripeAlipayError" type="error">{{ stripeAlipayError }}</n-alert>
        <n-button v-if="stripeAlipayError" size="small" type="warning" :loading="stripeAlipayLoading" @click="runStripeAlipay(gotoProLink)">重试自动提交</n-button>
        <n-alert v-else-if="stripeAlipayPopupBlocked" type="warning">
          已提交 Alipay（{{ stripeAlipayAmountText }}）。浏览器拦截了弹窗，请点击下方按钮打开付款页。
        </n-alert>
        <n-alert v-else-if="stripeAlipayUrl" type="success">
          已提交 Alipay（{{ stripeAlipayAmountText }}）。付款页已在独立窗口打开，请直接扫码。
        </n-alert>
        <n-input v-if="stripeAlipayUrl" :value="stripeAlipayUrl" readonly type="textarea" :rows="2"
          style="font-size: 12px; word-break: break-all" />
        <n-space v-if="stripeAlipayUrl">
          <n-button size="small" @click="handleCopyStripeAlipay">复制链接</n-button>
          <n-button size="small" type="primary" @click="handleOpenStripeAlipay">打开支付宝付款页</n-button>
        </n-space>
        <n-alert v-if="stripePollStatus" :type="stripePollStatusType">{{ stripePollStatus }}</n-alert>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showGotoProModal = false">关闭</n-button>
          <n-button type="success" @click="handleGotoProSuccess">支付成功</n-button>
          <n-button type="error" @click="handleGotoProFail">支付失败</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 半价提链：填写 uid 与套餐 -->
    <n-modal v-model:show="showHalfPriceModal" title="半价提链" preset="dialog" positive-text="下一步"
      negative-text="取消" :positive-button-props="{ loading: halfPriceLoading, disabled: halfPriceLoading }"
      @positive-click="handleHalfPriceNext" style="width: 480px">
      <n-form label-placement="left" label-width="90px" style="margin-top: 20px">
        <n-form-item label="UID" required>
          <n-input v-model:value="halfPriceForm.uid" placeholder="请输入活动页 uid"
            @blur="loadHalfPriceQuota(halfPriceForm.uid)" />
        </n-form-item>
        <n-form-item label="当前余量">
          <span>{{ halfPriceQuota || '—' }}</span>
        </n-form-item>
        <n-form-item label="选择套餐" required>
          <n-radio-group v-model:value="halfPriceForm.tier">
            <n-space>
              <n-radio value="pro">Pro 半价注册</n-radio>
              <n-radio value="pro_plus">Pro Plus 半价注册</n-radio>
              <n-radio value="ultra">Ultra 半价注册</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 提链-支付成功：单卡升级成品弹窗 -->
    <n-modal v-model:show="showGotoProSuccessModal" title="支付成功 - 更新为成品" preset="dialog" positive-text="确认更新"
      negative-text="取消" @positive-click="handleGotoProSuccessSubmit" style="width: 600px">
      <n-form label-placement="left" label-width="120px" style="margin-top: 20px">
        <n-grid :cols="2" :x-gap="24" :y-gap="12">
          <n-form-item-gi label="订阅类型">
            <n-select v-model:value="upgradeForm.subscription_type" :options="subscriptionTypeOptions"
              placeholder="选择或输入订阅类型" filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅时间">
            <n-date-picker v-model:value="upgradeForm.subscription_time" type="datetime" placeholder="默认为当前时间"
              style="width: 100%" clearable />
          </n-form-item-gi>

          <n-form-item-gi label="订阅剩余天数">
            <n-input-number v-model:value="upgradeForm.subscription_remaining_days" :min="0" placeholder="默认30天"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买价格(追加)">
            <n-input-number v-model:value="upgradeForm.purchase_price" :min="0" :precision="2" placeholder="追加到已有价格"
              style="width: 100%" />
          </n-form-item-gi>

          <n-form-item-gi label="购买平台">
            <n-select v-model:value="upgradeForm.purchase_from" :options="purchasePlatformOptions"
              placeholder="选择或输入购买平台" filterable tag clearable />
          </n-form-item-gi>

          <n-form-item-gi label="购买时间" :span="2">
            <n-date-picker v-model:value="upgradeForm.purchase_date" type="datetime" placeholder="默认为当前时间"
              style="width: 100%" clearable />
          </n-form-item-gi>
        </n-grid>

        <n-alert type="info" style="margin-top: 12px">
          将为该账号更新为成品，订阅状态设置为"已订阅"、账号类型设置为"成品"。
        </n-alert>
      </n-form>
    </n-modal>

    <!-- 提链-支付失败：填写备注弹窗 -->
    <n-modal v-model:show="showGotoProFailModal" title="支付失败 - 填写备注" preset="dialog" positive-text="确认保存"
      negative-text="取消" @positive-click="handleGotoProFailSubmit" style="width: 480px">
      <n-form label-placement="left" label-width="80px" style="margin-top: 20px">
        <n-form-item label="备注原因">
          <n-input v-model:value="gotoProFailRemark" type="textarea" :rows="4" placeholder="请输入支付失败的备注原因" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 批量冻结/解冻弹窗 -->
    <n-modal v-model:show="showFreezeModal"
      :title="freezeAction === 1 ? `批量冻结 (${checkedRowKeys.length})` : `批量解冻 (${checkedRowKeys.length})`"
      preset="dialog"
      :positive-text="freezeAction === 1 ? '确认冻结' : '确认解冻'"
      negative-text="取消"
      @positive-click="handleFreezeSubmit"
      style="width: 480px">
      <n-form label-placement="left" label-width="80px" style="margin-top: 20px">
        <n-form-item v-if="freezeAction === 1" label="冻结备注">
          <n-input v-model:value="freezeRemark" type="textarea" :rows="3" placeholder="请输入冻结原因（选填）" />
        </n-form-item>
        <n-alert v-if="freezeAction === 1" type="warning" style="margin-top: 8px">
          冻结后该卡密将无法提链和更新为成品，请谨慎操作。
        </n-alert>
      </n-form>
    </n-modal>

    <!-- 已售列表：批量提链结果（账号----链接） -->
    <n-modal v-model:show="showBatchGotoProModal" title="批量提链结果" preset="card" style="width: 720px">
      <n-input :value="batchGotoProResultText" type="textarea" readonly :rows="14"
        placeholder="无展示链接（可能均为 dashboard：仅掉订阅会回滚，其余见下方说明）" style="font-family: monospace; font-size: 13px" />
      <template #footer>
        <n-space justify="end">
          <n-button @click="handleCopyBatchGotoProResult">复制全部</n-button>
          <n-button type="primary" @click="showBatchGotoProModal = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted, onUnmounted, watch } from 'vue'
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
  batchDashboardGotoResolve,
  batchEnableOnDemandSpend,
  gotoProUpgrade,
  submitStripeAlipay,
  pollCardSubscription,
  halfPriceCheckout,
  getHalfPriceQuota,
  updateCardRemark,
  batchFreezeCards,
  batchDeleteCards,
  type Card,
  type CardRequest,
} from '@/api/card'
import { getDigisellerPrices, type DigisellerPrice } from '@/api/digiseller'

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

// 仅 cards_cursor 才带出 session-token / Auto-Login 说明字段
const isCursorCategory = computed(() => category.value === 'cursor')
const CURSOR_AUTO_LOGIN_URL = 'https://docs.aiguoguo199.com/doc-9320337'

const sessionTokenField = (card: Card, sep: string): string => {
  if (!isCursorCategory.value) return ''
  const token = (card.token || '').trim()
  return token ? `${sep}session-token: ${token}` : ''
}

const cursorAutoLoginField = (sep: string): string => {
  if (!isCursorCategory.value) return ''
  return `${sep}Account restrictions triggering SMS verification have been frequent recently; logging in via token is strongly recommended. Please refer to the following: ${CURSOR_AUTO_LOGIN_URL}`
}

const hasPhoneReceive = (card: Card): boolean => {
  return !!(card.phone || '').trim() && !!(card.phone_link || '').trim()
}

const phoneFields = (card: Card, sep: string): string => {
  if (!hasPhoneReceive(card)) return ''
  return `${sep}phone: ${(card.phone || '').trim()}${sep}phone-login: ${(card.phone_link || '').trim()}`
}

const phoneDashSuffix = (card: Card): string => {
  if (!hasPhoneReceive(card)) return ''
  return `----${(card.phone || '').trim()}----${(card.phone_link || '').trim()}`
}

const phoneDashTitle = (card: Card): string => {
  return hasPhoneReceive(card) ? '----手机号----短信接码地址' : ''
}

// 国内格式固定追加两列（空值也占位）
const domesticPhoneTitle = '----手机号----短信接码地址'
const domesticPhoneSuffix = (card: Card): string => {
  return `----${(card.phone || '').trim()}----${(card.phone_link || '').trim()}`
}

const formatDigiseller = (card: Card): string => {
  return `account: ${card.account}\npass: ${card.password || ''}\nmail-pass: ${card.mail_password || ''}${sessionTokenField(card, '\n')}\n\nmail-login: ${card.mail_url || ''}${cursorAutoLoginField('\n')}${phoneFields(card, '\n')}`
}

const formatDigisellerAuto = (card: Card): string => {
  return `account: ${card.account}<br>pass: ${card.password || ''}<br>mail-pass: ${card.mail_password || ''}${sessionTokenField(card, '<br>')}<br>mail-login: ${card.mail_url || ''}${cursorAutoLoginField('<br>')}${phoneFields(card, '<br>')}<br>Если вам удобно, не могли бы вы оставить нам хороший отзыв? https://ibb.co/tTgSNRLP<br>Подписывайтесь на наш канал, чтобы получать больше выгодных предложений: https://t.me/AI_GUO_GUO`
}

// 取货卡密信息格式化
const pickupCardInfo = computed(() => {
  if (!pickedCard.value) return ''

  const card = pickedCard.value
  if (pickupForm.value.format === 'digiseller') {
    // 密码和邮箱密码均为空时，使用邮箱验证码登录格式
    if (!card.password && !card.mail_password) {
      return `Пожалуйста, войдите в систему, используя код подтверждения, отправленный на электронную почту:

${card.account}

mail-login: ${card.mail_url || ''}${phoneFields(card, '\n')}

Пожалуйста, выполните следующие шаги заново:
1. Введите аккаунт: ${card.account}
2. Нажмите «Далее»
3. Нажмите кнопку: «Email sign-in code»`
    }
    // 常规 digiseller 订阅格式
    return formatDigiseller(card)
  } else if (pickupForm.value.format === 'reverse') {
    // 逆向格式
    return `账号----token${phoneDashTitle(card)}
${card.account}----${card.token || ''}${phoneDashSuffix(card)}`
  } else {
    // 国内订阅格式：账号----密码----邮箱密码----token----手机号----短信接码地址
    return `账号----密码----邮箱密码----token${domesticPhoneTitle}
${card.account}----${card.password || ''}----${card.mail_password || ''}----${card.token || ''}${domesticPhoneSuffix(card)}`
  }
})

// 状态
const loading = ref(false)
const showModal = ref(false)
const showBatchModal = ref(false)
const isEdit = ref(false)
const cardList = ref<Card[]>([])
const formRef = ref<FormInst | null>(null)
const searchAccountsText = ref('')
const searchSubscriptionType = ref('')
const searchSubscriptionStatus = ref(0)
const searchSellTo = ref('')
const searchPurchaseBy = ref('')
const searchIsCheck = ref(0)
const searchFreezeStatus = ref(0)
// 购买日期 / 冻结时间筛选（n-date-picker 返回毫秒）
const searchPurchaseDate = ref<number | null>(null)
const searchFreezeTime = ref<number | null>(null)
const batchCheckLoading = ref(false)
const batchOnDemandLoading = ref(false)
const batchImportText = ref('')

// 普号列表批量勾选
const checkedRowKeys = ref<number[]>([])

// 导出相关
const showExportModal = ref(false)
const exportLoading = ref(false)
const exportMode = ref<'selected' | 'filter'>('filter')
const exportSelectedFields = ref<string[]>(['account', 'code_method'])
// 导出格式：digiseller / digiseller_auto 走与复制相同的多行格式，其余走 ---- 字段拼接
const exportFormatMode = ref<'fields' | 'digiseller' | 'digiseller_auto'>('fields')

// 打开导出面板：有勾选时默认「导出勾选记录」
const handleOpenExportModal = () => {
  exportMode.value = checkedRowKeys.value.length > 0 ? 'selected' : 'filter'
  showExportModal.value = true
}

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
  { label: '接码方式', value: 'code_method' },
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
  { label: 'Client ID', value: 'client_id' },
  { label: '2FA', value: '2fa' },
  { label: '手机号', value: 'phone' },
  { label: '短信接码地址', value: 'phone_link' },
  { label: '备注', value: 'remark' },
]

const exportFormatHint = computed(() => {
  if (exportFormatMode.value === 'fields') {
    return '选择导出字段（顺序即为列顺序，分隔符固定为 ----）'
  }
  return isCursorCategory.value
    ? '当前为 Digiseller 专用格式（cards_cursor 有 token 时带出 session-token，并附加 Auto-Login）'
    : '当前为 Digiseller 专用格式'
})

// 格式预览：Digiseller 预设显示实际导出模板，其余为 ---- 字段拼接
const exportPreview = computed(() => {
  const cursorHint = isCursorCategory.value ? ' / session-token（有 token 才带出）' : ''
  const autoLoginHint = isCursorCategory.value ? ' / Auto-Login' : ''
  if (exportFormatMode.value === 'digiseller') {
    return `account / pass / mail-pass${cursorHint} / mail-login${autoLoginHint} / phone / phone-login（有则带出）`
  }
  if (exportFormatMode.value === 'digiseller_auto') {
    const autoCursor = isCursorCategory.value ? '<br>session-token（有 token 才带出）' : ''
    const autoLogin = isCursorCategory.value ? '<br>Auto-Login' : ''
    return `account<br>pass<br>mail-pass${autoCursor}<br>mail-login${autoLogin}<br>phone / phone-login（有则带出）<br>评价引导<br>频道订阅`
  }
  return exportSelectedFields.value
    .map(v => {
      if (v === 'code_method') return '接码链接----账号----邮箱密码'
      return exportFieldOptions.find(o => o.value === v)?.label ?? v
    })
    .join('----')
})

const handleExportFieldsUpdate = (val: Array<string | number>) => {
  exportSelectedFields.value = val.map(String)
  exportFormatMode.value = 'fields'
}

// 快速预设（与列表复制 / 批量取货格式对应）
const applyExportPreset = (preset: string) => {
  switch (preset) {
    case 'digiseller':
    case 'digiseller_auto':
      exportFormatMode.value = preset
      exportSelectedFields.value = isCursorCategory.value
        ? ['account', 'password', 'mail_password', 'token', 'mail_url']
        : ['account', 'password', 'mail_password', 'mail_url']
      break
    case 'domestic':
      // 国内格式：账号----密码----邮箱密码----token----手机号----短信接码地址
      exportFormatMode.value = 'fields'
      exportSelectedFields.value = ['account', 'password', 'mail_password', 'token', 'phone', 'phone_link']
      break
    case 'reverse':
      // 逆向格式：账号----token
      exportFormatMode.value = 'fields'
      exportSelectedFields.value = ['account', 'token']
      break
    case 'code_method':
      // 接码方式：账号----接码链接----账号----邮箱密码
      exportFormatMode.value = 'fields'
      exportSelectedFields.value = ['account', 'code_method']
      break
    case 'all':
      exportFormatMode.value = 'fields'
      exportSelectedFields.value = exportFieldOptions.map(o => o.value)
      break
    case 'clear':
      exportFormatMode.value = 'fields'
      exportSelectedFields.value = []
      break
  }
}

// 批量升级为成品弹窗
const showUpgradeModal = ref(false)
const upgradeForm = ref({
  subscription_type: '',
  subscription_time: Date.now() as number | undefined,   // 毫秒（n-date-picker）
  subscription_remaining_days: 30 as number | undefined,
  purchase_price: undefined as number | undefined,
  purchase_from: '支付宝',
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
  format: 'digiseller' as 'digiseller' | 'digiseller_auto' | 'domestic' | 'reverse' | 'code_method',
  sell_price: 20 as number | undefined,
  sell_to: 'Digiseller',
})

// 取货相关状态
const showPickupModal = ref(false)
const pickupStep = ref(1) // 1: 选择条件, 2: 预览确认
const unsoldSubscriptionTypes = ref<{ label: string; value: string }[]>([])
const pickedCard = ref<Card | null>(null)
// Digiseller价格配置缓存
const digisellerPrices = ref<DigisellerPrice[]>([])

// 提链相关状态
const gotoProLoading = ref<Record<number, boolean>>({})
const showGotoProModal = ref(false)
const gotoProLink = ref('')
const gotoProCardId = ref<number>(0)
const gotoProAccount = ref('')
const gotoProSubscriptionType = ref('')
const showGotoProSuccessModal = ref(false)
const showGotoProFailModal = ref(false)
const gotoProFailRemark = ref('')
const stripeAlipayLoading = ref(false)
const stripeAlipayUrl = ref('')
const stripeAlipayAmountText = ref('')
const stripeAlipayError = ref('')
const stripeAlipayPopupBlocked = ref(false)
const stripeAlipaySigned = ref(false)
const stripeAlipayCompleted = ref(false)
const stripePollStatus = ref('')
const stripePollStatusType = ref<'default' | 'info' | 'success' | 'warning' | 'error'>('info')
const showHalfPriceModal = ref(false)
const halfPriceLoading = ref(false)
const halfPriceCard = ref<Card | null>(null)
const halfPriceForm = ref({
  uid: '',
  tier: 'pro' as 'pro' | 'pro_plus' | 'ultra',
})
const HALF_PRICE_UID_COOKIE = 'cursor_half_price_uid'
const halfPriceQuota = ref('')

// 已售列表：批量提链结果弹窗
const showBatchGotoProModal = ref(false)
const batchGotoProResultText = ref('')
const batchGotoProLoading = ref(false)

// 未售列表：订阅状态 -2 标签点击复制（俄语说明）
const CURSOR_PRO_DELAY_RU_TEXT =
  'Подписка Cursor иногда обновляется с задержкой. Вы можете ещё раз нажать кнопку обновления подписки. Если после обновления страницы отображается Pro — всё в порядке.\n\nЕсли по-прежнему отображается Free или вас сразу перенаправляет на страницу оплаты — пожалуйста, свяжитесь со мной для решения проблемы.'

// 批量冻结/解冻弹窗
const showFreezeModal = ref(false)
const freezeAction = ref<1 | -1>(1)
const freezeRemark = ref('')
// 提链订阅类型选项（下拉）
const gotoProTypeOptions = [
  { label: 'Pro', value: 'pro' },
  { label: 'Pro+', value: 'pro_plus' },
  { label: 'Ultra', value: 'ultra' },
]
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
  purchase_from: '支付宝',
  purchase_by: '',
  purchase_date: undefined as number | undefined,
  mail_url: 'https://login.live.com',
  sell_status: 1,
  subscription_status: 1,
  account_type: 2,
  code_link: 'https://tool.toolsvip.cc/easy-mailbox/frontend',
  phone_link: '',
  remark: '',
  field_mapping: {
    account: 1,
    password: 2,
    mail_password: 3,
    subscription_type: 0,
    subscription_time: 0,
    subscription_expired_time: 0,
    token: 0,
    client_id: 0,
    api_key: 0,
    '2fa': 0,
    phone: 0,
    phone_link: 0,
  },
})

// 分页
const pagination = ref<PaginationProps>({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100, 500, 1000],
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
  client_id: '',
  mail_url: '',
  remark: '',
  code_link: '',
  phone: '',
  phone_link: '',
  sell_order_no: '',
  subscription_credits: undefined as number | undefined,
  subscription_time: undefined as number | undefined,
  subscription_expired_time: undefined as number | undefined,
  purchase_date: undefined as number | undefined,
  sell_date: undefined as number | undefined,
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
  { label: '已订阅需点击pro', value: -2 },
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
  { label: '封禁', value: -2 },
  { label: '正常', value: 1 },
  { label: '禁用', value: 2 },
]

// 订阅类型选项（支持手动输入）
const subscriptionTypeOptions = [
  { label: 'Free', value: 'free' },
  { label: 'Pro', value: 'pro' },
  { label: 'Pro+', value: 'pro_plus' },
  { label: 'Pro x5', value: 'pro_x5' },
  { label: 'Pro x20', value: 'pro_x20' },
  { label: 'Ultra', value: 'ultra' },
  { label: 'Go', value: 'go' },
  { label: 'Plus', value: 'plus' },
  { label: 'Team', value: 'team' },
]

// 购买平台选项（支持手动输入）
const purchasePlatformOptions = [
  { label: '支付宝', value: '支付宝' },
  { label: '微信', value: '微信' },
  { label: 'Telegram', value: 'Telegram' },
  { label: '闲鱼', value: '闲鱼' },
  { label: '淘宝', value: '淘宝' },
  { label: '卡充', value: '卡充' },
]

// 邮箱地址选项（支持手动输入）
const mailUrlOptions = [
  { label: 'https://login.live.com', value: 'https://login.live.com' },
  { label: 'https://mail.com', value: 'https://mail.com' },
  { label: 'https://gmx.us', value: 'https://gmx.us' },
  { label: 'https://gmail.com', value: 'https://gmail.com' },
]

const codeLinkOptions = [
  { label: 'https://tool.toolsvip.cc/easy-mailbox/frontend', value: 'https://tool.toolsvip.cc/easy-mailbox/frontend' },
  { label: 'https://www.xckj.site/easy-mailbox/frontend/', value: 'https://www.xckj.site/easy-mailbox/frontend/' },
  { label: 'https://ms.lqqq.cc/web/', value: 'https://ms.lqqq.cc/web/' },
  { label: 'https://emails.520952.xyz/', value: 'https://emails.520952.xyz/' },
]

const getSubscriptionTypeTagType = (subscriptionType?: string): 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' => {
  const t = (subscriptionType || '').toLowerCase()
  const map: Record<string, 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'> = {
    pro: 'success',
    pro_plus: 'warning',
    ultra: 'error',
    go: 'info',
    plus: 'primary',
    team: 'default',
  }
  return map[t] ?? 'info'
}

const getSubscriptionStatusTagInfo = (subscriptionStatus?: number): { label: string; type: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' } => {
  const map: Record<number, { label: string; type: 'default' | 'primary' | 'info' | 'success' | 'warning' | 'error' }> = {
    1: { label: '已订阅', type: 'success' },
    2: { label: '未订阅', type: 'warning' },
    [-1]: { label: '掉订阅', type: 'error' },
    [-2]: { label: '已订阅需点击pro', type: 'warning' },
  }
  const v = subscriptionStatus ?? 0
  return map[v] ?? { label: v ? String(v) : '—', type: 'default' }
}

// 格式化 Unix 时间戳为可读时间
const formatTimestamp = (ts?: number | null): string => {
  if (!ts) return '—'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/** 根据订阅类型推导订阅状态与账号类型 */
const resolveSubscriptionMeta = (subscriptionType?: string) => {
  const type = (subscriptionType || '').trim()
  if (!type) {
    return {
      subscription_type: undefined as string | undefined,
      subscription_status: batchConfig.value.subscription_status,
      account_type: batchConfig.value.account_type,
    }
  }
  const isFree = type.toLowerCase() === 'free'
  return {
    subscription_type: type,
    subscription_status: isFree ? 2 : 1,
    account_type: isFree ? 1 : 2,
  }
}

/** 批量导入时解析时间：支持 Unix 秒/毫秒、YYYY-MM-DD HH:mm:ss */
const parseImportTimestamp = (raw: string): number | undefined => {
  const s = raw.trim()
  if (!s) return undefined
  if (/^\d+$/.test(s)) {
    const n = Number(s)
    if (s.length >= 13) return Math.floor(n / 1000)
    return n
  }
  const d = new Date(s.replace(' ', 'T'))
  if (!Number.isNaN(d.getTime())) return Math.floor(d.getTime() / 1000)
  return undefined
}

// 格式化 time.Time JSON（RFC3339 / 常见格式）为可读时间
const formatDateTimeString = (val?: string | null): string => {
  if (!val) return '—'
  const d = new Date(val)
  if (Number.isNaN(d.getTime())) return val
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const formatYYYYMMDD = (ms: number) => {
  const d = new Date(ms)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// 解析多行账号搜索文本（按行分隔，去重）
const parseSearchAccounts = (text: string): string[] => {
  const seen = new Set<string>()
  const accounts: string[] = []
  for (const line of text.split(/\r?\n/)) {
    const account = line.trim()
    if (!account || seen.has(account)) continue
    seen.add(account)
    accounts.push(account)
  }
  return accounts
}

const buildCardQueryParams = () => {
  const purchaseDateParam =
    searchPurchaseDate.value != null ? formatYYYYMMDD(searchPurchaseDate.value) : ''
  const accounts = parseSearchAccounts(searchAccountsText.value)
  return {
    category: category.value,
    type: cardType.value,
    ...(accounts.length > 0 ? { accounts: accounts.join('\n') } : {}),
    ...(searchSubscriptionType.value ? { subscription_type: searchSubscriptionType.value } : {}),
    ...(searchSubscriptionStatus.value !== 0 ? { subscription_status: searchSubscriptionStatus.value } : {}),
    ...(searchSellTo.value.trim() ? { sell_to: searchSellTo.value.trim() } : {}),
    ...(searchPurchaseBy.value.trim() ? { purchase_by: searchPurchaseBy.value.trim() } : {}),
    ...(searchIsCheck.value !== 0 ? { is_check: searchIsCheck.value } : {}),
    ...(purchaseDateParam ? { purchase_date: purchaseDateParam } : {}),
    ...(searchFreezeStatus.value !== 0 ? { freeze_status: searchFreezeStatus.value } : {}),
    ...(searchFreezeTime.value != null ? { freeze_time: formatYYYYMMDD(searchFreezeTime.value) } : {}),
  }
}

// 表格列定义（computed，根据列表类型展示不同列）
const columns = computed<DataTableColumns<Card>>(() => {
  const isSold = cardType.value === 'sold'
  const isAll = cardType.value === 'all'
  const isUnsold = cardType.value === 'unsold'

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
      render: (row: Card) => {
        // 未售/已售：账号单独一行，标签换行在下面一行
        if (!isAll) {
          const tags: any[] = []
          if (row.subscription_type) {
            tags.push(
              h(
                NTag,
                { type: getSubscriptionTypeTagType(row.subscription_type), size: 'small', bordered: false },
                { default: () => row.subscription_type }
              )
            )
          }
          const statusInfo = getSubscriptionStatusTagInfo(row.subscription_status)
          if (statusInfo.label !== '—') {
            if (isUnsold && row.subscription_status === -2) {
              tags.push(
                h(
                  NTag,
                  {
                    type: statusInfo.type,
                    size: 'small',
                    bordered: false,
                    style: { cursor: 'pointer' },
                    title: '点击复制俄语说明',
                    onClick: async () => {
                      try {
                        await navigator.clipboard.writeText(CURSOR_PRO_DELAY_RU_TEXT)
                        message.success('已复制到剪贴板')
                      } catch {
                        message.error('复制失败')
                      }
                    },
                  },
                  { default: () => statusInfo.label }
                )
              )
            } else {
              tags.push(
                h(
                  NTag,
                  { type: statusInfo.type, size: 'small', bordered: false },
                  { default: () => statusInfo.label }
                )
              )
            }
          }
          if (tags.length > 0) {
            return h(
              'div',
              {
                style: {
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '6px',
                  lineHeight: '1.2',
                },
              },
              [
                h('div', { style: { fontWeight: 500 } }, row.account),
                h(NSpace, { size: 6, wrap: true }, { default: () => tags }),
              ]
            )
          }
        }
        return h('span', row.account)
      },
    },
    // 普号列表显示密码、邮箱密码和备注
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
      {
        title: '备注',
        key: 'remark',
        width: 160,
        render: (row: Card) => row.remark || '—',
      },
    ] as DataTableColumns<Card> : []),
    // 订阅类型、订阅状态：按需求不再单独成列（未售/已售展示在账号后面）
    ...((isSold || isUnsold) ? [{
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
    // 普号列表：删除订阅时间/剩余天数，改为显示创建时间
    ...(isAll ? [{
      title: '创建时间',
      key: 'created_at',
      width: 170,
      render: (row: Card) => formatDateTimeString(row.created_at),
    }] as DataTableColumns<Card> : []),
    // 普号列表：冻结时间、冻结备注
    ...(isAll ? [
      {
        title: '冻结时间',
        key: 'freeze_time',
        width: 170,
        render: (row: Card) => formatTimestamp(row.freeze_time),
      },
      {
        title: '冻结备注',
        key: 'freeze_remark',
        width: 160,
        render: (row: Card) => row.freeze_remark || '—',
      },
    ] as DataTableColumns<Card> : []),
    // 未售/已售：购买时间（用 purchase_date）
    ...(!isAll ? [{
      title: '购买时间',
      key: 'purchase_date',
      width: 170,
      render: (row: Card) => formatTimestamp(row.purchase_date),
    }] as DataTableColumns<Card> : []),
    // 未售列表：订阅到期时间
    ...(isUnsold ? [{
      title: '订阅到期时间',
      key: 'subscription_expired_time',
      width: 170,
      render: (row: Card) => formatTimestamp(row.subscription_expired_time),
    }] as DataTableColumns<Card> : []),
    // 未售/已售：订阅额度（美元）
    ...((isUnsold || isSold) ? [{
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
          1: { label: '检查成功', type: 'success' },
          2: { label: '检查失败', type: 'error' },
        }
        const val = row.is_check ?? -1
        const info = map[val] ?? { label: String(val), type: 'default' as const }
        return h(NTag, { type: info.type, size: 'small' }, { default: () => info.label })
      },
    }] as DataTableColumns<Card> : []),
    // 已售列表显示出售对方、出售时间
    ...(isSold ? [
      {
        title: '出售对方',
        key: 'sell_to',
        width: 120,
        render: (row: Card) => row.sell_to || '—',
      },
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
    { label: 'Digiseller自动发货', key: 'digiseller_auto' },
    { label: '国内格式', key: 'domestic' },
    { label: '逆向格式', key: 'reverse' },
    { label: '接码方式', key: 'code_method' },
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

      // 普号/已售列表中，token 存在时显示"提链"下拉按钮
      // 普号列表：冻结时禁用；已售列表：始终可用
      if ((cardType.value === 'all' || cardType.value === 'sold') && row.token) {
        const isFrozen = cardType.value === 'all' && row.freeze_status === 1
        buttons.push(
          h(
            NDropdown,
            {
              trigger: 'click',
              options: [
                { label: '半价提链', key: 'half_price' },
                ...gotoProTypeOptions.map(opt => ({ label: opt.label, key: opt.value })),
              ],
              onSelect: (key: string) => {
                if (isFrozen) return
                if (key === 'half_price') {
                  handleOpenHalfPrice(row)
                  return
                }
                handleGotoProWithType(row, key)
              },
              disabled: isFrozen,
            },
            {
              default: () => h(
                NButton,
                {
                  size: 'small',
                  type: isFrozen ? 'default' : 'warning',
                  loading: !!gotoProLoading.value[row.id],
                  disabled: isFrozen,
                  title: isFrozen ? '已冻结，无法提链' : undefined,
                },
                { default: () => isFrozen ? '已冻结' : '提链' }
              ),
            }
          )
        )
      }

      // 所有列表中，邮件接码或短信接码有地址时显示下拉
      if (row.code_link || row.phone_link) {
        buttons.push(
          h(
            NDropdown,
            {
              trigger: 'click',
              options: [
                { label: '邮件接码', key: 'mail', disabled: !row.code_link },
                { label: '短信接码', key: 'sms', disabled: !row.phone_link },
              ],
              onSelect: (key: string) => {
                if (key === 'mail') {
                  if (!row.code_link) {
                    message.warning('未配置邮件接码链接')
                    return
                  }
                  window.open(`${row.code_link}/${row.account}----${row.mail_password || ''}`, '_blank')
                  return
                }
                if (!row.phone_link) {
                  message.warning('未配置短信接码地址')
                  return
                }
                window.open(row.phone_link, '_blank')
              },
            },
            {
              default: () => h(
                NButton,
                { size: 'small', type: 'info' },
                { default: () => '接码' }
              ),
            }
          )
        )
      }

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
      ...buildCardQueryParams(),
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
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

// 批量开启按需付费
const handleBatchEnableOnDemand = async () => {
  const ids = checkedRowKeys.value as number[]
  if (ids.length === 0) {
    message.warning('请先勾选要操作的记录')
    return
  }
  batchOnDemandLoading.value = true
  try {
    const response = await batchEnableOnDemandSpend(category.value, ids)
    if (response.code === 200) {
      message.success(response.message || '操作完成')
    } else {
      message.error(response.message || '操作失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '操作失败')
  } finally {
    batchOnDemandLoading.value = false
  }
}

// 提链返回链接规范化（与后端一致：可能为 JSON 字符串）
const normalizeGotoLinkFromResponse = (raw: string): string => {
  let s = String(raw || '').trim()
  try {
    const parsed = JSON.parse(s) as unknown
    if (typeof parsed === 'string') s = parsed.trim()
  } catch {
    /* 非 JSON */
  }
  return s.replace(/\/+$/, '')
}

const isCursorDashboardLink = (raw: string): boolean =>
  normalizeGotoLinkFromResponse(raw) === 'https://cursor.com/dashboard'

const mapSubscriptionToGotoProType = (subscriptionType?: string): string => {
  const t = (subscriptionType || '').toLowerCase()
  if (t === 'pro' || t === 'pro_plus' || t === 'ultra') return t
  return 'pro'
}

// 已售列表：批量提链（展示 账号----链接；dashboard 且掉订阅(-1) 才回滚+标记 -2）
const handleBatchGotoPro = async () => {
  const ids = checkedRowKeys.value as number[]
  if (ids.length === 0) {
    message.warning('请先勾选记录')
    return
  }
  batchGotoProLoading.value = true
  const lines: string[] = []
  const dashboardIds: number[] = []
  const errLines: string[] = []

  try {
    for (const id of ids) {
      const row =
        selectedCardsMap.value.get(id) ?? cardList.value.find(c => c.id === id)
      if (!row?.token) {
        errLines.push(`${row?.account || `#${id}`}: 无 Token，已跳过`)
        continue
      }
      const subType = mapSubscriptionToGotoProType(row.subscription_type)
      try {
        const response = await gotoProUpgrade(row.token, subType)
        if (response.code !== 200 || !response.data) {
          errLines.push(`${row.account}: ${response.message || '获取链接失败'}`)
          continue
        }
        const link = normalizeGotoLinkFromResponse(String(response.data))
        if (isCursorDashboardLink(String(response.data))) {
          if (row.subscription_status !== -1) {
            errLines.push(
              `${row.account}: 提链结果为 dashboard，仅「掉订阅」可回滚，当前状态不可回滚`
            )
          } else {
            dashboardIds.push(row.id)
          }
        } else {
          lines.push(`${row.account}----${link}`)
        }
      } catch (e: any) {
        errLines.push(`${row.account}: ${e.response?.data?.message || e.message || '请求失败'}`)
      }
    }

    let dashboardResultMessage = ''
    let dashboardBatchOk = false
    if (dashboardIds.length > 0) {
      const res = await batchDashboardGotoResolve({
        category: category.value,
        ids: dashboardIds,
      })
      if (res.code !== 200) {
        errLines.push(`Dashboard 批量处理: ${res.message || '失败'}`)
      } else {
        dashboardBatchOk = true
        dashboardResultMessage = res.message || ''
      }
    }

    const parts: string[] = []
    if (lines.length) {
      parts.push('--- 账号----链接 ---')
      parts.push(...lines)
    }
    if (errLines.length) {
      parts.push('--- 异常或跳过 ---')
      parts.push(...errLines)
    }
    if (dashboardIds.length) {
      parts.push(
        dashboardResultMessage
          ? `--- ${dashboardResultMessage} ---`
          : `--- 已为 ${dashboardIds.length} 条 dashboard（掉订阅）执行回滚并标记「已订阅需点击pro」---`
      )
    }
    batchGotoProResultText.value = parts.join('\n') || '（无链接行输出，请查看上方说明）'
    showBatchGotoProModal.value = true

    const ok = lines.length + (dashboardBatchOk ? dashboardIds.length : 0)
    message.success(
      ok > 0
        ? `处理完成：展示链接 ${lines.length} 条，dashboard 回滚/标记 ${dashboardIds.length} 条（仅掉订阅可回滚）`
        : '处理结束（请查看结果中的错误信息）'
    )
    checkedRowKeys.value = []
    selectedCardsMap.value.clear()
    await loadCards()
  } finally {
    batchGotoProLoading.value = false
  }
}

const handleCopyBatchGotoProResult = async () => {
  try {
    await navigator.clipboard.writeText(batchGotoProResultText.value)
    message.success('已复制')
  } catch {
    message.error('复制失败')
  }
}

// 提链：获取 Cursor Pro 付款链接
// 提链：选择订阅类型后直接发起请求
const handleGotoProWithType = async (row: Card, subscriptionType: string) => {
  if (!row.token) return
  gotoProLoading.value[row.id] = true
  try {
    const response = await gotoProUpgrade(row.token, subscriptionType)
    if (response.code === 200 && response.data) {
      gotoProCardId.value = row.id
      gotoProAccount.value = row.account || ''
      gotoProSubscriptionType.value = subscriptionType
      gotoProLink.value = response.data
      showGotoProModal.value = true
      runStripeAlipay(response.data)
    } else {
      message.error(response.message || '获取付款链接失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '获取付款链接失败')
  } finally {
    gotoProLoading.value[row.id] = false
  }
}

const readCookie = (name: string): string => {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : ''
}

const writeCookie = (name: string, value: string) => {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=31536000`
}

const handleOpenHalfPrice = (row: Card) => {
  if (!row.token) return
  halfPriceCard.value = row
  halfPriceForm.value = {
    uid: readCookie(HALF_PRICE_UID_COOKIE),
    tier: 'pro',
  }
  showHalfPriceModal.value = true
  loadHalfPriceQuota(halfPriceForm.value.uid)
}

const loadHalfPriceQuota = async (uid: string) => {
  const id = uid.trim()
  if (!id) {
    halfPriceQuota.value = ''
    return
  }
  try {
    const res = await getHalfPriceQuota(id)
    if (res.code === 200 && res.data) {
      halfPriceQuota.value = res.data
    } else {
      halfPriceQuota.value = ''
    }
  } catch {
    // 余量读取失败不影响提链
  }
}

let halfPriceQuotaTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => halfPriceForm.value.uid,
  (uid) => {
    if (!showHalfPriceModal.value) return
    if (halfPriceQuotaTimer) clearTimeout(halfPriceQuotaTimer)
    halfPriceQuotaTimer = setTimeout(() => loadHalfPriceQuota(uid), 400)
  }
)

const handleHalfPriceNext = async () => {
  const row = halfPriceCard.value
  const uid = halfPriceForm.value.uid.trim()
  if (!row?.token) {
    message.error('当前卡密没有 Token')
    return false
  }
  if (!uid) {
    message.warning('请输入 uid')
    return false
  }

  writeCookie(HALF_PRICE_UID_COOKIE, uid)
  halfPriceLoading.value = true
  gotoProLoading.value[row.id] = true
  try {
    const response = await halfPriceCheckout({
      uid,
      token: row.token,
      tier: halfPriceForm.value.tier,
    })
    if (response.code === 200 && response.data) {
      gotoProCardId.value = row.id
      gotoProAccount.value = row.account || ''
      gotoProSubscriptionType.value = halfPriceForm.value.tier
      gotoProLink.value = response.data
      showHalfPriceModal.value = false
      showGotoProModal.value = true
      runStripeAlipay(response.data)
      return true
    }
    message.error(response.message || '半价提链失败')
    return false
  } catch (error: any) {
    message.error(error.response?.data?.message || '半价提链失败')
    return false
  } finally {
    halfPriceLoading.value = false
    gotoProLoading.value[row.id] = false
  }
}

const handleCopyGotoProLink = async () => {
  try {
    await navigator.clipboard.writeText(gotoProLink.value)
    message.success('已复制到剪贴板')
  } catch {
    message.error('复制失败，请手动复制')
  }
}

const handleCopyStripeAlipay = async () => {
  try {
    await navigator.clipboard.writeText(stripeAlipayUrl.value)
    message.success('已复制支付宝付款链接')
  } catch {
    message.error('复制失败，请手动复制')
  }
}

const ALIPAY_PAY_WINDOW = 'stripe_alipay_pay'
let stripeAlipayWindow: Window | null = null

const closeAlipayPayWindow = () => {
  if (stripeAlipayWindow && !stripeAlipayWindow.closed) {
    stripeAlipayWindow.close()
  }
  stripeAlipayWindow = null
}

const openAlipayPayWindow = (url: string) => {
  if (!url) return
  if (stripeAlipayWindow && !stripeAlipayWindow.closed) {
    stripeAlipayWindow.location.replace(url)
    stripeAlipayWindow.focus()
    stripeAlipayPopupBlocked.value = false
    return
  }
  stripeAlipayWindow = window.open(url, ALIPAY_PAY_WINDOW, 'popup=yes,width=480,height=760,left=80,top=80')
  stripeAlipayPopupBlocked.value = !stripeAlipayWindow
  if (!stripeAlipayWindow) {
    message.warning('浏览器拦截了弹窗，请点击「打开支付宝付款页」')
  }
}

const runStripeAlipay = async (link: string) => {
  const checkoutUrl = (link || '').trim()
  stopSubscriptionPoll()
  if (!checkoutUrl.includes('checkout.stripe.com')) {
    stripeAlipayError.value = '当前不是 Stripe 结账链接，已跳过自动提交'
    stripeAlipayUrl.value = ''
    stripeAlipayAmountText.value = ''
    stripeAlipayPopupBlocked.value = false
    return
  }
  stripeAlipayLoading.value = true
  stripeAlipayUrl.value = ''
  stripeAlipayAmountText.value = ''
  stripeAlipayError.value = ''
  stripeAlipayPopupBlocked.value = false
  stripeAlipaySigned.value = false
  stripeAlipayCompleted.value = false
  stripePollStatus.value = ''
  try {
    const response = await submitStripeAlipay(checkoutUrl)
    if (response.code === 200 && response.data?.alipay_url) {
      stripeAlipayUrl.value = response.data.alipay_url
      openAlipayPayWindow(response.data.alipay_url)
      const amount = (response.data.amount || 0) / 100
      const currency = (response.data.currency || 'usd').toUpperCase()
      stripeAlipayAmountText.value = `${currency} ${amount.toFixed(2)}`
      message.success('Alipay 已提交，请在弹出窗口扫码')
      startSubscriptionPoll()
      return
    }
    stripeAlipayError.value = response.message || '自动提交 Alipay 失败'
  } catch (error: any) {
    stripeAlipayError.value = error.response?.data?.message || '自动提交 Alipay 失败'
  } finally {
    stripeAlipayLoading.value = false
  }
}

const handleOpenStripeAlipay = () => {
  if (!stripeAlipayUrl.value) return
  openAlipayPayWindow(stripeAlipayUrl.value)
}

let subscriptionPollTimer: ReturnType<typeof setInterval> | null = null
let subscriptionWaitTimer: ReturnType<typeof setInterval> | null = null
let subscriptionWaitSeconds = 0
let subscriptionWaitMembership = 'free'

const stopSubscriptionPoll = () => {
  if (subscriptionPollTimer) {
    clearInterval(subscriptionPollTimer)
    subscriptionPollTimer = null
  }
  if (subscriptionWaitTimer) {
    clearInterval(subscriptionWaitTimer)
    subscriptionWaitTimer = null
  }
}

const renderWaitStatus = (failed = false) => {
  if (stripeAlipayCompleted.value) {
    stripePollStatusType.value = 'success'
    stripePollStatus.value = `授权已完成，等待订阅生效… ${subscriptionWaitSeconds} 秒`
    return
  }
  if (stripeAlipaySigned.value) {
    stripePollStatusType.value = 'info'
    stripePollStatus.value = `授权已确认，正在完成支付… ${subscriptionWaitSeconds} 秒`
    return
  }
  stripePollStatusType.value = failed ? 'warning' : 'info'
  stripePollStatus.value = failed
    ? `等待付款中 ${subscriptionWaitSeconds} 秒，查询失败，继续等待…`
    : `等待付款中 ${subscriptionWaitSeconds} 秒，当前：${subscriptionWaitMembership}`
}

const startSubscriptionPoll = () => {
  stopSubscriptionPoll()
  if (!gotoProCardId.value) return
  subscriptionWaitSeconds = 0
  subscriptionWaitMembership = 'free'
  renderWaitStatus()
  subscriptionWaitTimer = setInterval(() => {
    if (!showGotoProModal.value) {
      stopSubscriptionPoll()
      return
    }
    subscriptionWaitSeconds += 1
    renderWaitStatus()
  }, 1000)
  subscriptionPollTimer = setInterval(async () => {
    if (!showGotoProModal.value) {
      stopSubscriptionPoll()
      return
    }
    try {
      const res = await pollCardSubscription(category.value, gotoProCardId.value)
      if (res.code === 200 && res.data?.subscribed) {
        stopSubscriptionPoll()
        const subType = (res.data.subscription_type || '').trim()
        if (subType) {
          gotoProSubscriptionType.value = subType
        }
        stripePollStatusType.value = 'success'
        stripePollStatus.value = `已检测到订阅 ${subType || '已付费'}，正在自动确认支付成功`
        message.success(`检测到订阅 ${subType || ''}，已自动确认支付成功`)
        closeAlipayPayWindow()
        handleGotoProSuccess()
        return
      }
      subscriptionWaitMembership = res.data?.subscription_type || 'free'
      renderWaitStatus()
    } catch {
      renderWaitStatus(true)
    }
  }, 5000)
}

watch(showGotoProModal, (show) => {
  if (!show) stopSubscriptionPoll()
})

onUnmounted(() => {
  stopSubscriptionPoll()
})

// 支付成功：关闭提链弹窗，打开升级成品弹窗
const handleGotoProSuccess = () => {
  stopSubscriptionPoll()
  closeAlipayPayWindow()
  showGotoProModal.value = false
  upgradeForm.value = {
    subscription_type: gotoProSubscriptionType.value,
    subscription_time: Date.now(),
    subscription_remaining_days: 30,
    purchase_price: undefined,
    purchase_from: '支付宝',
    purchase_date: Date.now(),
  }
  showGotoProSuccessModal.value = true
}

// 支付成功确认：单卡升级为成品
const handleGotoProSuccessSubmit = async () => {
  const subscriptionType = (upgradeForm.value.subscription_type || gotoProSubscriptionType.value || '').trim()
  if (!subscriptionType) {
    message.warning('请选择订阅类型')
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
      ids: [gotoProCardId.value],
      subscription_type: subscriptionType,
      subscription_time: subscriptionTime,
      subscription_remaining_days: upgradeForm.value.subscription_remaining_days,
      purchase_price: upgradeForm.value.purchase_price,
      purchase_from: upgradeForm.value.purchase_from || undefined,
      purchase_date: purchaseDate,
    })

    if (response.code === 200) {
      message.success(response.message || '更新成功')
      showGotoProSuccessModal.value = false
      await loadCards()
    } else {
      message.error(response.message || '更新失败')
      return false
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '更新失败')
    return false
  }
}

// 支付失败：关闭提链弹窗，打开备注弹窗
const handleGotoProFail = () => {
  stopSubscriptionPoll()
  closeAlipayPayWindow()
  showGotoProModal.value = false
  gotoProFailRemark.value = ''
  showGotoProFailModal.value = true
}

// 支付失败确认：保存备注到数据库
const handleGotoProFailSubmit = async () => {
  try {
    const response = await updateCardRemark(category.value, gotoProCardId.value, gotoProFailRemark.value)
    if (response.code === 200) {
      message.success('备注已保存')
      showGotoProFailModal.value = false
      await loadCards()
    } else {
      message.error(response.message || '保存失败')
      return false
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '保存失败')
    return false
  }
}

// 重置
const handleReset = () => {
  searchAccountsText.value = ''
  searchSubscriptionType.value = ''
  searchSubscriptionStatus.value = 0
  searchSellTo.value = ''
  searchPurchaseBy.value = ''
  searchIsCheck.value = 0
  searchPurchaseDate.value = null
  searchFreezeStatus.value = 0
  searchFreezeTime.value = null
  pagination.value.page = 1
  loadCards()
}

// 批量删除（status=-1 软删）
const handleBatchDelete = () => {
  const ids = checkedRowKeys.value as number[]
  if (ids.length === 0) {
    message.warning('请先勾选要删除的记录')
    return
  }
  dialog.error({
    title: '确认批量删除',
    content: `确定要删除已勾选的 ${ids.length} 条记录吗？删除后将不会在列表中显示。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await batchDeleteCards({ category: category.value, ids })
        if (res.code === 200) {
          message.success(res.message || `已删除 ${ids.length} 条`)
          checkedRowKeys.value = []
          selectedCardsMap.value.clear()
          await loadCards()
        } else {
          message.error(res.message || '删除失败')
        }
      } catch (e: any) {
        message.error(e.response?.data?.message || '删除失败')
      }
    },
  })
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
    client_id: '',
    mail_url: '',
    remark: '',
    code_link: '',
    phone: '',
    phone_link: '',
    sell_order_no: '',
    subscription_credits: undefined,
    subscription_time: undefined,
    subscription_expired_time: undefined,
    purchase_date: undefined,
    sell_date: undefined,
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
    client_id: card.client_id || '',
    mail_url: card.mail_url || '',
    remark: card.remark || '',
    code_link: card.code_link || '',
    phone: card.phone || '',
    phone_link: card.phone_link || '',
    sell_order_no: card.sell_order_no || '',
    subscription_credits: card.subscription_credits,
    subscription_time: card.subscription_time ? card.subscription_time * 1000 : undefined,
    subscription_expired_time: card.subscription_expired_time ? card.subscription_expired_time * 1000 : undefined,
    purchase_date: card.purchase_date ? card.purchase_date * 1000 : undefined,
    sell_date: card.sell_date ? card.sell_date * 1000 : undefined,
  }
  showModal.value = true
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
  const price = card.subscription_type ? getDigisellerPrice(card.subscription_type) : undefined
  shippedForm.value = {
    sell_price: price ?? 20,
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
const toUnixSec = (ms?: number | null) => (ms ? Math.floor(ms / 1000) : undefined)

const handleSubmit = async () => {
  // 验证表单
  await formRef.value?.validate()

  const payload: CardRequest = {
    ...formData.value,
    subscription_time: toUnixSec(formData.value.subscription_time),
    subscription_expired_time: toUnixSec(formData.value.subscription_expired_time),
    purchase_date: toUnixSec(formData.value.purchase_date),
    sell_date: toUnixSec(formData.value.sell_date),
  }

  try {
    if (isEdit.value) {
      // 更新卡密
      const response = await updateCard(category.value, currentEditId.value, payload)
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
      const response = await createCard(category.value, payload)
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

// 复制卡密信息到剪贴板（---- 分隔格式带标题行）
const handleCopy = async (card: Card, format: string) => {
  let text = ''
  if (format === 'digiseller') {
    text = formatDigiseller(card)
  } else if (format === 'digiseller_auto') {
    text = formatDigisellerAuto(card)
  } else if (format === 'reverse') {
    text = `账号----token${phoneDashTitle(card)}\n${card.account}----${card.token || ''}${phoneDashSuffix(card)}`
  } else if (format === 'code_method') {
    // 接码方式：标题 账号----接码方式；数据 账号----接码链接----账号----邮箱密码
    text = `账号----接码方式${phoneDashTitle(card)}\n${card.account}----${card.code_link || ''}----${card.account}----${card.mail_password || ''}${phoneDashSuffix(card)}`
  } else {
    // 国内格式：账号----密码----邮箱密码----token----手机号----短信接码地址
    text = `账号----密码----邮箱密码----token${domesticPhoneTitle}\n${card.account}----${card.password || ''}----${card.mail_password || ''}----${card.token || ''}${domesticPhoneSuffix(card)}`
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
    case 'code_method':
      // 接码方式：接码链接----账号----邮箱密码
      return `${card.code_link || ''}----${card.account || ''}----${card.mail_password || ''}`
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
      if (card.status === -2) return '封禁'
      if (card.status === 2) return '禁用'
      return '正常'
    default:
      return String((card as unknown as Record<string, unknown>)[field] ?? '')
  }
}

// 生成并下载导出文件
const doDownload = (cards: Card[]) => {
  let content = ''
  if (exportFormatMode.value === 'digiseller') {
    content = cards.map(formatDigiseller).join('\n\n')
  } else if (exportFormatMode.value === 'digiseller_auto') {
    content = cards.map(formatDigisellerAuto).join('\n')
  } else {
    const lines = cards.map(card => {
      return exportSelectedFields.value.map(f => getFieldValue(card, f)).join('----')
    })
    content = lines.join('\n')
  }
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
    const response = await exportCards(buildCardQueryParams())

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

// 打开批量升级为成品弹窗（重置表单，避免沿用上次的订阅类型）
const handleOpenUpgradeModal = () => {
  upgradeForm.value = {
    subscription_type: 'pro',
    subscription_time: Date.now(),
    subscription_remaining_days: 30,
    purchase_price: undefined,
    purchase_from: '支付宝',
    purchase_date: Date.now(),
  }
  showUpgradeModal.value = true
}

// 批量升级为成品
const handleBatchUpgrade = async () => {
  if (checkedRowKeys.value.length === 0) {
    message.warning('请先勾选要升级的记录')
    return false
  }

  const subscriptionType = (upgradeForm.value.subscription_type || '').trim()
  if (!subscriptionType) {
    message.warning('请选择订阅类型')
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
      subscription_type: subscriptionType,
      subscription_time: subscriptionTime,
      subscription_remaining_days: upgradeForm.value.subscription_remaining_days,
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
        subscription_remaining_days: 30,
        purchase_price: undefined,
        purchase_from: '支付宝',
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
    purchase_from: '支付宝',
    purchase_by: '',
    purchase_date: undefined,
    mail_url: 'https://login.live.com',
    sell_status: 1,
    subscription_status: 1,
    account_type: 2,
    code_link: 'https://tool.toolsvip.cc/easy-mailbox/frontend',
    phone_link: '',
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
          subscriptionTime = parseImportTimestamp(parts[mapping.subscription_time - 1]) ?? currentTimestamp
        } else {
          subscriptionTime = currentTimestamp
        }
      }

      // 订阅过期时间：优先从行数据读取，否则用公共「订阅剩余天数」推算
      let rowSubscriptionExpiredTime: number | undefined = undefined
      if (mapping.subscription_expired_time > 0 && parts[mapping.subscription_expired_time - 1]) {
        rowSubscriptionExpiredTime = parseImportTimestamp(parts[mapping.subscription_expired_time - 1])
      }

      const token = mapping.token > 0 && parts[mapping.token - 1] ? parts[mapping.token - 1] : ''
      const clientId = mapping.client_id > 0 && parts[mapping.client_id - 1] ? parts[mapping.client_id - 1] : ''
      const apiKey = mapping.api_key > 0 && parts[mapping.api_key - 1] ? parts[mapping.api_key - 1] : ''
      const twoFA = mapping['2fa'] > 0 && parts[mapping['2fa'] - 1] ? parts[mapping['2fa'] - 1] : ''
      const phone = mapping.phone > 0 && parts[mapping.phone - 1] ? parts[mapping.phone - 1] : ''
      const phoneLink = mapping.phone_link > 0 && parts[mapping.phone_link - 1]
        ? parts[mapping.phone_link - 1]
        : (batchConfig.value.phone_link || '')

      let rowSubscriptionType: string | undefined
      if (mapping.subscription_type > 0 && parts[mapping.subscription_type - 1]) {
        rowSubscriptionType = parts[mapping.subscription_type - 1]
      }
      const subscriptionMeta = resolveSubscriptionMeta(
        rowSubscriptionType || batchConfig.value.subscription_type || undefined
      )

      const cardData: CardRequest = {
        account,
        password: password || undefined,
        mail_password: mailPassword || undefined,
        subscription_status: subscriptionMeta.subscription_status,
        subscription_type: subscriptionMeta.subscription_type,
        subscription_time: subscriptionTime,
        subscription_expired_time: rowSubscriptionExpiredTime ?? subscriptionExpiredTime,
        purchase_date: purchaseDate,
        purchase_price: batchConfig.value.purchase_price,
        purchase_from: batchConfig.value.purchase_from || undefined,
        purchase_by: batchConfig.value.purchase_by || undefined,
        sell_status: batchConfig.value.sell_status,
        account_type: subscriptionMeta.account_type,
        status: 1,
        token: token || undefined,
        client_id: clientId || undefined,
        api_key: apiKey || undefined,
        '2fa': twoFA || undefined,
        mail_url: batchConfig.value.mail_url || undefined,
        code_link: batchConfig.value.code_link || undefined,
        phone: phone || undefined,
        phone_link: phoneLink || undefined,
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

// 打开批量取货弹窗，根据已勾选卡密的订阅类型自动填充售价
const handleOpenBatchPickup = () => {
  const selectedCards = checkedRowKeys.value
    .map(id => selectedCardsMap.value.get(id as number))
    .filter((c): c is Card => !!c)

  const subscriptionTypes = [...new Set(selectedCards.map(c => c.subscription_type).filter((t): t is string => !!t))]
  let defaultPrice: number | undefined = undefined
  if (subscriptionTypes.length === 1) {
    defaultPrice = getDigisellerPrice(subscriptionTypes[0])
  }

  batchPickupForm.value = {
    format: 'digiseller',
    sell_price: defaultPrice ?? 20,
    sell_to: 'Digiseller',
  }
  showBatchPickupModal.value = true
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
        if (!card.password && !card.mail_password) {
          return `Пожалуйста, войдите в систему, используя код подтверждения, отправленный на электронную почту:\n\n${card.account}\n\nmail-login: ${card.mail_url || ''}${phoneFields(card, '\n')}\n\nПожалуйста, выполните следующие шаги заново:\n1. Введите аккаунт: ${card.account}\n2. Нажмите «Далее»\n3. Нажмите кнопку: «Email sign-in code»`
        }
        return formatDigiseller(card)
      } else if (fmt === 'digiseller_auto') {
        return formatDigisellerAuto(card)
      } else if (fmt === 'reverse') {
        return `${card.account}----${card.token || ''}${phoneDashSuffix(card)}`
      } else if (fmt === 'code_method') {
        // 接码方式：账号----接码链接----账号----邮箱密码
        return `${card.account}----${card.code_link || ''}----${card.account}----${card.mail_password || ''}${phoneDashSuffix(card)}`
      } else {
        return `${card.account}----${card.password || ''}----${card.mail_password || ''}----${card.token || ''}${domesticPhoneSuffix(card)}`
      }
    })

    // ---- 分隔格式：标题行 + 数据行；Digiseller 多行格式用空行分隔
    const phoneTitle = cards.some(hasPhoneReceive) ? '----手机号----短信接码地址' : ''
    let clipboardText = ''
    if (fmt === 'domestic') {
      clipboardText = `账号----密码----邮箱密码----token${domesticPhoneTitle}\n` + lines.join('\n')
    } else if (fmt === 'reverse') {
      clipboardText = `账号----token${phoneTitle}\n` + lines.join('\n')
    } else if (fmt === 'code_method') {
      clipboardText = `账号----接码方式${phoneTitle}\n` + lines.join('\n')
    } else if (fmt === 'digiseller_auto') {
      clipboardText = lines.join('\n')
    } else {
      clipboardText = lines.join('\n\n')
    }

    try {
      await navigator.clipboard.writeText(clipboardText)
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

  // 并行加载未售订阅类型和Digiseller价格配置
  try {
    const [typesResponse, pricesResponse] = await Promise.all([
      getUnsoldSubscriptionTypes(category.value),
      getDigisellerPrices(),
    ])

    // 缓存价格配置
    if (pricesResponse.code === 200 && pricesResponse.data) {
      digisellerPrices.value = pricesResponse.data
    }

    if (typesResponse.code === 200) {
      unsoldSubscriptionTypes.value = (typesResponse.data || []).map((type) => ({
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
      message.error(typesResponse.message || '获取订阅类型失败')
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

        // 根据订阅类型自动填充Digiseller配置的售价
        const matchedPrice = digisellerPrices.value.find(
          (p) => p.subscription_type === pickupForm.value.subscription_type
        )
        if (matchedPrice !== undefined) {
          completeForm.value.sell_price = matchedPrice.price
        }

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
        const defaultText = `Скорость — наше всё, а отзыв — твой победный жест!

Ваш заказ выполнен ! 

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

// 监听代充弹窗订阅类型变化，自动填充Digiseller售价
watch(
  () => rechargeForm.value.subscription_type,
  (type) => {
    if (type) {
      const price = getDigisellerPrice(type)
      if (price !== undefined) {
        rechargeForm.value.sell_price = price
      }
    }
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

      // 重置分页、搜索条件和勾选状态
      pagination.value.page = 1
      searchAccountsText.value = ''
      searchSubscriptionType.value = ''
      searchSubscriptionStatus.value = 0
      searchSellTo.value = ''
      searchPurchaseBy.value = ''
      searchIsCheck.value = 0
      searchPurchaseDate.value = null
      searchFreezeStatus.value = 0
      searchFreezeTime.value = null
      checkedRowKeys.value = []
      selectedCardsMap.value.clear()

      // 重新加载数据
      loadCards()
    }
  }
)

// 打开批量冻结/解冻弹窗
const handleOpenFreezeModal = (action: 1 | -1) => {
  freezeAction.value = action
  freezeRemark.value = ''
  showFreezeModal.value = true
}

// 提交批量冻结/解冻
const handleFreezeSubmit = async () => {
  const ids = checkedRowKeys.value as number[]
  if (ids.length === 0) {
    message.warning('请先勾选要操作的记录')
    return false
  }
  try {
    const response = await batchFreezeCards({
      category: category.value,
      ids,
      freeze: freezeAction.value,
      remark: freezeAction.value === 1 ? freezeRemark.value : '',
    })
    if (response.code === 200) {
      message.success(response.message || '操作成功')
      showFreezeModal.value = false
      checkedRowKeys.value = []
      selectedCardsMap.value.clear()
      await loadCards()
    } else {
      message.error(response.message || '操作失败')
      return false
    }
  } catch (error: any) {
    message.error(error.response?.data?.message || '操作失败')
    return false
  }
}

// 加载Digiseller价格配置
const loadDigisellerPrices = async () => {
  try {
    const res = await getDigisellerPrices()
    if (res.code === 200 && res.data) {
      digisellerPrices.value = res.data
    }
  } catch {
    // 静默失败，不影响主流程
  }
}

// 根据订阅类型获取Digiseller配置的售价
const getDigisellerPrice = (subscriptionType: string): number | undefined => {
  const matched = digisellerPrices.value.find(p => p.subscription_type === subscriptionType)
  return matched?.price
}

// 初始化
onMounted(() => {
  loadCards()
  loadDigisellerPrices()
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
