package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ========== 密码加密参数 ==========

const (
	saltLength = 16        // 盐值长度（字节）
	keyLength  = 32        // 密钥长度（字节）
	time       = 1         // 时间成本（迭代次数）
	memory     = 64 * 1024 // 内存成本（KB）= 64MB
	threads    = 4         // 并行度（线程数）
)

// ========== 错误定义 ==========

var (
	// ErrInvalidHash Hash 格式无效
	ErrInvalidHash = errors.New("invalid hash format")

	// ErrIncompatibleVersion Argon2 版本不兼容
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")

	// ErrPasswordTooShort 密码太短
	ErrPasswordTooShort = errors.New("password too short")
)

// ========== 密码加密 ==========

// HashPassword 使用 Argon2id 加密密码
func HashPassword(password string) (string, error) {
	// 验证密码长度
	if len(password) < 1 {
		return "", ErrPasswordTooShort
	}

	// 生成随机盐值
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用 Argon2id 加密
	// Argon2id = Argon2i + Argon2d 的组合，最安全
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLength)

	// 编码为 Base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 拼接为标准格式
	// 格式：$argon2id$v=版本$m=内存,t=时间,p=线程$盐值$哈希值
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, time, threads, b64Salt, b64Hash)

	return encodedHash, nil
}

// ========== 密码验证 ==========

// VerifyPassword 验证密码是否正确
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析 hash 字符串
	salt, hash, params, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同参数对输入密码加密
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.threads,
		params.keyLength,
	)

	// 使用 constant-time comparison 防止时序攻击
	// subtle.ConstantTimeCompare 会花费固定时间，攻击者无法通过响应时间判断密码相似度
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// ========== 辅助结构 ==========

// params Argon2 参数
type params struct {
	memory    uint32 // 内存成本（KB）
	time      uint32 // 时间成本（迭代次数）
	threads   uint8  // 并行度（线程数）
	keyLength uint32 // 密钥长度（字节）
}

// ========== 解析 Hash ==========

// decodeHash 解析加密后的密码字符串
func decodeHash(encodedHash string) (salt, hash []byte, p *params, err error) {
	// 按 $ 分割字符串
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	// 验证算法类型
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	// 解析版本号
	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	// 解析参数（m=内存,t=时间,p=线程）
	p = &params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse params: %w", err)
	}

	// 解码盐值（Base64）
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}
	p.keyLength = uint32(len(salt))

	// 解码哈希值（Base64）
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return salt, hash, p, nil
}

// ========== 密码强度验证 ==========

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(
	password string,
	minLength int,
	requireUpper, requireLower, requireNumber, requireSpecial bool,
) error {
	// 检查长度
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters", minLength)
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// 检查字符类型
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case isSpecialChar(char):
			hasSpecial = true
		}
	}

	// 验证要求
	if requireUpper && !hasUpper {
		return errors.New("password must contain uppercase letter")
	}
	if requireLower && !hasLower {
		return errors.New("password must contain lowercase letter")
	}
	if requireNumber && !hasNumber {
		return errors.New("password must contain number")
	}
	if requireSpecial && !hasSpecial {
		return errors.New("password must contain special character")
	}

	return nil
}

// isSpecialChar 判断是否是特殊字符
func isSpecialChar(char rune) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	for _, sc := range specialChars {
		if char == sc {
			return true
		}
	}
	return false
}
