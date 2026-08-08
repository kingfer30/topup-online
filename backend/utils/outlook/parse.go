package outlook

import (
	"fmt"
	"strings"
)

// Account 解析后的账号信息
type Account struct {
	Email        string
	Password     string
	RefreshToken string
	ClientID     string
}

// ParseAccountLine 按格式解析账号行（1-5，与注册机 web_ui Outlook 格式一致）
//
//	1: 邮箱|邮箱密码|refresh_token|client_id
//	2: 邮箱----邮箱密码----refresh_token----client_id
//	3: 邮箱----邮箱密码----client_id----refresh_token
//	4: 邮箱----GPT密码----邮箱密码----client_id----refresh_token
//	5: 邮箱----邮箱密码----GPT密码----client_id----refresh_token
func ParseAccountLine(line, format string) (*Account, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("账号行为空")
	}

	format = normalizeAccountFormat(format)
	parts := splitAccountLine(line, format)
	if len(parts) < 2 {
		return nil, fmt.Errorf("格式错误：字段不足")
	}

	acc := &Account{Email: strings.TrimSpace(parts[0])}
	switch format {
	case "4":
		if len(parts) < 3 {
			return nil, fmt.Errorf("格式4 需要至少 3 段（邮箱----GPT密码----邮箱密码）")
		}
		acc.Password = strings.TrimSpace(parts[2])
		if len(parts) > 3 {
			acc.ClientID = strings.TrimSpace(parts[3])
		}
		if len(parts) > 4 {
			acc.RefreshToken = strings.TrimSpace(parts[4])
		}
	case "5":
		acc.Password = strings.TrimSpace(parts[1])
		if len(parts) > 3 {
			acc.ClientID = strings.TrimSpace(parts[3])
		}
		if len(parts) > 4 {
			acc.RefreshToken = strings.TrimSpace(parts[4])
		}
	case "1", "2":
		acc.Password = strings.TrimSpace(parts[1])
		if len(parts) > 2 {
			acc.RefreshToken = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			acc.ClientID = strings.TrimSpace(parts[3])
		}
	default: // 3
		acc.Password = strings.TrimSpace(parts[1])
		if len(parts) > 2 {
			acc.ClientID = strings.TrimSpace(parts[2])
		}
		if len(parts) > 3 {
			acc.RefreshToken = strings.TrimSpace(parts[3])
		}
	}

	if acc.Email == "" {
		return nil, fmt.Errorf("账号（邮箱）不能为空")
	}
	if acc.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token 不能为空")
	}
	if acc.ClientID == "" {
		return nil, fmt.Errorf("client_id 不能为空")
	}

	return acc, nil
}

func splitAccountLine(line, format string) []string {
	if format == "1" {
		return strings.Split(line, "|")
	}
	return strings.Split(line, "----")
}

func normalizeAccountFormat(f string) string {
	switch f {
	case "f5", "1":
		return "1"
	case "f4", "2":
		return "2"
	case "f1", "3":
		return "3"
	case "f2", "4":
		return "4"
	case "f3", "5":
		return "5"
	case "":
		return "1"
	default:
		return "1"
	}
}
