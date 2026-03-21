package Auth

import "time"

const (
	UserStatusNormal   = 1 // 用户状态：正常
	UserStatusDisabled = 2 // 用户状态：禁用
	UserStatusDeleted  = 3 // 用户状态：已删除
)

// Claims JWT 声明（Token 中包含的信息）
type Claims struct {
	UserID      string                 `json:"user_id"`         // 用户唯一ID
	Username    string                 `json:"username"`        // 用户名
	Roles       []string               `json:"roles"`           // 用户角色列表
	Permissions []string               `json:"permissions"`     // 用户权限列表
	Extra       map[string]interface{} `json:"extra,omitempty"` // 扩展字段（如 tenant_id）
	IssuedAt    int64                  `json:"iat"`             // 签发时间（Unix 时间戳）
	ExpiresAt   int64                  `json:"exp"`             // 过期时间（Unix 时间戳）
	NotBefore   int64                  `json:"nbf,omitempty"`   // 生效时间（可选）
	Issuer      string                 `json:"iss,omitempty"`   // 签发者（可选）
	Subject     string                 `json:"sub,omitempty"`   // 主题（可选）
}

// GetUserID 暴露统一用户 ID 访问方法，便于中间件间解耦读取。
func (c *Claims) GetUserID() string {
	if c == nil {
		return ""
	}
	return c.UserID
}

// IsExpired 判断 Token 是否过期
func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.ExpiresAt
}

// IsValid 判断 Token 是否在有效期内
func (c *Claims) IsValid() bool {
	now := time.Now().Unix()
	// 检查是否过期
	if now > c.ExpiresAt {
		return false
	}
	// 检查是否还未生效
	if c.NotBefore > 0 && now < c.NotBefore {
		return false
	}
	return true
}

// HasRole 检查是否有指定角色
func (c *Claims) HasRole(roleCode string) bool {
	for _, role := range c.Roles {
		if role == roleCode {
			return true
		}
	}
	return false
}

// HasPermission 检查是否有指定权限
func (c *Claims) HasPermission(permCode string) bool {
	for _, perm := range c.Permissions {
		if perm == permCode {
			return true
		}
	}
	return false
}

// TokenPair Token 对（Access Token + Refresh Token）
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // 访问令牌（短期，15分钟）
	RefreshToken string `json:"refresh_token"` // 刷新令牌（长期，7天）
	ExpiresIn    int64  `json:"expires_in"`    // Access Token 过期时间（秒）
	TokenType    string `json:"token_type"`    // Token 类型（通常是 Bearer）
}

// NewTokenPair 创建 TokenPair
func NewTokenPair(accessToken, refreshToken string, expiresIn int64) *TokenPair {
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}
}

// User 用户信息
type User struct {
	ID          string    `json:"id"`                    // 用户ID
	Username    string    `json:"username"`              // 用户名
	Email       string    `json:"email"`                 // 邮箱
	Phone       string    `json:"phone,omitempty"`       // 手机号（可选）
	Roles       []string  `json:"roles"`                 // 角色列表（代码，如 admin, user）
	Permissions []string  `json:"permissions,omitempty"` // 权限列表（代码，如 user:read）
	Status      int       `json:"status"`                // 状态：1=正常 2=禁用 3=删除
	CreatedAt   time.Time `json:"created_at"`            // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`            // 更新时间
}

// IsActive 判断用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusNormal
}

// IsDisabled 判断用户是否被禁用
func (u *User) IsDisabled() bool {
	return u.Status == UserStatusDisabled
}

// IsDeleted 判断用户是否被删除
func (u *User) IsDeleted() bool {
	return u.Status == UserStatusDeleted
}

// HasRole 检查用户是否有指定角色
func (u *User) HasRole(roleCode string) bool {
	for _, role := range u.Roles {
		if role == roleCode {
			return true
		}
	}
	return false
}

// HasPermission 检查用户是否有指定权限
func (u *User) HasPermission(permCode string) bool {
	for _, perm := range u.Permissions {
		if perm == permCode {
			return true
		}
	}
	return false
}

// ToClaims 将 User 转换为 Claims（用于生成 Token）
func (u *User) ToClaims() *Claims {
	return &Claims{
		UserID:      u.ID,
		Username:    u.Username,
		Roles:       u.Roles,
		Permissions: u.Permissions,
	}
}

// Role 角色定义
type Role struct {
	ID          string    `json:"id"`          // 角色ID
	Name        string    `json:"name"`        // 角色名称（显示用，如 "管理员"）
	Code        string    `json:"code"`        // 角色代码（代码用，如 "admin"）
	Description string    `json:"description"` // 角色描述
	Permissions []string  `json:"permissions"` // 该角色拥有的权限列表
	Status      int       `json:"status"`      // 状态：1=启用 2=禁用
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// IsActive 判断角色是否启用
func (r *Role) IsActive() bool {
	return r.Status == 1
}

// HasPermission 检查角色是否有指定权限
func (r *Role) HasPermission(permCode string) bool {
	for _, perm := range r.Permissions {
		if perm == permCode {
			return true
		}
	}
	return false
}

// Permission 权限定义
type Permission struct {
	ID          string    `json:"id"`          // 权限ID
	Resource    string    `json:"resource"`    // 资源类型（如 user, post, comment）
	Action      string    `json:"action"`      // 操作类型（如 read, write, delete, *）
	Code        string    `json:"code"`        // 权限代码（如 user:read）
	Description string    `json:"description"` // 权限描述
	Status      int       `json:"status"`      // 状态：1=启用 2=禁用
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// IsActive 判断权限是否启用
func (p *Permission) IsActive() bool {
	return p.Status == 1
}

// Match 判断权限是否匹配（支持通配符）
func (p *Permission) Match(targetCode string) bool {
	// 精确匹配
	if p.Code == targetCode {
		return true
	}

	// 通配符匹配：user:* 匹配 user:read, user:write 等
	if p.Action == "*" {
		targetResource := ""
		for i, char := range targetCode {
			if char == ':' {
				targetResource = targetCode[:i]
				break
			}
		}
		if targetResource == p.Resource {
			return true
		}
	}

	// 超级权限：*:* 匹配所有
	if p.Code == "*:*" || p.Code == "*" {
		return true
	}

	return false
}

// NewPermission 创建权限（自动生成 Code）
func NewPermission(resource, action, description string) *Permission {
	return &Permission{
		Resource:    resource,
		Action:      action,
		Code:        resource + ":" + action,
		Description: description,
		Status:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Session 会话信息（Redis 分布式 Session）
type Session struct {
	SessionID string                 `json:"session_id"` // Session 唯一ID（64位随机字符串）
	UserID    string                 `json:"user_id"`    // 关联的用户ID
	Data      map[string]interface{} `json:"data"`       // Session 数据（灵活存储）
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	ExpiresAt time.Time              `json:"expires_at"` // 过期时间
	IP        string                 `json:"ip"`         // 客户端IP地址
	UserAgent string                 `json:"user_agent"` // 浏览器 User-Agent
}

// IsExpired 判断 Session 是否过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid 判断 Session 是否有效
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.SessionID != "" && s.UserID != ""
}

// Get 从 Session 数据中获取值
func (s *Session) Get(key string) (interface{}, bool) {
	if s.Data == nil {
		return nil, false
	}
	val, ok := s.Data[key]
	return val, ok
}

// Set 设置 Session 数据
func (s *Session) Set(key string, value interface{}) {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	s.Data[key] = value
}

// Delete 删除 Session 数据中的某个键
func (s *Session) Delete(key string) {
	if s.Data != nil {
		delete(s.Data, key)
	}
}

// OAuth2Config OAuth2 配置（支持 GitHub、Google 等）
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`     // 应用ID
	ClientSecret string   `json:"client_secret"` // 应用密钥
	RedirectURL  string   `json:"redirect_url"`  // 回调地址
	Scopes       []string `json:"scopes"`        // 权限范围
	AuthURL      string   `json:"auth_url"`      // 授权URL
	TokenURL     string   `json:"token_url"`     // Token交换URL
	UserInfoURL  string   `json:"user_info_url"` // 用户信息URL（可选）
}

// Validate 验证 OAuth2 配置是否完整
func (c *OAuth2Config) Validate() error {
	if c.ClientID == "" {
		return ErrOAuth2InvalidCode
	}
	if c.ClientSecret == "" {
		return ErrOAuth2InvalidCode
	}
	if c.RedirectURL == "" {
		return ErrOAuth2InvalidCode
	}
	if c.AuthURL == "" {
		return ErrOAuth2InvalidCode
	}
	if c.TokenURL == "" {
		return ErrOAuth2InvalidCode
	}
	return nil
}

// OAuth2Token OAuth2 令牌
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`  // 访问令牌
	RefreshToken string    `json:"refresh_token"` // 刷新令牌
	TokenType    string    `json:"token_type"`    // 令牌类型（通常是 Bearer）
	ExpiresIn    int64     `json:"expires_in"`    // 过期时间（秒）
	ExpiresAt    time.Time `json:"expires_at"`    // 具体过期时间
	Scope        string    `json:"scope"`         // 权限范围
}

// IsExpired 判断 OAuth2 Token 是否过期
func (t *OAuth2Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid 判断 OAuth2 Token 是否有效
func (t *OAuth2Token) IsValid() bool {
	return !t.IsExpired() && t.AccessToken != ""
}

// CacheEntry 缓存条目（用于本地缓存）
type CacheEntry struct {
	Key       string      `json:"key"`        // 缓存键
	Value     interface{} `json:"value"`      // 缓存值
	ExpiresAt time.Time   `json:"expires_at"` // 过期时间
}

// IsExpired 判断缓存是否过期
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// IsValid 判断缓存是否有效
func (e *CacheEntry) IsValid() bool {
	return !e.IsExpired() && e.Key != ""
}

// ========== 请求/响应类型（可选，用于应用层） ==========

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
	DeviceID string `json:"device_id,omitempty"`         // 设备ID（可选）
}

// LoginResponse 登录响应
type LoginResponse struct {
	TokenPair *TokenPair `json:"token_pair"` // Token 对
	User      *User      `json:"user"`       // 用户信息
	SessionID string     `json:"session_id"` // Session ID（如果使用 Session）
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // Refresh Token
}

// VerifyTokenRequest 验证 Token 请求
type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required"` // Access Token
}

// VerifyTokenResponse 验证 Token 响应
type VerifyTokenResponse struct {
	Valid  bool    `json:"valid"`  // 是否有效
	Claims *Claims `json:"claims"` // Token 声明（如果有效）
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID     string `json:"user_id" binding:"required"`    // 用户ID
	Permission string `json:"permission" binding:"required"` // 权限代码
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	HasPermission bool `json:"has_permission"` // 是否有权限
}

// UserProfile 用户资料（公开信息，不包含敏感字段）
type UserProfile struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FromUser 从 User 创建 UserProfile
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
	}
}

// PermissionGroup 权限分组（用于 UI 展示）
type PermissionGroup struct {
	Resource    string        `json:"resource"`    // 资源名称
	Permissions []*Permission `json:"permissions"` // 该资源下的权限列表
}

// RoleWithUsers 角色及其用户（用于管理界面）
type RoleWithUsers struct {
	Role      *Role    `json:"role"`       // 角色信息
	UserCount int      `json:"user_count"` // 拥有该角色的用户数量
	UserIDs   []string `json:"user_ids"`   // 用户ID列表
}
