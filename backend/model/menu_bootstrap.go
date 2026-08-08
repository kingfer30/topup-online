package model

// EnsureGptBusinessMenu 幂等确保「GPT业务」父级菜单及子菜单（含 GPT卡密、GPT-CDK、充值与链接页）
func EnsureGptBusinessMenu() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var parent Menu
	err := tx.Where("`key` = ? AND is_delete = ?", "gpt-business-root", -1).First(&parent).Error
	if err != nil {
		parent = Menu{
			ParentId: 0,
			Title:    "GPT业务",
			Key:      "gpt-business-root",
			Path:     "",
			Icon:     "🤖",
			Sort:     7,
			Status:   1,
			IsDelete: -1,
		}
		if err := tx.Create(&parent).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&Menu{}).Where("id = ?", parent.Id).Updates(map[string]any{
			"parent_id": 0,
			"title":     "GPT业务",
			"path":      "",
			"icon":      "🤖",
			"sort":      7,
			"status":    1,
			"is_delete": -1,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	pid := parent.Id

	children := []struct {
		key, title, path, icon string
		sort                   int
	}{
		{"gpt-business", "充值与链接", "/admin/gpt-business", "", 1},
		{"gpt-cards", "GPT卡密", "/admin/gpt-cards", "🃏", 2},
		{"gpt-cdk", "GPT-CDK", "/admin/gpt-cdk", "🔑", 3},
		{"gpt-rt-licenses", "GPT RT 许可证", "/admin/gpt-rt-licenses", "📜", 4},
		{"ad-configs", "广告配置", "/admin/ad-configs", "📢", 5},
	}

	for _, c := range children {
		var m Menu
		err := tx.Where("`key` = ? AND is_delete = ?", c.key, -1).First(&m).Error
		if err != nil {
			m = Menu{
				ParentId: pid,
				Title:    c.title,
				Key:      c.key,
				Path:     c.path,
				Icon:     c.icon,
				Sort:     c.sort,
				Status:   1,
				IsDelete: -1,
			}
			if err := tx.Create(&m).Error; err != nil {
				tx.Rollback()
				return err
			}
			continue
		}
		up := map[string]any{
			"parent_id": pid,
			"title":     c.title,
			"path":      c.path,
			"sort":      c.sort,
			"status":    1,
			"is_delete": -1,
		}
		if c.icon != "" {
			up["icon"] = c.icon
		}
		if err := tx.Model(&Menu{}).Where("id = ?", m.Id).Updates(up).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 「系统设置」排序顺延到 GPT业务 之后
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "system", -1).Updates(map[string]any{"sort": 8}).Error

	return tx.Commit().Error
}

// EnsureOutlookMenu 幂等确保「outlook邮箱」父级菜单及「Oauth取件」子菜单
func EnsureOutlookMenu() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var parent Menu
	err := tx.Where("`key` = ? AND is_delete = ?", "outlook-email-root", -1).First(&parent).Error
	if err != nil {
		parent = Menu{
			ParentId: 0,
			Title:    "outlook邮箱",
			Key:      "outlook-email-root",
			Path:     "",
			Icon:     "📧",
			Sort:     8,
			Status:   1,
			IsDelete: -1,
		}
		if err := tx.Create(&parent).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&Menu{}).Where("id = ?", parent.Id).Updates(map[string]any{
			"parent_id": 0,
			"title":     "outlook邮箱",
			"path":      "",
			"icon":      "📧",
			"sort":      8,
			"status":    1,
			"is_delete": -1,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	pid := parent.Id

	var child Menu
	childErr := tx.Where("`key` = ? AND is_delete = ?", "outlook-oauth-fetch", -1).First(&child).Error
	if childErr != nil {
		child = Menu{
			ParentId: pid,
			Title:    "Oauth取件",
			Key:      "outlook-oauth-fetch",
			Path:     "/admin/outlook-oauth-fetch",
			Icon:     "📬",
			Sort:     1,
			Status:   1,
			IsDelete: -1,
		}
		if err := tx.Create(&child).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&Menu{}).Where("id = ?", child.Id).Updates(map[string]any{
			"parent_id": pid,
			"title":     "Oauth取件",
			"path":      "/admin/outlook-oauth-fetch",
			"icon":      "📬",
			"sort":      1,
			"status":    1,
			"is_delete": -1,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 系统设置顺延到 outlook 之后
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "system", -1).Updates(map[string]any{"sort": 10}).Error

	return tx.Commit().Error
}

// EnsureWebMailMenus 幂等确保「lqqq取件」和「toolsvip取件」子菜单挂在 outlook邮箱 父级下
func EnsureWebMailMenus() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 找到 outlook 父菜单
	var parent Menu
	if err := tx.Where("`key` = ? AND is_delete = ?", "outlook-email-root", -1).First(&parent).Error; err != nil {
		tx.Rollback()
		return nil // outlook 父菜单不存在则跳过
	}
	pid := parent.Id

	type childDef struct {
		key  string
		title string
		path  string
		sort  int
	}
	children := []childDef{
		{"lqqq-fetch", "lqqq取件", "/admin/lqqq-fetch", 2},
		{"toolsvip-fetch", "toolsvip取件", "/admin/toolsvip-fetch", 3},
	}

	for _, ch := range children {
		var child Menu
		if err := tx.Where("`key` = ? AND is_delete = ?", ch.key, -1).First(&child).Error; err != nil {
			child = Menu{
				ParentId: pid,
				Title:    ch.title,
				Key:      ch.key,
				Path:     ch.path,
				Icon:     "📨",
				Sort:     ch.sort,
				Status:   1,
				IsDelete: -1,
			}
			if err := tx.Create(&child).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			_ = tx.Model(&Menu{}).Where("id = ?", child.Id).Updates(map[string]any{
				"parent_id": pid,
				"title":     ch.title,
				"path":      ch.path,
				"icon":      "📨",
				"sort":      ch.sort,
				"status":    1,
				"is_delete": -1,
			}).Error
		}
	}

	return tx.Commit().Error
}

// EnsureMicrosoftMailMenus 幂等确保「微软邮箱未售/已售」子菜单挂在 outlook邮箱 父级下
func EnsureMicrosoftMailMenus() error {
	if DB == nil {
		return nil
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var parent Menu
	if err := tx.Where("`key` = ? AND is_delete = ?", "outlook-email-root", -1).First(&parent).Error; err != nil {
		tx.Rollback()
		return nil
	}
	pid := parent.Id

	type childDef struct {
		key   string
		title string
		path  string
		sort  int
		icon  string
	}
	children := []childDef{
		{"microsoft-mail-unsold", "微软邮箱未售", "/admin/microsoft-mails?type=unsold", 4, "📭"},
		{"microsoft-mail-sold", "微软邮箱已售", "/admin/microsoft-mails?type=sold", 5, "📬"},
	}

	for _, ch := range children {
		var child Menu
		if err := tx.Where("`key` = ? AND is_delete = ?", ch.key, -1).First(&child).Error; err != nil {
			child = Menu{
				ParentId: pid,
				Title:    ch.title,
				Key:      ch.key,
				Path:     ch.path,
				Icon:     ch.icon,
				Sort:     ch.sort,
				Status:   1,
				IsDelete: -1,
			}
			if err := tx.Create(&child).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			if err := tx.Model(&Menu{}).Where("id = ?", child.Id).Updates(map[string]any{
				"parent_id": pid,
				"title":     ch.title,
				"path":      ch.path,
				"icon":      ch.icon,
				"sort":      ch.sort,
				"status":    1,
				"is_delete": -1,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
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
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "gpt-business-root", -1).Updates(map[string]any{"sort": 7}).Error
	_ = tx.Model(&Menu{}).Where("`key` = ? AND is_delete = ?", "system", -1).Updates(map[string]any{"sort": 8}).Error

	return tx.Commit().Error
}
