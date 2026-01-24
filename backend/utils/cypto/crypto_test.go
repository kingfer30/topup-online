package crypto

import (
	"fmt"
	"testing"
)

func TestEncryptPasswordAESCFB(t *testing.T) {
	// 测试加密
	password := "FOKYie"
	
	encrypted, err := EncryptPasswordAESCFB(password)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	
	fmt.Printf("原始密码: %s\n", password)
	fmt.Printf("加密结果: %s\n", encrypted)
	
	// 测试解密
	decrypted, err := DecryptPasswordAESCFB(encrypted)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	
	fmt.Printf("解密结果: %s\n", decrypted)
	
	// 验证解密后的密码是否正确
	if decrypted != password {
		t.Errorf("解密结果不匹配: 期望 %s, 得到 %s", password, decrypted)
	}
}

func TestEncryptPasswordAESCFB_MultipleRuns(t *testing.T) {
	// 测试多次加密，验证每次结果都不同（因为 IV 是随机的）
	password := "FOKYie"
	
	results := make([]string, 5)
	for i := 0; i < 5; i++ {
		encrypted, err := EncryptPasswordAESCFB(password)
		if err != nil {
			t.Fatalf("第 %d 次加密失败: %v", i+1, err)
		}
		results[i] = encrypted
		fmt.Printf("第 %d 次加密: %s\n", i+1, encrypted)
	}
	
	// 验证每次加密结果都不同
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i] == results[j] {
				t.Errorf("第 %d 次和第 %d 次加密结果相同，应该不同", i+1, j+1)
			}
		}
	}
	
	// 验证所有加密结果都能正确解密
	for i, encrypted := range results {
		decrypted, err := DecryptPasswordAESCFB(encrypted)
		if err != nil {
			t.Fatalf("第 %d 次解密失败: %v", i+1, err)
		}
		if decrypted != password {
			t.Errorf("第 %d 次解密结果不匹配: 期望 %s, 得到 %s", i+1, password, decrypted)
		}
	}
}

// 基准测试
func BenchmarkEncryptPasswordAESCFB(b *testing.B) {
	password := "FOKYie"
	for i := 0; i < b.N; i++ {
		_, _ = EncryptPasswordAESCFB(password)
	}
}

