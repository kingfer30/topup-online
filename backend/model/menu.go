package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Menu 菜单模型
type Menu struct {
	Id        int            `json:"id" gorm:"primaryKey;autoIncrement"`
	ParentId  int            `json:"parent_id" gorm:"type:bigint;default:0;index;comment:父菜单ID, 0为顶级菜单"`
	Title     string         `json:"title" gorm:"type:varchar(100);not null;comment:菜单标题"`
	Key       string         `json:"key" gorm:"type:varchar(100);not null;index;comment:菜单唯一key"`
	Path      string         `json:"path" gorm:"type:varchar(200);comment:路由路径"`
	Icon      string         `json:"icon" gorm:"type:varchar(50);comment:菜单图标(emoji)"`
	Sort      int            `json:"sort" gorm:"type:bigint;default:0;comment:排序权重，数值越小越靠前"`
	Status    int            `json:"status" gorm:"type:tinyint(1);default:1;comment:状态: 1启用 0禁用"`
	IsDelete  int            `json:"is_delete" gorm:"type:tinyint(1);default:-1;index;comment:是否删除: 1是 -1否"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Children  []*Menu        `json:"children,omitempty" gorm:"-"` // 子菜单，不存储到数据库
}

// MenuTreeNode 菜单树节点
type MenuTreeNode struct {
	Menu
	Children []*MenuTreeNode `json:"children,omitempty"`
}

// GetAllMenus 获取所有菜单（扁平列表）
func GetAllMenus() ([]Menu, error) {
	var menus []Menu
	err := DB.Where("status = ? AND is_delete = ?", 1, -1).Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

// GetMenuTree 获取菜单树结构
func GetMenuTree() ([]*Menu, error) {
	var menus []Menu
	err := DB.Where("status = ? AND is_delete = ?", 1, -1).Order("sort asc, id asc").Find(&menus).Error
	if err != nil {
		return nil, err
	}

	// 构建菜单树
	menuMap := make(map[int]*Menu)
	var rootMenus []*Menu

	// 第一遍遍历，创建所有菜单的引用
	for i := range menus {
		menu := &menus[i]
		menu.Children = []*Menu{}
		menuMap[menu.Id] = menu
	}

	// 第二遍遍历，建立父子关系（保持原始排序）
	for i := range menus {
		menu := menuMap[menus[i].Id]
		if menu.ParentId == 0 {
			// 顶级菜单
			rootMenus = append(rootMenus, menu)
		} else {
			// 子菜单
			if parent, ok := menuMap[menu.ParentId]; ok {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return rootMenus, nil
}

// GetMenuById 根据ID获取菜单
func GetMenuById(id int) (*Menu, error) {
	if id == 0 {
		return nil, errors.New("id 为空")
	}
	var menu Menu
	err := DB.Where("id = ? AND is_delete = ?", id, -1).First(&menu).Error
	return &menu, err
}

// CreateMenu 创建菜单
func (menu *Menu) Create() error {
	if menu.Title == "" || menu.Key == "" {
		return errors.New("标题和key不能为空")
	}

	// 检查key是否已存在（排除已删除的记录）
	var count int64
	DB.Model(&Menu{}).Where("key = ? AND is_delete = ?", menu.Key, -1).Count(&count)
	if count > 0 {
		return errors.New("菜单key已存在")
	}

	// 设置默认值
	if menu.IsDelete == 0 {
		menu.IsDelete = -1
	}

	return DB.Create(menu).Error
}

// UpdateMenu 更新菜单
func (menu *Menu) Update() error {
	if menu.Id == 0 {
		return errors.New("id 为空")
	}
	if menu.Title == "" || menu.Key == "" {
		return errors.New("标题和key不能为空")
	}

	// 检查key是否与其他菜单冲突（排除已删除的记录）
	var count int64
	DB.Model(&Menu{}).Where("key = ? AND id != ? AND is_delete = ?", menu.Key, menu.Id, -1).Count(&count)
	if count > 0 {
		return errors.New("菜单key已存在")
	}

	// 不允许将菜单设置为自己的子菜单
	if menu.ParentId == menu.Id {
		return errors.New("不能将菜单设置为自己的子菜单")
	}

	return DB.Model(menu).Updates(menu).Error
}

// DeleteMenu 删除菜单（软删除）
func (menu *Menu) Delete() error {
	if menu.Id == 0 {
		return errors.New("id 为空")
	}

	// 检查是否有未删除的子菜单
	var count int64
	DB.Model(&Menu{}).Where("parent_id = ? AND is_delete = ?", menu.Id, -1).Count(&count)
	if count > 0 {
		return errors.New("该菜单下有子菜单，无法删除")
	}

	// 软删除：设置 is_delete = 1 并使用 GORM 的软删除（设置 deleted_at）
	err := DB.Model(menu).Where("id = ?", menu.Id).Update("is_delete", 1).Error
	if err != nil {
		return err
	}

	// 同时执行 GORM 的软删除
	return DB.Delete(menu).Error
}

// DeleteMenuById 根据ID删除菜单
func DeleteMenuById(id int) error {
	if id == 0 {
		return errors.New("id 为空")
	}
	menu := Menu{Id: id}
	return menu.Delete()
}

// GetMenusByParentId 获取指定父菜单的子菜单
func GetMenusByParentId(parentId int) ([]Menu, error) {
	var menus []Menu
	err := DB.Where("parent_id = ? AND status = ? AND is_delete = ?", parentId, 1, -1).Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

// GetAllMenusForManagement 获取所有菜单（用于管理界面，包括禁用的，但不包括已删除的）
func GetAllMenusForManagement() ([]Menu, error) {
	var menus []Menu
	err := DB.Where("is_delete = ?", -1).Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

// CreateCardMenuWithTable 创建卡密菜单（父菜单+3个子菜单）并创建对应的数据表
func CreateCardMenuWithTable(category, menuName, icon string, sort int) error {
	// 开启事务
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建父菜单
	parentMenu := &Menu{
		ParentId: 0,
		Title:    menuName,
		Key:      "card_" + category,
		Path:     "",
		Icon:     icon,
		Sort:     sort,
		Status:   1,
		IsDelete: -1,
	}

	if err := tx.Create(parentMenu).Error; err != nil {
		tx.Rollback()
		return errors.New("创建父菜单失败: " + err.Error())
	}

	// 2. 创建3个子菜单
	childMenus := []Menu{
		{
			ParentId: parentMenu.Id,
			Title:    "普号列表",
			Key:      "card_" + category + "_all",
			Path:     "/admin/cards?category=" + category + "&type=all",
			Icon:     "",
			Sort:     1,
			Status:   1,
			IsDelete: -1,
		},
		{
			ParentId: parentMenu.Id,
			Title:    "未售列表",
			Key:      "card_" + category + "_unsold",
			Path:     "/admin/cards?category=" + category + "&type=unsold",
			Icon:     "",
			Sort:     2,
			Status:   1,
			IsDelete: -1,
		},
		{
			ParentId: parentMenu.Id,
			Title:    "已售列表",
			Key:      "card_" + category + "_sold",
			Path:     "/admin/cards?category=" + category + "&type=sold",
			Icon:     "",
			Sort:     3,
			Status:   1,
			IsDelete: -1,
		},
	}

	for _, childMenu := range childMenus {
		if err := tx.Create(&childMenu).Error; err != nil {
			tx.Rollback()
			return errors.New("创建子菜单失败: " + err.Error())
		}
	}

	// 3. 创建对应的数据表
	tableName := "cards_" + category
	if err := CreateCardTable(tx, tableName); err != nil {
		tx.Rollback()
		return errors.New("创建数据表失败: " + err.Error())
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.New("提交事务失败: " + err.Error())
	}

	return nil
}

// CreateCardTable 创建卡密数据表
func CreateCardTable(tx *gorm.DB, tableName string) error {
	// 构建创建表的SQL语句
	createTableSQL := `
CREATE TABLE IF NOT EXISTS ` + "`" + tableName + "`" + ` (
  ` + "`id`" + ` bigint(20) NOT NULL AUTO_INCREMENT,
  ` + "`account`" + ` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  ` + "`password`" + ` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  ` + "`mail_password`" + ` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  ` + "`subscription_status`" + ` tinyint(2) DEFAULT 1 COMMENT '订阅状态 1已订阅 2未订阅',
  ` + "`subscription_type`" + ` varchar(30) DEFAULT NULL COMMENT '订阅类型',
  ` + "`subscription_time`" + ` bigint(20) DEFAULT NULL COMMENT '订阅时间',
  ` + "`subscription_expired_time`" + ` bigint(20) DEFAULT NULL COMMENT '订阅过期时间',
  ` + "`subscription_credits`" + ` decimal(10,2) DEFAULT NULL COMMENT '订阅额度',
  ` + "`is_check`" + ` tinyint(2) NOT NULL DEFAULT -1 COMMENT '检查状态 -1未检查 1检查成功 2检查失败',
  ` + "`check_time`" + ` bigint(20) DEFAULT NULL COMMENT '检查时间',
  ` + "`purchase_date`" + ` bigint(20) DEFAULT NULL COMMENT '购买时间',
  ` + "`purchase_price`" + ` decimal(10,2) DEFAULT NULL COMMENT '购买价格(成本)',
  ` + "`purchase_from`" + ` varchar(50) DEFAULT NULL COMMENT '购买平台',
  ` + "`purchase_by`" + ` varchar(100) DEFAULT NULL COMMENT '卖家名称',
  ` + "`sell_price`" + ` decimal(10,2) DEFAULT NULL COMMENT '出售价格',
  ` + "`sell_date`" + ` bigint(20) DEFAULT NULL COMMENT '出售时间',
  ` + "`sell_to`" + ` varchar(50) DEFAULT NULL COMMENT '出售对方',
  ` + "`sell_order_no`" + ` varchar(100) DEFAULT NULL COMMENT '出售订单号',
  ` + "`sell_status`" + ` tinyint(2) DEFAULT 1 COMMENT '出售状态 1 未出售 2发货中 3已出售',
  ` + "`account_type`" + ` tinyint(2) DEFAULT 1 COMMENT '账号类型 1普号 2成品',
  ` + "`status`" + ` tinyint(2) DEFAULT 1 COMMENT '状态 1 正常 2 禁用',
  ` + "`api_key`" + ` varchar(300) COLLATE utf8mb4_unicode_ci  NULL COMMENT 'apikey',
  ` + "`token`" + ` text COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'token',
  ` + "`2fa`" + ` varchar(100) COLLATE utf8mb4_unicode_ci  NULL COMMENT '2fa',
  ` + "`mail_url`" + ` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱地址',
  ` + "`remark`" + ` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  ` + "`code_link`" + ` varchar(300) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '接码链接',
  ` + "`freeze_status`" + ` tinyint(2) NOT NULL DEFAULT -1 COMMENT '冻结状态 -1未冻结 1已冻结',
  ` + "`freeze_time`" + ` bigint(20) DEFAULT NULL COMMENT '冻结时间',
  ` + "`freeze_remark`" + ` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '冻结备注',
  ` + "`created_at`" + ` datetime(3) DEFAULT NULL,
  ` + "`updated_at`" + ` datetime(3) DEFAULT NULL,
  ` + "`deleted_at`" + ` datetime(3) DEFAULT NULL,
  PRIMARY KEY (` + "`id`" + `),
  UNIQUE KEY ` + "`idx_account`" + ` (` + "`account`" + `),
  KEY ` + "`idx_subscription_sell`" + ` (` + "`subscription_type`" + `, ` + "`sell_status`" + `),
  KEY ` + "`idx_subscription_type`" + ` (` + "`subscription_type`" + `),
  KEY ` + "`idx_subscription_time`" + ` (` + "`subscription_time`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='卡密表';
`

	return tx.Exec(createTableSQL).Error
}
