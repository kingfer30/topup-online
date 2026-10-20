package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const listenAddr = "127.0.0.1:17321"

type openRequest struct {
	Token    string `json:"token"`
	URL      string `json:"url"`
	Line1    string `json:"line1"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postal   string `json:"postal"`
	Country  string `json:"country"`
	Number   string `json:"number"`
	ExpMonth string `json:"expMonth"`
	ExpYear  string `json:"expYear"`
	CVC      string `json:"cvc"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w, r)
		if r.Method == http.MethodOptions {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/open", handleOpen)
	mux.HandleFunc("/close", handleClose)
	log.Printf("pay-browser listening on http://%s", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleOpen(w http.ResponseWriter, r *http.Request) {
	writeCORS(w, r)
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"message": "只接受 POST"})
		return
	}
	var req openRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "请求无法解析"})
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	req.URL = strings.TrimSpace(req.URL)
	if req.Token == "" || !strings.HasPrefix(req.URL, "https://cursor.com/") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"message": "缺少账号或付款页地址"})
		return
	}
	filled, err := openPayWindow(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"message": err.Error()})
		return
	}
	message := "已打开调试窗口里的付款页，请在页面里完成付款"
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "filled": filled, "message": message})
}

func openPayWindow(req openRequest) (bool, error) {
	browserPath, err := findBrowser()
	if err != nil {
		return false, err
	}
	dataDir, err := os.MkdirTemp("", "pay-browser-")
	if err != nil {
		return false, err
	}
	cmd := exec.Command(browserPath,
		"--user-data-dir="+dataDir,
		"--remote-debugging-port=0",
		"--remote-allow-origins=*",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-sync",
		"about:blank",
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("启动浏览器失败: %w", err)
	}
	endpoint, err := waitDevTools(dataDir, 12*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return false, err
	}

	port := strings.TrimPrefix(endpoint, "http://127.0.0.1:")
	pageWS, err := waitPageSocket(port, 8*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		return false, err
	}
	workosID := req.Token
	if parts := strings.SplitN(req.Token, "::", 2); len(parts) == 2 {
		workosID = parts[0]
	}
	if err := writeSession(pageWS, workosID, req.Token, req.URL); err != nil {
		_ = cmd.Process.Kill()
		return false, err
	}
	rememberVisible(cmd.Process)
	return false, nil
}

var (
	visibleMu   sync.Mutex
	visibleProc *os.Process
)

func rememberVisible(proc *os.Process) {
	visibleMu.Lock()
	visibleProc = proc
	visibleMu.Unlock()
}

func handleClose(w http.ResponseWriter, r *http.Request) {
	writeCORS(w, r)
	if r.Method == http.MethodOptions {
		return
	}
	visibleMu.Lock()
	proc := visibleProc
	visibleProc = nil
	visibleMu.Unlock()
	if proc != nil {
		_ = proc.Kill()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func waitPageSocket(port string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:" + port + "/json/list")
		if err == nil {
			var pages []struct {
				Type                 string `json:"type"`
				WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
			}
			err = json.NewDecoder(resp.Body).Decode(&pages)
			resp.Body.Close()
			if err == nil {
				for _, page := range pages {
					if page.Type == "page" && page.WebSocketDebuggerURL != "" {
						return page.WebSocketDebuggerURL, nil
					}
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("没有找到浏览器标签")
}

func writeSession(wsURL, workosID, token, payURL string) error {
	conn, _, _, err := ws.Dial(context.Background(), wsURL)
	if err != nil {
		return fmt.Errorf("连接浏览器失败: %w", err)
	}
	defer conn.Close()
	if err := cdpCall(conn, 1, "Page.navigate", map[string]any{"url": "https://cursor.com/"}); err != nil {
		return err
	}
	time.Sleep(1200 * time.Millisecond)
	expires := time.Now().Add(30 * 24 * time.Hour).Unix()
	if err := cdpCall(conn, 2, "Network.setCookie", map[string]any{
		"name": "workos_id", "value": workosID, "url": "https://cursor.com/", "domain": ".cursor.com",
		"path": "/", "secure": true, "sameSite": "Lax", "expires": expires,
	}); err != nil {
		return err
	}
	if err := cdpCall(conn, 3, "Network.setCookie", map[string]any{
		"name": "WorkosCursorSessionToken", "value": token, "url": "https://cursor.com/", "domain": ".cursor.com",
		"path": "/", "secure": true, "httpOnly": true, "sameSite": "Lax", "expires": expires,
	}); err != nil {
		return err
	}
	raw, err := cdpCallRaw(conn, 4, "Network.getCookies", map[string]any{"urls": []string{"https://cursor.com/"}})
	if err != nil {
		return err
	}
	if !strings.Contains(string(raw), "WorkosCursorSessionToken") {
		return fmt.Errorf("登录 cookie 没有写入浏览器")
	}
	return cdpCall(conn, 5, "Page.navigate", map[string]any{"url": payURL})
}

func cdpCall(conn interface {
	io.ReadWriter
	SetReadDeadline(time.Time) error
}, id int, method string, params map[string]any) error {
	_, err := cdpCallRaw(conn, id, method, params)
	return err
}

func cdpCallRaw(conn interface {
	io.ReadWriter
	SetReadDeadline(time.Time) error
}, id int, method string, params map[string]any) ([]byte, error) {
	payload, err := json.Marshal(map[string]any{"id": id, "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	if err := wsutil.WriteClientText(conn, payload); err != nil {
		return nil, err
	}
	_ = conn.SetReadDeadline(time.Now().Add(8 * time.Second))
	for {
		raw, err := wsutil.ReadServerText(conn)
		if err != nil {
			return nil, fmt.Errorf("%s 没有响应", method)
		}
		var msg struct {
			ID    int `json:"id"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(raw, &msg) != nil || msg.ID != id {
			continue
		}
		if msg.Error != nil {
			return nil, fmt.Errorf("%s: %s", method, msg.Error.Message)
		}
		return raw, nil
	}
}

func waitDevTools(dataDir string, timeout time.Duration) (string, error) {
	path := filepath.Join(dataDir, "DevToolsActivePort")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(path)
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
			if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
				return "http://127.0.0.1:" + strings.TrimSpace(lines[0]), nil
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("浏览器调试端口没有起来")
}

func findBrowser() (string, error) {
	candidates := []string{
		os.Getenv("CURSOR_PAY_BROWSER"),
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("未找到 Chrome 或 Edge")
}

func writeCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Private-Network", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
	}
}

func writeJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
