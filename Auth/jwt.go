package Auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

// ========== JWT Header ==========

type jwtHeader struct {
	Alg string `json:"alg"` // 算法类型（HS256）
	Typ string `json:"typ"` // Token 类型（JWT）
}

// ========== JWT 生成 ==========

// GenerateJWT 生成 JWT Token
func GenerateJWT(claims *Claims, secret string, expireAt time.Time) (string, error) {
	// 设置过期时间
	claims.ExpiresAt = expireAt.Unix()
	claims.IssuedAt = time.Now().Unix()

	// 1. 创建 Header
	header := jwtHeader{
		Alg: "HS256",
		Typ: "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// 2. 创建 Payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// 3. 创建签名
	message := headerB64 + "." + payloadB64
	signature := createSignature(message, secret)

	// 4. 拼接 Token
	token := message + "." + signature

	return token, nil
}

// ========== JWT 验证 ==========

// VerifyJWT 验证 JWT Token 并返回 Claims
func VerifyJWT(token, secret string) (*Claims, error) {
	// 1. 分割 Token（格式：header.payload.signature）
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidTokenFormat
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// 2. 验证签名
	message := headerB64 + "." + payloadB64
	expectedSignature := createSignature(message, secret)
	if signatureB64 != expectedSignature {
		return nil, ErrInvalidSignature
	}

	// 3. 解析 Payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// 4. 验证时间
	if !claims.IsValid() {
		if claims.IsExpired() {
			return nil, ErrExpiredToken
		}
		return nil, ErrTokenNotYetValid
	}

	return &claims, nil
}

// ========== 签名工具 ==========

// createSignature 使用 HMAC-SHA256 创建签名
func createSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(signature)
}

// ========== Token 过期时间计算 ==========

// GetTokenExpiresAt 计算 Token 过期时间
func GetTokenExpiresAt(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

// GetRemainingTime 获取 Token 剩余有效时间
func GetRemainingTime(expiresAt int64) time.Duration {
	expireTime := time.Unix(expiresAt, 0)
	remaining := time.Until(expireTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
