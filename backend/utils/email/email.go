package email

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/kingfer30/topup-online/utils/logger"
)

// SendAlertEmail 发送告警邮件给管理员
func SendAlertEmail(subject, body string) error {
	// 从环境变量获取邮件配置
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	// 如果没有配置邮件，记录日志但不报错
	if smtpHost == "" || smtpUser == "" || smtpPass == "" || adminEmail == "" {
		logger.SysLog("邮件配置不完整，无法发送告警邮件。请配置 SMTP_HOST, SMTP_USER, SMTP_PASS, ADMIN_EMAIL")
		logger.SysLog("告警内容: " + subject + " - " + body)
		return nil
	}

	if smtpPort == "" {
		smtpPort = "587"
	}

	// 构建邮件内容
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", adminEmail, subject, body))

	// 配置认证信息
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// 发送邮件
	addr := smtpHost + ":" + smtpPort
	to := strings.Split(adminEmail, ",")
	err := smtp.SendMail(addr, auth, smtpUser, to, message)
	if err != nil {
		logger.SysLog("发送告警邮件失败: " + err.Error())
		return err
	}

	logger.SysLog("告警邮件已发送到: " + adminEmail)
	return nil
}
