package crypto

import (
	"fmt"
	"testing"
)

func TestBuildChatShareLoginRequest(t *testing.T) {
	username := "30T9tNPQGxw"
	password := "FOKYie"

	req, err := BuildChatShareLoginRequest(username, password)
	if err != nil {
		t.Fatalf("构建请求失败: %v", err)
	}

	fmt.Printf("用户名: %s\n", req.Username)
	fmt.Printf("加密后的密码: %s\n", req.Password)
	fmt.Printf("时间戳: %s\n", req.Timestamp)
	fmt.Printf("签名: %s\n", req.Sign)

	// 验证字段
	if req.Username != username {
		t.Errorf("用户名不匹配: 期望 %s, 得到 %s", username, req.Username)
	}

	if req.Password == "" {
		t.Error("加密后的密码为空")
	}

	if req.Password == password {
		t.Error("密码没有被加密")
	}

	if req.Timestamp == "" {
		t.Error("时间戳为空")
	}

	// 验证密码可以被解密
	decrypted, err := DecryptPasswordAESCFB(req.Password)
	if err != nil {
		t.Fatalf("解密密码失败: %v", err)
	}

	if decrypted != password {
		t.Errorf("解密后的密码不匹配: 期望 %s, 得到 %s", password, decrypted)
	}
}

func TestBuildChatShareLoginRequest_MultipleTimes(t *testing.T) {
	username := "testuser"
	password := "testpass"

	// 多次构建请求，验证密码每次都不同（因为 IV 随机）
	passwords := make([]string, 3)
	for i := 0; i < 3; i++ {
		req, err := BuildChatShareLoginRequest(username, password)
		if err != nil {
			t.Fatalf("第 %d 次构建请求失败: %v", i+1, err)
		}
		passwords[i] = req.Password
		fmt.Printf("第 %d 次加密结果: %s\n", i+1, req.Password)
	}

	// 验证每次加密结果都不同
	if passwords[0] == passwords[1] || passwords[1] == passwords[2] || passwords[0] == passwords[2] {
		t.Error("多次加密结果应该不同")
	}
}

// 注意: 这个测试会实际调用外部 API，默认跳过
// 要运行此测试，使用: go test -v -run TestCallChatShareLogin
func TestCallChatShareLogin(t *testing.T) {
	t.Skip("跳过实际 API 调用测试，避免依赖外部服务")

	// 如果要测试，取消下面的注释
	/*
		username := "30T9tNPQGxw"
		password := "FOKYie"
		baseURL := "https://node3.chatshare.biz"

		resp, err := CallChatShareLogin(username, password, baseURL)
		if err != nil {
			t.Fatalf("登录失败: %v", err)
		}

		fmt.Printf("登录响应: %+v\n", resp)

		if resp.Code != 200 {
			t.Errorf("登录失败，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
		}
	*/
}
