package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// EncryptPasswordAESCFB 使用 AES-CFB 模式加密密码
// 用于第三方接口调用的密码加密（如 ChatShare 登录）
// 返回 Base64 编码的加密字符串（包含随机IV + 密文）
func EncryptPasswordAESCFB(password string) (string, error) {
	// 固定密钥（32字节的字符串）
	keyString := "0a1b2c3d4e5f6071829aabbccddee123"
	key := []byte(keyString)

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %v", err)
	}

	// 生成随机的 16 字节 IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("生成随机 IV 失败: %v", err)
	}

	// 创建 CFB 加密器
	stream := cipher.NewCFBEncrypter(block, iv)

	// 加密密码
	plaintext := []byte(password)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// 拼接 IV + 密文
	result := append(iv, ciphertext...)

	// Base64 编码
	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}

// DecryptPasswordAESCFB 使用 AES-CFB 模式解密密码
// 用于解密第三方接口返回的加密密码
func DecryptPasswordAESCFB(encryptedBase64 string) (string, error) {
	// 固定密钥
	keyString := "0a1b2c3d4e5f6071829aabbccddee123"
	key := []byte(keyString)

	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %v", err)
	}

	// 检查数据长度
	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("加密数据长度不足")
	}

	// 提取 IV 和密文
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES cipher 失败: %v", err)
	}

	// 创建 CFB 解密器
	stream := cipher.NewCFBDecrypter(block, iv)

	// 解密
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
