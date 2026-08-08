package outlook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const graphBaseURL = "https://graph.microsoft.com/v1.0"

// FetchViaGraph 通过 Microsoft Graph 拉取收件箱与垃圾箱（IMAP 不可用时的回退）
func FetchViaGraph(httpClient *http.Client, accessTokens []string) (inbox []MailItem, junk []MailItem, err error) {
	var lastErr error
	for _, token := range accessTokens {
		if token == "" {
			continue
		}
		inbox, err = fetchGraphFolder(httpClient, token, "inbox", "INBOX")
		if err != nil {
			lastErr = err
			continue
		}
		junk, junkErr := fetchGraphFolder(httpClient, token, "junkemail", "Junk Email")
		if junkErr != nil {
			// 垃圾箱失败不阻断收件箱结果
			junk = []MailItem{}
		}
		if junk == nil {
			junk = []MailItem{}
		}
		return inbox, junk, nil
	}
	if lastErr != nil {
		return nil, nil, lastErr
	}
	return nil, nil, fmt.Errorf("Graph 取件失败：无可用 access_token")
}

// FetchBodyByGraphID 按 Graph message id 拉取单封正文
func FetchBodyByGraphID(httpClient *http.Client, accessTokens []string, messageID string) (*MailDetail, error) {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return nil, fmt.Errorf("message_id 为空")
	}
	var lastErr error
	for _, token := range accessTokens {
		if token == "" {
			continue
		}
		detail, err := fetchGraphMessageBody(httpClient, token, messageID)
		if err != nil {
			lastErr = err
			continue
		}
		return detail, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("Graph 拉取正文失败：无可用 access_token")
}

type graphMessageList struct {
	Value []graphMessage `json:"value"`
}

type graphMessage struct {
	ID               string `json:"id"`
	Subject          string `json:"subject"`
	BodyPreview      string `json:"bodyPreview"`
	ReceivedDateTime string `json:"receivedDateTime"`
	IsRead           bool   `json:"isRead"`
	From             *struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
	Body *struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
}

func fetchGraphFolder(httpClient *http.Client, token, wellKnown, folderLabel string) ([]MailItem, error) {
	q := url.Values{}
	q.Set("$top", "20")
	q.Set("$orderby", "receivedDateTime desc")
	q.Set("$select", "id,subject,from,receivedDateTime,bodyPreview,isRead,body")
	apiURL := fmt.Sprintf("%s/me/mailFolders/%s/messages?%s", graphBaseURL, wellKnown, q.Encode())

	data, err := graphGET(httpClient, token, apiURL)
	if err != nil {
		return nil, err
	}
	var list graphMessageList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("解析 Graph 邮件列表失败: %w", err)
	}

	items := make([]MailItem, 0, len(list.Value))
	for i, m := range list.Value {
		fromStr := ""
		if m.From != nil {
			name := strings.TrimSpace(m.From.EmailAddress.Name)
			addr := strings.TrimSpace(m.From.EmailAddress.Address)
			if name != "" && addr != "" {
				fromStr = name + " <" + addr + ">"
			} else if addr != "" {
				fromStr = addr
			} else {
				fromStr = name
			}
		}

		plain, htmlBody := "", ""
		if m.Body != nil {
			if strings.EqualFold(m.Body.ContentType, "html") {
				htmlBody = m.Body.Content
				plain = collapseSpaces(stripHTML(htmlBody))
			} else {
				plain = collapseSpaces(m.Body.Content)
			}
		}
		if plain == "" {
			plain = collapseSpaces(m.BodyPreview)
		}
		preview := truncateRunes(plain, 200)
		if preview == "" {
			preview = truncateRunes(m.BodyPreview, 200)
		}
		code := ExtractCode(m.Subject, plain+"\n"+stripHTML(htmlBody)+"\n"+m.BodyPreview)

		items = append(items, MailItem{
			ID:         m.ID,
			SeqNum:     uint32(i + 1),
			Folder:     folderLabel,
			Subject:    m.Subject,
			From:       fromStr,
			ReceivedAt: formatGraphTime(m.ReceivedDateTime),
			Preview:    preview,
			Body:       plain,
			HtmlBody:   htmlBody,
			Code:       code,
			IsRead:     m.IsRead,
		})
	}
	return items, nil
}

func fetchGraphMessageBody(httpClient *http.Client, token, messageID string) (*MailDetail, error) {
	apiURL := fmt.Sprintf("%s/me/messages/%s?$select=subject,body,bodyPreview",
		graphBaseURL, url.PathEscape(messageID))
	data, err := graphGET(httpClient, token, apiURL)
	if err != nil {
		return nil, err
	}
	var m graphMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("解析 Graph 邮件正文失败: %w", err)
	}

	plain, htmlBody := "", ""
	if m.Body != nil {
		if strings.EqualFold(m.Body.ContentType, "html") {
			htmlBody = m.Body.Content
			plain = collapseSpaces(stripHTML(htmlBody))
		} else {
			plain = collapseSpaces(m.Body.Content)
		}
	}
	if plain == "" {
		plain = collapseSpaces(m.BodyPreview)
	}
	code := ExtractCode(m.Subject, plain+"\n"+stripHTML(htmlBody)+"\n"+m.BodyPreview)
	return &MailDetail{
		Body:     plain,
		HtmlBody: htmlBody,
		Code:     code,
	}, nil
}

func graphGET(httpClient *http.Client, token, apiURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("构建 Graph 请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Graph 请求失败: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview := string(data)
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, fmt.Errorf("Graph HTTP %d: %s", resp.StatusCode, preview)
	}
	return data, nil
}

func formatGraphTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.Local().Format("2006-01-02 15:04:05")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local().Format("2006-01-02 15:04:05")
	}
	return s
}
