package model

// EnsureGptMenus 幂等插入 GPT卡密 和 GPT-CDK 菜单
func EnsureGptMenus() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	menus := []Menu{
		{
			ParentId: 0,
			Title:    "GPT卡密",
			Key:      "gpt-cards",
			Path:     "/admin/gpt-cards",
			Icon:     "🃏",
			Sort:     8,
			Status:   1,
			IsDelete: -1,
		},
		{
			ParentId: 0,
			Title:    "GPT-CDK",
			Key:      "gpt-cdk",
			Path:     "/admin/gpt-cdk",
			Icon:     "🔑",
			Sort:     9,
			Status:   1,
			IsDelete: -1,
		},
	}

	for _, m := range menus {
		var existing Menu
		err := tx.Where("`key` = ? AND is_delete = ?", m.Key, -1).First(&existing).Error
		if err != nil {
			// 不存在则创建
			if err := tx.Create(&m).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		// 已存在则跳过
	}

	return tx.Commit().Error
}

// EnsureAIMenuPlacement 确保「AI翻译」位于顶级菜单，排序在「用户管理」之前（用于已初始化数据库升级）
func EnsureAIMenuPlacement() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var m Menu
	err := tx.Where("`key` = ? AND is_delete = ?", "ai-translate", -1).First(&m).Error
	if err != nil {
		// 不存在则创建
		m = Menu{
			ParentId: 0,
			Title:    "AI翻译",
			Key:      "ai-translate",
			Path:     "/admin/ai-translate",
			Icon:     "🌐",
			Sort:     2,
			Status:   1,
			IsDelete: -1,
		}
		if err := tx.Create(&m).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 已存在则搬到顶级并修正字段
		if err := tx.Model(&Menu{}).Where("id = ?", m.Id).Updates(map[string]any{
			"parent_id": 0,
			"title":     "AI翻译",
			"path":      "/admin/ai-translate",
			"icon":      "🌐",
			"sort":      2,
			"status":    1,
			"is_delete": -1,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 同步调整其它顶级菜单排序（不强制要求存在）
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "user", -1).Updates(map[string]any{"sort": 3}).Error
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "order", -1).Updates(map[string]any{"sort": 4}).Error
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "sales-talk", -1).Updates(map[string]any{"sort": 5}).Error
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "mirror", -1).Updates(map[string]any{"sort": 6}).Error
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "system", -1).Updates(map[string]any{"sort": 7}).Error

	return tx.Commit().Error
}
