package Auth

import (
	"time"

	"github.com/wangjibin555/midware/Auth/crypto"
)

// ========== Session 存储接口 ==========

// SessionStore Session 存储接口
type SessionStore interface {
	// Save 保存 Session
	Save(session *Session) error

	// Get 获取 Session
	Get(sessionID string) (*Session, error)

	// Delete 删除 Session
	Delete(sessionID string) error

	// DeleteByUserID 删除用户的所有 Session（强制登出）
	DeleteByUserID(userID string) error

	// Refresh 刷新 Session 过期时间
	Refresh(sessionID string, duration time.Duration) error
}

// ========== Session 管理 ==========

// GenerateSession 生成 Session
func (a *Auth) GenerateSession(user *User, ip, userAgent string) (*Session, error) {
	sessionID, err := crypto.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Data:      make(map[string]interface{}),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(a.config.SessionExpire),
		IP:        ip,
		UserAgent: userAgent,
	}, nil
}

// ValidateSession 验证 Session 是否有效
func (a *Auth) ValidateSession(sessionID string, store SessionStore) (*Session, error) {
	// 1. 从存储中获取 Session
	session, err := store.Get(sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// 2. 验证 Session 是否有效
	if !session.IsValid() {
		if session.IsExpired() {
			return nil, ErrSessionExpired
		}
		return nil, ErrSessionInvalid
	}

	return session, nil
}

// RefreshSession 刷新 Session 过期时间
func (a *Auth) RefreshSession(sessionID string, store SessionStore) error {
	return store.Refresh(sessionID, a.config.SessionExpire)
}

// DestroySession 销毁 Session（登出）
func (a *Auth) DestroySession(sessionID string, store SessionStore) error {
	return store.Delete(sessionID)
}

// DestroyAllUserSessions 销毁用户的所有 Session（强制登出所有设备）
func (a *Auth) DestroyAllUserSessions(userID string, store SessionStore) error {
	return store.DeleteByUserID(userID)
}

// ========== API Key 管理 ==========

// GenerateAPIKey 生成 API Key
func (a *Auth) GenerateAPIKey(prefix string) (string, error) {
	if prefix == "" {
		prefix = "sk_" // 默认前缀
	}
	return crypto.GenerateAPIKey(prefix)
}

// ValidateAPIKey 验证 API Key（需要应用层实现存储和验证逻辑）
// 这里只提供生成功能，验证逻辑应该在应用层的 APIKeyStore 中实现
func (a *Auth) ValidateAPIKey(apiKey string) (*User, error) {
	// 提示：应用层需要实现：
	// 1. 从数据库查询 API Key
	// 2. 验证是否有效（未过期、未撤销）
	// 3. 返回关联的用户信息
	return nil, ErrNotImplemented
}

// ========== 验证码管理 ==========

// GenerateVerificationCode 生成数字验证码（如邮箱验证、手机验证）
func (a *Auth) GenerateVerificationCode(length int) (string, error) {
	if length <= 0 {
		length = 6 // 默认6位
	}
	return crypto.GenerateVerificationCode(length)
}

// ========== 密码重置 Token ==========

// GeneratePasswordResetToken 生成密码重置 Token
func (a *Auth) GeneratePasswordResetToken() (string, error) {
	return crypto.GeneratePasswordResetToken()
}

// ValidatePasswordResetToken 验证密码重置 Token
// 需要应用层配合存储实现（通常存储在 Redis，15分钟过期）
func (a *Auth) ValidatePasswordResetToken(token string) (userID string, err error) {
	// 提示：应用层需要实现：
	// 1. 从 Redis 查询 token → userID 的映射
	// 2. 验证是否过期
	// 3. 返回用户ID
	return "", ErrNotImplemented
}

// ========== 邮箱验证 Token ==========

// GenerateEmailVerificationToken 生成邮箱验证 Token
func (a *Auth) GenerateEmailVerificationToken() (string, error) {
	return crypto.GenerateEmailVerificationToken()
}

// ValidateEmailVerificationToken 验证邮箱验证 Token
// 需要应用层配合存储实现（通常存储在 Redis，1小时过期）
func (a *Auth) ValidateEmailVerificationToken(token string) (email string, err error) {
	// 提示：应用层需要实现：
	// 1. 从 Redis 查询 token → email 的映射
	// 2. 验证是否过期
	// 3. 返回邮箱地址
	return "", ErrNotImplemented
}

// ========== CSRF Token 管理 ==========

// GenerateCSRFToken 生成 CSRF Token
func (a *Auth) GenerateCSRFToken() (string, error) {
	return crypto.GenerateCSRFToken()
}

// ValidateCSRFToken 验证 CSRF Token（通常与 Session 或 Cookie 配合使用）
func (a *Auth) ValidateCSRFToken(token, expected string) bool {
	// 使用恒定时间比较，防止时序攻击
	if len(token) != len(expected) {
		return false
	}
	var result byte
	for i := 0; i < len(token); i++ {
		result |= token[i] ^ expected[i]
	}
	return result == 0
}

// ========== OAuth2 State 管理 ==========

// GenerateOAuth2State 生成 OAuth2 State 参数（防止 CSRF 攻击）
func (a *Auth) GenerateOAuth2State() (string, error) {
	return crypto.GenerateOAuth2State()
}

// ========== UUID 生成 ==========

// GenerateUUID 生成 UUID 风格的随机 Token
func (a *Auth) GenerateUUID() (string, error) {
	return crypto.GenerateUUID()
}

// ========== 空实现（用于测试） ==========

// NoopSessionStore Session 存储的空实现
type NoopSessionStore struct{}

func (s *NoopSessionStore) Save(session *Session) error {
	return ErrNotImplemented
}

func (s *NoopSessionStore) Get(sessionID string) (*Session, error) {
	return nil, ErrNotImplemented
}

func (s *NoopSessionStore) Delete(sessionID string) error {
	return ErrNotImplemented
}

func (s *NoopSessionStore) DeleteByUserID(userID string) error {
	return ErrNotImplemented
}

func (s *NoopSessionStore) Refresh(sessionID string, duration time.Duration) error {
	return ErrNotImplemented
}

// ========== 使用示例 ==========

/*
使用示例：

// ===== 1. Session 管理 =====

// 登录时创建 Session
session, err := auth.GenerateSession(user, r.RemoteAddr, r.UserAgent())
sessionStore.Save(session)
http.SetCookie(w, &http.Cookie{
    Name:  "session_id",
    Value: session.SessionID,
    MaxAge: int(auth.config.SessionExpire.Seconds()),
    HttpOnly: true,
    Secure: true,
})

// 验证 Session
sessionID := getSessionIDFromCookie(r)
session, err := auth.ValidateSession(sessionID, sessionStore)
if err != nil {
    // Session 无效
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

// 刷新 Session
auth.RefreshSession(sessionID, sessionStore)

// 登出
auth.DestroySession(sessionID, sessionStore)

// 强制登出所有设备
auth.DestroyAllUserSessions(userID, sessionStore)

// ===== 2. API Key 管理 =====

// 生成 API Key
apiKey, _ := auth.GenerateAPIKey("sk_")
// 结果: sk_a3f2c9b8e1d4f7a2c9b8e1d4f7a2c9b8e1d4f7a2c9b8

// 存储到数据库
user.APIKey = apiKey
db.Save(&user)

// 验证 API Key（HTTP 请求）
apiKey := r.Header.Get("X-API-Key")
user, err := apiKeyStore.ValidateAPIKey(apiKey)

// ===== 3. 邮箱验证 =====

// 生成验证 Token
verifyToken, _ := auth.GenerateEmailVerificationToken()

// 存储到 Redis（1小时过期）
redis.Set("email_verify:"+verifyToken, user.Email, 1*time.Hour)

// 发送邮件
sendEmail(user.Email, "https://example.com/verify?token="+verifyToken)

// 用户点击链接后验证
token := r.URL.Query().Get("token")
email, _ := redis.Get("email_verify:" + token).Result()
if email != "" {
    // 验证成功，更新用户状态
    user.EmailVerified = true
    db.Save(&user)
    redis.Del("email_verify:" + token)
}

// ===== 4. 密码重置 =====

// 生成重置 Token
resetToken, _ := auth.GeneratePasswordResetToken()

// 存储到 Redis（15分钟过期）
redis.Set("password_reset:"+resetToken, user.ID, 15*time.Minute)

// 发送邮件
sendEmail(user.Email, "https://example.com/reset?token="+resetToken)

// 用户点击链接后验证
token := r.URL.Query().Get("token")
userID, _ := redis.Get("password_reset:" + token).Result()
if userID != "" {
    // 验证成功，允许重置密码
    newPassword := r.FormValue("password")
    hash, _ := auth.HashPassword(newPassword)
    db.Model(&User{}).Where("id = ?", userID).Update("password_hash", hash)
    redis.Del("password_reset:" + token)
}

// ===== 5. CSRF 防护 =====

// 生成 CSRF Token
csrfToken, _ := auth.GenerateCSRFToken()

// 存储到 Cookie
http.SetCookie(w, &http.Cookie{
    Name:  "csrf_token",
    Value: csrfToken,
    HttpOnly: false, // 前端需要读取
    Secure: true,
})

// 验证 CSRF Token
tokenFromHeader := r.Header.Get("X-CSRF-Token")
tokenFromCookie, _ := r.Cookie("csrf_token")
if !auth.ValidateCSRFToken(tokenFromHeader, tokenFromCookie.Value) {
    http.Error(w, "CSRF Token Invalid", http.StatusForbidden)
    return
}

// ===== 6. OAuth2 State =====

// 生成 State
state, _ := auth.GenerateOAuth2State()

// 存储到 Session
session.Set("oauth2_state", state)

// 跳转到 OAuth2 授权页面
authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s", clientID, state)
http.Redirect(w, r, authURL, http.StatusFound)

// OAuth2 回调验证
stateFromQuery := r.URL.Query().Get("state")
stateFromSession := session.Get("oauth2_state")
if stateFromQuery != stateFromSession {
    http.Error(w, "Invalid OAuth2 State", http.StatusForbidden)
    return
}

// ===== 7. 验证码 =====

// 生成6位数字验证码
code, _ := auth.GenerateVerificationCode(6)
// 结果: 123456

// 发送短信
sendSMS(user.Phone, "Your verification code is: "+code)

// 存储到 Redis（5分钟过期）
redis.Set("sms_code:"+user.Phone, code, 5*time.Minute)

// 用户输入验证码后验证
inputCode := r.FormValue("code")
storedCode, _ := redis.Get("sms_code:" + user.Phone).Result()
if inputCode == storedCode {
    // 验证成功
    redis.Del("sms_code:" + user.Phone)
}
*/
