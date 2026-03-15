package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// ========== 随机 Token 生成 ==========

// GenerateRandomToken 生成随机 Token（Base64 编码）
func GenerateRandomToken(length int) (string, error) {
	// 生成随机字节
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 URL 安全编码（可用于 URL）
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ========== 十六进制 Token 生成 ==========
func GenerateRandomHex(length int) (string, error) {
	// 生成随机字节
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 十六进制编码
	return hex.EncodeToString(bytes), nil
}

// ========== Session ID 生成 ==========

// GenerateSessionID 生成 Session ID（64位十六进制）
func GenerateSessionID() (string, error) {
	return GenerateRandomHex(32) // 32 字节 = 256 位 = 64 字符
}

// ========== Refresh Token 生成 ==========

// GenerateRefreshToken 生成 Refresh Token（Base64 编码）
func GenerateRefreshToken() (string, error) {
	return GenerateRandomToken(32) // 32 字节 = 256 位
}

// ========== API Key 生成 ==========

// GenerateAPIKey 生成 API Key（带前缀）
func GenerateAPIKey(prefix string) (string, error) {
	token, err := GenerateRandomHex(24) // 24 字节 = 48 字符
	if err != nil {
		return "", err
	}
	return prefix + token, nil
}

// ========== 验证码生成 ==========

// GenerateVerificationCode 生成数字验证码
func GenerateVerificationCode(length int) (string, error) {
	// 生成随机字节
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 转换为数字（0-9）
	code := make([]byte, length)
	for i, b := range bytes {
		code[i] = '0' + (b % 10) // 确保是 0-9
	}

	return string(code), nil
}

// ========== 邮箱验证 Token ==========

// GenerateEmailVerificationToken 生成邮箱验证 Token
func GenerateEmailVerificationToken() (string, error) {
	return GenerateRandomToken(24) // 24 字节 = 192 位
}

// ========== 重置密码 Token ==========

// GeneratePasswordResetToken 生成密码重置 Token
func GeneratePasswordResetToken() (string, error) {
	return GenerateRandomHex(32) // 32 字节 = 256 位
}

// ========== CSRF Token 生成 ==========

// GenerateCSRFToken 生成 CSRF Token
func GenerateCSRFToken() (string, error) {
	return GenerateRandomHex(16) // 16 字节 = 128 位
}

// ========== State 参数生成（OAuth2） ==========

// GenerateOAuth2State 生成 OAuth2 State 参数
func GenerateOAuth2State() (string, error) {
	return GenerateRandomHex(16) // 16 字节 = 128 位
}

// ========== UUID 风格 Token ==========

// GenerateUUID 生成 UUID 风格的 Token（非标准 UUID）
func GenerateUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// 格式化为 UUID 风格（8-4-4-4-12）
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	), nil
}

// ========== 辅助函数 ==========

// GenerateRandomBytes 生成随机字节（底层方法）
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}
