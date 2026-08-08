package outlook

import (
	"bytes"
	"errors"
	"html"
	"io"
	"mime/quotedprintable"
	"regexp"
	"strings"

	gomail "github.com/emersion/go-message/mail"
)

var (
	sixDigitRe    = regexp.MustCompile(`\b(\d{6})\b`)
	codeKeywordRe = regexp.MustCompile(`(?i)(code|verification|verify|otp|passcode|验证码|安全代码|验证代码|代码)`)
	labeledCodeRe = regexp.MustCompile(`(?i)(?:code(?:\s+is)?|verification(?:\s+code)?|passcode|otp|验证码|安全代码)[：:\s]*([0-9]{4,8})`)
	htmlTagRe     = regexp.MustCompile(`<[^>]+>`)
	wsRe          = regexp.MustCompile(`\s+`)

	// subjectFetchRe 仅用于判断是否需要拉取邮件正文，严格匹配：code/otp/验证码/代码
	subjectFetchRe = regexp.MustCompile(`(?i)\bcode\b|验证码|代码|\botp\b`)
)

// HasCodeKeyword 判断标题是否需要拉取正文
func HasCodeKeyword(subject string) bool {
	return subjectFetchRe.MatchString(subject)
}

// ExtractCode 提取验证码：
// 1. 优先从标题匹配 6 位数字
// 2. 若标题含 code 等关键词但无 6 位数字，则进入正文匹配（先带标签，再裸 6 位数）
func ExtractCode(subject, body string) string {
	if m := sixDigitRe.FindStringSubmatch(subject); len(m) >= 2 {
		return m[1]
	}
	if codeKeywordRe.MatchString(subject) {
		if m := labeledCodeRe.FindStringSubmatch(body); len(m) >= 2 {
			return m[1]
		}
		if m := sixDigitRe.FindStringSubmatch(body); len(m) >= 2 {
			return m[1]
		}
	}
	return ""
}

// ParseBodies 从邮件原始字节解析 plain / html 正文
func ParseBodies(raw []byte) (plain string, htmlBody string) {
	if len(raw) == 0 {
		return "", ""
	}
	if mr, err := gomail.CreateReader(bytes.NewReader(raw)); err == nil {
		plain, htmlBody = readBodies(mr)
	}
	// 兜底：MIME 解析失败或结果为空，手动切掉邮件头后提取正文
	if strings.TrimSpace(plain) == "" && strings.TrimSpace(htmlBody) == "" {
		htmlBody = extractBodyFromRaw(raw)
	}
	return plain, htmlBody
}

// extractBodyFromRaw 从 RFC 2822 原始字节中提取正文部分（跳过邮件头），
// 若 Content-Transfer-Encoding 为 quoted-printable 则自动解码。
func extractBodyFromRaw(raw []byte) string {
	// 找到头部与正文的分隔空行
	body := raw
	headerEnd := -1
	if idx := bytes.Index(raw, []byte("\r\n\r\n")); idx >= 0 {
		headerEnd = idx
		body = raw[idx+4:]
	} else if idx := bytes.Index(raw, []byte("\n\n")); idx >= 0 {
		headerEnd = idx
		body = raw[idx+2:]
	}

	// 若找到了头部，判断是否需要 QP 解码
	if headerEnd > 0 {
		header := bytes.ToLower(raw[:headerEnd])
		if bytes.Contains(header, []byte("quoted-printable")) {
			if decoded, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(body))); err == nil && len(decoded) > 0 {
				return string(decoded)
			}
		}
	}

	return string(body)
}

// readBodies 读取 MIME 邮件各 part 的 plain / html 正文
func readBodies(mr *gomail.Reader) (plain string, htmlBody string) {
	var pb, hb strings.Builder
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		// part 不为空时继续处理（charset 等软错误不中断）
		if part == nil {
			if err != nil {
				break
			}
			continue
		}

		var ct string
		switch h := part.Header.(type) {
		case *gomail.InlineHeader:
			ct, _, _ = h.ContentType()
		case *gomail.AttachmentHeader:
			ct, _, _ = h.ContentType()
		default:
			continue
		}

		b, _ := io.ReadAll(part.Body)
		low := strings.ToLower(ct)
		if strings.HasPrefix(low, "text/plain") {
			pb.Write(b)
			pb.WriteString("\n")
		} else if strings.HasPrefix(low, "text/html") {
			hb.Write(b)
			hb.WriteString("\n")
		}
	}
	return pb.String(), hb.String()
}

func stripHTML(input string) string {
	input = html.UnescapeString(input)
	return htmlTagRe.ReplaceAllString(input, " ")
}

func collapseSpaces(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
