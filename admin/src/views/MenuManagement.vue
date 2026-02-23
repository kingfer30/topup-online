<template>
  <div class="menu-management">
    <n-card title="菜单管理">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="handleAdd">
            <template #icon>
              <span>➕</span>
            </template>
            新增菜单
          </n-button>
          <n-button type="success" @click="handleAddCardMenu">
            <template #icon>
              <span>🎫</span>
            </template>
            新增卡密菜单
          </n-button>
        </n-space>
      </template>

      <!-- 菜单表格 -->
      <n-data-table
        :columns="columns"
        :data="menuList"
        :pagination="false"
        :bordered="false"
        :single-line="false"
        :loading="loading"
        :row-key="(row: Menu) => row.id"
        default-expand-all
      />
    </n-card>

    <!-- 新增卡密菜单对话框 -->
    <n-modal
      v-model:show="showCardMenuModal"
      title="新增卡密菜单"
      preset="dialog"
      positive-text="创建"
      negative-text="取消"
      @positive-click="handleCardMenuSubmit"
      style="width: 600px"
    >
      <n-form
        ref="cardMenuFormRef"
        :model="cardMenuFormData"
        :rules="cardMenuRules"
        label-placement="left"
        label-width="120px"
        require-mark-placement="left"
        style="margin-top: 20px"
      >
        <n-form-item label="卡密类别" path="category">
          <n-input
            v-model:value="cardMenuFormData.category"
            placeholder="请输入卡密类别（如：cursor）"
          />
          <template #feedback>
            <span class="text-gray-500 text-xs">用于数据库表名，只能包含字母、数字和下划线</span>
          </template>
        </n-form-item>

        <n-form-item label="菜单名称" path="menuName">
          <n-input
            v-model:value="cardMenuFormData.menuName"
            placeholder="请输入菜单名称（如：cursor卡密）"
          />
        </n-form-item>

        <n-form-item label="菜单图标" path="icon">
          <n-input v-model:value="cardMenuFormData.icon" placeholder="请输入emoji图标（如：🎫）" />
        </n-form-item>

        <n-form-item label="排序" path="sort">
          <n-input-number
            v-model:value="cardMenuFormData.sort"
            :min="0"
            placeholder="数值越小越靠前"
            style="width: 100%"
          />
        </n-form-item>

        <n-alert type="info" style="margin-top: 10px">
          将自动创建父菜单和3个子菜单：普号列表、未售列表、已售列表
        </n-alert>
      </n-form>
    </n-modal>

    <!-- 新增/编辑对话框 -->
    <n-modal
      v-model:show="showModal"
      :title="isEdit ? '编辑菜单' : '新增菜单'"
      preset="dialog"
      :positive-text="isEdit ? '保存' : '创建'"
      negative-text="取消"
      @positive-click="handleSubmit"
      style="width: 600px"
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
        <n-form-item label="父菜单" path="parent_id">
          <n-tree-select
            v-model:value="formData.parent_id"
            :options="parentMenuOptions"
            placeholder="选择父菜单（不选则为顶级菜单）"
            clearable
            :default-value="0"
          />
        </n-form-item>

        <n-form-item label="菜单标题" path="title">
          <n-input v-model:value="formData.title" placeholder="请输入菜单标题" />
        </n-form-item>

        <n-form-item label="菜单Key" path="key">
          <n-input v-model:value="formData.key" placeholder="请输入菜单唯一key" />
        </n-form-item>

        <n-form-item label="路由路径" path="path">
          <n-input
            v-model:value="formData.path"
            placeholder="请输入路由路径（如：/admin/users）"
          />
          <template #feedback>
            <span class="text-gray-500 text-xs">父菜单可不填路径</span>
          </template>
        </n-form-item>

        <n-form-item label="菜单图标" path="icon">
          <n-input v-model:value="formData.icon" placeholder="请输入emoji图标（如：📊）" />
        </n-form-item>

        <n-form-item label="排序" path="sort">
          <n-input-number
            v-model:value="formData.sort"
            :min="0"
            placeholder="数值越小越靠前"
            style="width: 100%"
          />
        </n-form-item>

        <n-form-item label="状态" path="status">
          <n-switch v-model:value="statusSwitch">
            <template #checked>启用</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import {
  NCard,
  NButton,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
  NTreeSelect,
  NSpace,
  NTag,
  NAlert,
  useMessage,
  useDialog,
  type DataTableColumns,
  type FormInst,
  type FormRules,
} from 'naive-ui'
import {
  getAllMenus,
  createMenu,
  updateMenu,
  deleteMenu,
  createCardMenu,
  type Menu,
  type MenuRequest,
  type CardMenuRequest,
} from '@/api/menu'

const message = useMessage()
const dialog = useDialog()

// 状态
const loading = ref(false)
const showModal = ref(false)
const isEdit = ref(false)
const menuList = ref<Menu[]>([])
const formRef = ref<FormInst | null>(null)

// 卡密菜单相关状态
const showCardMenuModal = ref(false)
const cardMenuFormRef = ref<FormInst | null>(null)
const cardMenuFormData = ref<CardMenuRequest>({
  category: '',
  menuName: '',
  icon: '🎫',
  sort: 0,
})

// 表单数据
const formData = ref<MenuRequest>({
  parent_id: 0,
  title: '',
  key: '',
  path: '',
  icon: '',
  sort: 0,
  status: 1,
})

// 当前编辑的菜单ID
const currentEditId = ref<number>(0)

// 状态开关（用于UI显示）
const statusSwitch = computed({
  get: () => formData.value.status === 1,
  set: (val) => {
    formData.value.status = val ? 1 : 0
  },
})

// 表单验证规则
const rules: FormRules = {
  title: [{ required: true, message: '请输入菜单标题', trigger: 'blur' }],
  key: [{ required: true, message: '请输入菜单key', trigger: 'blur' }],
}

// 卡密菜单表单验证规则
const cardMenuRules: FormRules = {
  category: [
    { required: true, message: '请输入卡密类别', trigger: 'blur' },
    {
      pattern: /^[a-zA-Z0-9_]+$/,
      message: '只能包含字母、数字和下划线',
      trigger: 'blur',
    },
  ],
  menuName: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
}

// 父菜单选项（用于下拉选择）
const parentMenuOptions = computed(() => {
  const options = [
    {
      label: '顶级菜单',
      key: 0,
      value: 0,
    },
  ]

  const buildOptions = (menus: Menu[], level = 0) => {
    menus.forEach((menu) => {
      // 编辑时不能选择自己作为父菜单
      if (isEdit.value && menu.id === currentEditId.value) {
        return
      }

      const prefix = '　'.repeat(level)
      options.push({
        label: prefix + menu.title,
        key: menu.id,
        value: menu.id,
      })

      // 递归处理子菜单
      if (menu.children && menu.children.length > 0) {
        buildOptions(menu.children, level + 1)
      }
    })
  }

  // menuList.value 已经是树形结构，直接使用
  buildOptions(menuList.value)

  return options
})

// 构建菜单树
const buildMenuTree = (menus: Menu[]): Menu[] => {
  const menuMap = new Map<number, Menu>()
  const rootMenus: Menu[] = []

  // 创建映射
  menus.forEach((menu) => {
    menuMap.set(menu.id, { ...menu, children: [] })
  })

  // 构建树形结构
  menus.forEach((menu) => {
    const menuNode = menuMap.get(menu.id)!
    if (menu.parent_id === 0) {
      rootMenus.push(menuNode)
    } else {
      const parent = menuMap.get(menu.parent_id)
      if (parent) {
        if (!parent.children) {
          parent.children = []
        }
        parent.children.push(menuNode)
      }
    }
  })

  return rootMenus
}

// 表格列定义
const columns: DataTableColumns<Menu> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: '菜单标题',
    key: 'title',
    width: 200,
  },
  {
    title: 'Key',
    key: 'key',
    width: 150,
  },
  {
    title: '路由路径',
    key: 'path',
    width: 200,
  },
  {
    title: '图标',
    key: 'icon',
    width: 60,
    render: (row) => {
      return h('span', { style: 'font-size: 20px' }, row.icon || '-')
    },
  },
  {
    title: '排序',
    key: 'sort',
    width: 80,
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
        { default: () => (row.status === 1 ? '启用' : '禁用') }
      )
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
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

// 获取菜单层级
const getMenuLevel = (menuId: number, menus: Menu[], level = 0): number => {
  for (const menu of menus) {
    if (menu.id === menuId) {
      return level
    }
    if (menu.children && menu.children.length > 0) {
      const childLevel = getMenuLevel(menuId, menu.children, level + 1)
      if (childLevel > level) {
        return childLevel
      }
    }
  }
  return level
}

// 加载菜单列表
const loadMenus = async () => {
  loading.value = true
  try {
    const response = await getAllMenus()
    if (response.code === 200) {
      // 直接使用树形结构，让 NDataTable 自动处理层级展示
      menuList.value = buildMenuTree(response.data || [])
    }
  } catch (error) {
    console.error('加载菜单失败', error)
    message.error('加载菜单失败')
  } finally {
    loading.value = false
  }
}

// 新增菜单
const handleAdd = () => {
  isEdit.value = false
  currentEditId.value = 0
  formData.value = {
    parent_id: 0,
    title: '',
    key: '',
    path: '',
    icon: '',
    sort: 0,
    status: 1,
  }
  showModal.value = true
}

// 编辑菜单
const handleEdit = (menu: Menu) => {
  isEdit.value = true
  currentEditId.value = menu.id
  formData.value = {
    parent_id: menu.parent_id,
    title: menu.title,
    key: menu.key,
    path: menu.path || '',
    icon: menu.icon || '',
    sort: menu.sort,
    status: menu.status,
  }
  showModal.value = true
}

// 删除菜单
const handleDelete = (menu: Menu) => {
  // 检查是否有子菜单
  if (menu.children && menu.children.length > 0) {
    message.error('该菜单下有子菜单，无法删除')
    return
  }

  dialog.warning({
    title: '确认删除',
    content: `确定要删除菜单"${menu.title}"吗？删除后该菜单将不再显示。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const response = await deleteMenu(menu.id)
        if (response.code === 200) {
          message.success('删除成功')
          await loadMenus()
          // 通知布局组件刷新菜单
          window.dispatchEvent(new Event('refreshMenus'))
        } else {
          message.error(response.message || '删除失败')
        }
      } catch (error: any) {
        console.error('删除菜单失败', error)
        message.error(error.response?.data?.message || '删除失败')
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
      // 更新菜单
      const response = await updateMenu(currentEditId.value, formData.value)
      if (response.code === 200) {
        message.success('更新成功')
        showModal.value = false
        await loadMenus()
        // 通知布局组件刷新菜单
        window.dispatchEvent(new Event('refreshMenus'))
      } else {
        message.error(response.message || '更新失败')
        return false
      }
    } else {
      // 创建菜单
      const response = await createMenu(formData.value)
      if (response.code === 200) {
        message.success('创建成功')
        showModal.value = false
        await loadMenus()
        // 通知布局组件刷新菜单
        window.dispatchEvent(new Event('refreshMenus'))
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

// 打开新增卡密菜单对话框
const handleAddCardMenu = () => {
  cardMenuFormData.value = {
    category: '',
    menuName: '',
    icon: '🎫',
    sort: 0,
  }
  showCardMenuModal.value = true
}

// 提交卡密菜单表单
const handleCardMenuSubmit = async () => {
  // 验证表单
  await cardMenuFormRef.value?.validate()

  try {
    const response = await createCardMenu(cardMenuFormData.value)
    if (response.code === 200) {
      message.success('创建卡密菜单成功')
      showCardMenuModal.value = false
      await loadMenus()
      // 通知布局组件刷新菜单
      window.dispatchEvent(new Event('refreshMenus'))
    } else {
      message.error(response.message || '创建失败')
      return false
    }
  } catch (error: any) {
    console.error('创建卡密菜单失败', error)
    message.error(error.response?.data?.message || '创建失败')
    return false
  }
}

// 初始化
onMounted(() => {
  loadMenus()
})
</script>

<style scoped>
.menu-management {
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
</style>

