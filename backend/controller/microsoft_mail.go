package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kingfer30/topup-online/model"
	"github.com/kingfer30/topup-online/scheduler"
	"github.com/kingfer30/topup-online/utils/logger"
)

// MicrosoftMailRequest 微软邮箱请求
type MicrosoftMailRequest struct {
	Account          string   `json:"account" binding:"required"`
	Password         string   `json:"password"`
	PurchaseDate     *int64   `json:"purchase_date"`
	PurchasePrice    *float64 `json:"purchase_price"`
	PurchaseFrom     string   `json:"purchase_from"`
	PurchaseBy       string   `json:"purchase_by"`
	SellPrice        *float64 `json:"sell_price"`
	SellDate         *int64   `json:"sell_date"`
	SellTo           string   `json:"sell_to"`
	SellOrderNo      string   `json:"sell_order_no"`
	SellStatus       int      `json:"sell_status"`
	Status           int      `json:"status"`
	Token            string   `json:"token"`
	TwoFA            string   `json:"2fa"`
	ClientId         string   `json:"client_id"`
	MailUrl          string   `json:"mail_url"`
	Remark           string   `json:"remark"`
	AccountCardId    *int     `json:"account_card_id"`
	AccountCardTable string   `json:"account_card_table"`
}

func parseMsMailPurchaseDate(raw string) int64 {
	if raw == "" {
		return 0
	}
	cst := time.FixedZone("CST", 8*3600)
	if ts, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if ts > 1e12 {
			ts = ts / 1000
		}
		t := time.Unix(ts, 0).In(cst)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
	}
	if t, err := time.ParseInLocation("2006-01-02", raw, cst); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", raw, cst); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, cst).Unix()
	}
	return 0
}

func toMicrosoftMail(req MicrosoftMailRequest) model.MicrosoftMail {
	mail := model.MicrosoftMail{
		Account:          req.Account,
		Password:         req.Password,
		PurchaseDate:     req.PurchaseDate,
		PurchasePrice:    req.PurchasePrice,
		PurchaseFrom:     req.PurchaseFrom,
		PurchaseBy:       req.PurchaseBy,
		SellPrice:        req.SellPrice,
		SellDate:         req.SellDate,
		SellTo:           req.SellTo,
		SellOrderNo:      req.SellOrderNo,
		SellStatus:       req.SellStatus,
		Status:           req.Status,
		Token:            req.Token,
		TwoFA:            req.TwoFA,
		ClientId:         req.ClientId,
		MailUrl:          req.MailUrl,
		Remark:           req.Remark,
		AccountCardId:    req.AccountCardId,
		AccountCardTable: req.AccountCardTable,
	}
	if mail.SellStatus == 0 {
		mail.SellStatus = 1
	}
	if mail.Status == 0 {
		mail.Status = model.CardStatusNormal
	}
	if mail.IsCheck == 0 {
		mail.IsCheck = -1
	}
	if mail.FreezeStatus == 0 {
		mail.FreezeStatus = -1
	}
	return mail
}

// GetMicrosoftMailList 列表
func GetMicrosoftMailList(c *gin.Context) {
	listType := c.DefaultQuery("type", "unsold")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	accounts := parseAccountsFromQuery(c)
	isCheck, _ := strconv.Atoi(c.DefaultQuery("is_check", "0"))
	purchaseDate := parseMsMailPurchaseDate(c.Query("purchase_date"))
	sellTo := strings.TrimSpace(c.Query("sell_to"))
	purchaseBy := strings.TrimSpace(c.Query("purchase_by"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	mails, total, err := model.GetMicrosoftMailList(listType, page, pageSize, accounts, isCheck, purchaseDate, sellTo, purchaseBy)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "获取列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"list":      mails,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetMicrosoftMailById 详情
func GetMicrosoftMailById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	mail, err := model.GetMicrosoftMailById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": mail})
}

// CreateMicrosoftMail 新增
func CreateMicrosoftMail(c *gin.Context) {
	var req MicrosoftMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	mail := toMicrosoftMail(req)
	if err := model.CreateMicrosoftMail(&mail); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": mail})
}

// UpdateMicrosoftMail 编辑
func UpdateMicrosoftMail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	var req MicrosoftMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	mail := toMicrosoftMail(req)
	if err := model.UpdateMicrosoftMail(id, &mail); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteMicrosoftMail 删除
func DeleteMicrosoftMail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	if err := model.DeleteMicrosoftMail(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// BatchImportMicrosoftMails 批量导入
func BatchImportMicrosoftMails(c *gin.Context) {
	var req struct {
		Mails            []MicrosoftMailRequest `json:"mails" binding:"required"`
		AccountCardTable string                 `json:"account_card_table"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	cardTable := strings.TrimSpace(req.AccountCardTable)
	matched := 0
	mails := make([]model.MicrosoftMail, 0, len(req.Mails))
	for _, item := range req.Mails {
		if strings.TrimSpace(item.Account) == "" {
			continue
		}
		mail := toMicrosoftMail(item)
		if cardTable != "" {
			mail.AccountCardId = nil
			mail.AccountCardTable = ""
			if id, ok, err := model.FindCardIdByAccount(cardTable, mail.Account); err == nil && ok {
				mail.AccountCardId = &id
				mail.AccountCardTable = cardTable
				matched++
			}
		}
		mails = append(mails, mail)
	}
	if err := model.BatchCreateMicrosoftMails(mails); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "批量导入失败: " + err.Error()})
		return
	}
	msg := "批量导入成功"
	if cardTable != "" {
		msg = fmt.Sprintf("批量导入成功，成功关联 %d 条卡密记录", matched)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": msg})
}

// GetMicrosoftMailByCard 根据所属卡密表+ID 查询关联的微软邮箱记录
func GetMicrosoftMailByCard(c *gin.Context) {
	table := strings.TrimSpace(c.Query("table"))
	cardId, _ := strconv.Atoi(c.Query("card_id"))
	if table == "" || cardId == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	mail, err := model.GetMicrosoftMailByCard(table, cardId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "未找到关联的微软邮箱记录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": mail})
}

// ExportMicrosoftMails 导出
func ExportMicrosoftMails(c *gin.Context) {
	listType := c.DefaultQuery("type", "unsold")
	accounts := parseAccountsFromQuery(c)
	isCheck, _ := strconv.Atoi(c.DefaultQuery("is_check", "0"))
	purchaseDate := parseMsMailPurchaseDate(c.Query("purchase_date"))
	sellTo := strings.TrimSpace(c.Query("sell_to"))
	purchaseBy := strings.TrimSpace(c.Query("purchase_by"))

	mails, err := model.GetAllMicrosoftMailsForExport(listType, accounts, isCheck, purchaseDate, sellTo, purchaseBy)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "导出失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": mails})
}

// PickupMicrosoftMail 取货
func PickupMicrosoftMail(c *gin.Context) {
	var req struct {
		Format string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	mail, err := model.PickupMicrosoftMail(req.Format)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "取货成功", "data": mail})
}

// CompleteMicrosoftMailPickup 完成取货
func CompleteMicrosoftMailPickup(c *gin.Context) {
	var req struct {
		Id        int      `json:"id" binding:"required"`
		SellPrice *float64 `json:"sell_price"`
		SellTo    string   `json:"sell_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := model.CompleteMicrosoftMailPickup(req.Id, req.SellPrice, req.SellTo); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "完成取货失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "完成取货成功"})
}

// RollbackMicrosoftMailPickup 回滚发货中
func RollbackMicrosoftMailPickup(c *gin.Context) {
	var req struct {
		Id int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := model.RollbackMicrosoftMailPickup(req.Id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "回滚失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "回滚成功"})
}

// RollbackMicrosoftMailSold 回滚已售
func RollbackMicrosoftMailSold(c *gin.Context) {
	var req struct {
		Id int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := model.RollbackMicrosoftMailSold(req.Id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "回滚失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "回滚成功"})
}

// BatchPickupMicrosoftMails 批量取货
func BatchPickupMicrosoftMails(c *gin.Context) {
	var req struct {
		IDs       []int    `json:"ids" binding:"required"`
		SellPrice *float64 `json:"sell_price"`
		SellTo    string   `json:"sell_to"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	affected, err := model.BatchPickupMicrosoftMails(req.IDs, req.SellPrice, req.SellTo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "批量取货失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功取货 %d 条", affected),
		"data":    affected,
	})
}

// BatchCheckMicrosoftMails 批量检查 refresh_token
func BatchCheckMicrosoftMails(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择要检查的记录"})
		return
	}
	mails, err := model.GetMicrosoftMailsByIds(req.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
		return
	}
	go func() {
		successCount := 0
		failCount := 0
		for _, mail := range mails {
			if err := scheduler.CheckSingleMicrosoftMail(mail); err != nil {
				logger.SysError(fmt.Sprintf("批量检查微软邮箱 [ID:%d, Account:%s] 失败: %v", mail.Id, mail.Account, err))
				failCount++
			} else {
				successCount++
			}
			time.Sleep(500 * time.Millisecond)
		}
		logger.SysLog(fmt.Sprintf("微软邮箱批量检查完成，成功: %d, 失败: %d", successCount, failCount))
	}()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("已提交 %d 条检查任务，正在后台执行", len(mails)),
		"data":    len(mails),
	})
}

// BatchDeleteMicrosoftMails 批量删除
func BatchDeleteMicrosoftMails(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	affected, err := model.BatchDeleteMicrosoftMails(req.IDs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功删除 %d 条记录", affected),
		"data":    affected,
	})
}

// UpdateMicrosoftMailRemark 更新备注
func UpdateMicrosoftMailRemark(c *gin.Context) {
	var req struct {
		Id     int    `json:"id" binding:"required"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := model.UpdateMicrosoftMailRemark(req.Id, req.Remark); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "更新备注失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "备注已更新"})
}
