package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecCHUAMatchesCapturedBrowsers(t *testing.T) {
	chrome := chromeProfile(144)
	if chrome.SecCHUA != `"Not(A:Brand";v="8", "Chromium";v="144", "Google Chrome";v="144"` {
		t.Fatalf("chrome 144 sec-ch-ua = %s", chrome.SecCHUA)
	}
	if !strings.Contains(chrome.UserAgent, "Chrome/144.0.0.0") || strings.Contains(chrome.UserAgent, "Edg/") {
		t.Fatalf("chrome ua = %s", chrome.UserAgent)
	}

	edge := edgeProfile(145)
	if edge.SecCHUA != `"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"` {
		t.Fatalf("edge 145 sec-ch-ua = %s", edge.SecCHUA)
	}
	if !strings.Contains(edge.UserAgent, "Chrome/145.0.0.0") || !strings.Contains(edge.UserAgent, "Edg/145.0.0.0") {
		t.Fatalf("edge ua = %s", edge.UserAgent)
	}
}

func TestBrowserProfilesStayConsistent(t *testing.T) {
	if len(browserProfiles) < 4 {
		t.Fatalf("profiles = %d", len(browserProfiles))
	}
	for _, profile := range browserProfiles {
		if profile.SecCHUAMobile != "?0" || profile.SecCHUAPlatform != `"Windows"` {
			t.Fatalf("%s platform = %s %s", profile.Name, profile.SecCHUAMobile, profile.SecCHUAPlatform)
		}
		if strings.Contains(profile.Name, "Edge") {
			if !strings.Contains(profile.SecCHUA, "Microsoft Edge") || strings.Contains(profile.SecCHUA, "Google Chrome") {
				t.Fatalf("%s sec-ch-ua = %s", profile.Name, profile.SecCHUA)
			}
		} else if !strings.Contains(profile.SecCHUA, "Google Chrome") {
			t.Fatalf("%s sec-ch-ua = %s", profile.Name, profile.SecCHUA)
		}
		version := strings.TrimPrefix(profile.Name, "Chrome ")
		version = strings.TrimPrefix(version, "Edge ")
		if !strings.Contains(profile.UserAgent, "Chrome/"+version+".") || !strings.Contains(profile.SecCHUA, `"`+version+`"`) {
			t.Fatalf("%s headers diverged ua=%s sec=%s", profile.Name, profile.UserAgent, profile.SecCHUA)
		}
	}
}

func TestBrowserTransportUsesOneProfile(t *testing.T) {
	var gotUA, gotCH string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotCH = r.Header.Get("Sec-Ch-Ua")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	profile := chromeProfile(146)
	client := &http.Client{Transport: &browserTransport{base: server.Client().Transport, profile: profile}}
	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("body"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("user-agent", "fixed-client")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if gotUA != profile.UserAgent || gotCH != profile.SecCHUA {
		t.Fatalf("ua=%s sec=%s", gotUA, gotCH)
	}
	if req.Header.Get("User-Agent") != "fixed-client" {
		t.Fatal("original request header was rewritten")
	}
}
