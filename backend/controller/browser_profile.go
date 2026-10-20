package controller

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// browserProfile 是一套互相匹配的桌面浏览器请求头。同一会话里的请求共用一套。
type browserProfile struct {
	Name            string
	UserAgent       string
	SecCHUA         string
	SecCHUAMobile   string
	SecCHUAPlatform string
	AcceptLanguage  string
}

func (p browserProfile) apply(header http.Header) {
	header.Set("user-agent", p.UserAgent)
	header.Set("sec-ch-ua", p.SecCHUA)
	header.Set("sec-ch-ua-mobile", p.SecCHUAMobile)
	header.Set("sec-ch-ua-platform", p.SecCHUAPlatform)
	header.Set("accept-language", p.AcceptLanguage)
}

// 固定的最近桌面 Chrome / Edge。sec-ch-ua 用 Chromium 的 GREASE 规则生成，版本、品牌、平台保持一致。
var browserProfiles = []browserProfile{
	chromeProfile(144),
	chromeProfile(145),
	chromeProfile(146),
	edgeProfile(145),
	edgeProfile(146),
}

func pickBrowserProfile() browserProfile {
	buf := make([]byte, 1)
	if _, err := rand.Read(buf); err != nil || len(browserProfiles) == 0 {
		return browserProfiles[0]
	}
	return browserProfiles[int(buf[0])%len(browserProfiles)]
}

func chromeProfile(major int) browserProfile {
	version := strconv.Itoa(major)
	return browserProfile{
		Name:            "Chrome " + version,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + ".0.0.0 Safari/537.36",
		SecCHUA:         secCHUA(major, "Google Chrome"),
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Windows"`,
		AcceptLanguage:  "en-US,en;q=0.9",
	}
}

func edgeProfile(major int) browserProfile {
	version := strconv.Itoa(major)
	return browserProfile{
		Name:            "Edge " + version,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + ".0.0.0 Safari/537.36 Edg/" + version + ".0.0.0",
		SecCHUA:         secCHUA(major, "Microsoft Edge"),
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Windows"`,
		AcceptLanguage:  "en-US,en;q=0.9",
	}
}

// secCHUA 复刻 Chromium GenerateBrandVersionList / GetGreasedUserAgentBrandVersion。
func secCHUA(major int, product string) string {
	chars := []string{" ", "(", ":", "-", ".", "/", ")", ";", "=", "?", "_"}
	versions := []string{"8", "99", "24"}
	grease := "Not" + chars[major%len(chars)] + "A" + chars[(major+1)%len(chars)] + "Brand"
	greaseVersion := versions[major%len(versions)]
	version := strconv.Itoa(major)
	brands := []string{
		fmt.Sprintf(`"%s";v="%s"`, grease, greaseVersion),
		fmt.Sprintf(`"Chromium";v="%s"`, version),
		fmt.Sprintf(`"%s";v="%s"`, product, version),
	}
	orders := [6][3]int{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 0, 1},
		{2, 1, 0},
	}
	order := orders[major%6]
	shuffled := make([]string, len(brands))
	for i, brand := range brands {
		shuffled[order[i]] = brand
	}
	return strings.Join(shuffled, ", ")
}

// browserTransport 给这一次会话的全部请求盖上同一套浏览器头。
type browserTransport struct {
	base    http.RoundTripper
	profile browserProfile
}

func (t *browserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	t.profile.apply(clone.Header)
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

func withBrowserProfile(base http.RoundTripper) (http.RoundTripper, browserProfile) {
	profile := pickBrowserProfile()
	return &browserTransport{base: base, profile: profile}, profile
}
