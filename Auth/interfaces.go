package Auth

import "time"

// ========== 存储接口定义 ==========
// Auth 模块只定义接口，具体实现由应用层提供

// UserStore 用户存储接口
type UserStore interface {
	// GetByID 根据用户ID获取用户信息
	GetByID(userID string) (*User, error)

	// GetByUsername 根据用户名获取用户信息
	GetByUsername(username string) (*User, error)

	// GetByEmail 根据邮箱获取用户信息（可选）
	GetByEmail(email string) (*User, error)

	// ValidateCredentials 验证用户凭证（用户名+密码）
	// 返回：用户信息、错误信息
	// 注意：应该使用 crypto.VerifyPassword 验证密码
	ValidateCredentials(username, password string) (*User, error)

	// GetUserRoles 获取用户的角色列表（代码）
	GetUserRoles(userID string) ([]string, error)

	// GetUserPermissions 获取用户的权限列表（代码）
	GetUserPermissions(userID string) ([]string, error)
}

// TokenStore Token 黑名单存储接口
type TokenStore interface {
	// AddToBlacklist 将 Token 加入黑名单
	AddToBlacklist(token string, expireAt time.Time) error

	// IsInBlacklist 检查 Token 是否在黑名单中
	IsInBlacklist(token string) (bool, error)

	// RemoveFromBlacklist 从黑名单中移除（可选，通常依赖 Redis TTL 自动过期）
	RemoveFromBlacklist(token string) error
}

// PermissionStore 权限存储接口（可选）
type PermissionStore interface {
	// GetRolePermissions 获取角色的权限列表
	GetRolePermissions(roleCode string) ([]string, error)

	// GetPermission 获取权限的详细信息
	GetPermission(permCode string) (*Permission, error)

	// CheckPermission 检查权限是否存在且有效
	CheckPermission(permCode string) (bool, error)
}

// RefreshTokenStore Refresh Token 存储接口（可选）
type RefreshTokenStore interface {
	// Store 存储 Refresh Token
	Store(token, userID string, expireAt time.Time) error

	// Validate 验证 Refresh Token 是否有效
	// 返回：用户ID、错误信息
	Validate(token string) (userID string, err error)

	// Revoke 撤销 Refresh Token
	Revoke(token string) error

	// RevokeAllByUserID 撤销用户的所有 Refresh Token（强制登出）
	RevokeAllByUserID(userID string) error
}

// ========== 存储接口的空实现（用于测试） ==========

// NoopUserStore 空实现（不做任何操作）
type NoopUserStore struct{}

func (s *NoopUserStore) GetByID(userID string) (*User, error) {
	return nil, ErrNotImplemented
}

func (s *NoopUserStore) GetByUsername(username string) (*User, error) {
	return nil, ErrNotImplemented
}

func (s *NoopUserStore) GetByEmail(email string) (*User, error) {
	return nil, ErrNotImplemented
}

func (s *NoopUserStore) ValidateCredentials(username, password string) (*User, error) {
	return nil, ErrNotImplemented
}

func (s *NoopUserStore) GetUserRoles(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

func (s *NoopUserStore) GetUserPermissions(userID string) ([]string, error) {
	return nil, ErrNotImplemented
}

// NoopTokenStore 空实现（不做任何操作）
type NoopTokenStore struct{}

func (s *NoopTokenStore) AddToBlacklist(token string, expireAt time.Time) error {
	return ErrNotImplemented
}

func (s *NoopTokenStore) IsInBlacklist(token string) (bool, error) {
	return false, ErrNotImplemented
}

func (s *NoopTokenStore) RemoveFromBlacklist(token string) error {
	return ErrNotImplemented
}

// ========== 使用示例 ==========

/*
应用层实现存储接口：

// ===== MySQL 实现 =====
type MySQLUserStore struct {
    db *gorm.DB
}

func (s *MySQLUserStore) GetByUsername(username string) (*Auth.User, error) {
    var user Auth.User
    err := s.db.Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, Auth.ErrUserNotFound
    }
    return &user, err
}

func (s *MySQLUserStore) ValidateCredentials(username, password string) (*Auth.User, error) {
    // 查询用户
    var dbUser struct {
        Auth.User
        PasswordHash string `gorm:"column:password_hash"`
    }
    err := s.db.Where("username = ?", username).First(&dbUser).Error
    if err != nil {
        return nil, Auth.ErrUserNotFound
    }

    // 验证密码
    valid, err := crypto.VerifyPassword(password, dbUser.PasswordHash)
    if err != nil || !valid {
        return nil, Auth.ErrInvalidCredentials
    }

    return &dbUser.User, nil
}

// ===== Redis 实现 =====
type RedisTokenStore struct {
    redis *redis.Client
}

func (s *RedisTokenStore) AddToBlacklist(token string, expireAt time.Time) error {
    ctx := context.Background()
    ttl := time.Until(expireAt)
    if ttl <= 0 {
        return nil // 已过期的 Token 不需要加入黑名单
    }
    return s.redis.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

func (s *RedisTokenStore) IsInBlacklist(token string) (bool, error) {
    ctx := context.Background()
    val, err := s.redis.Get(ctx, "blacklist:"+token).Result()
    if errors.Is(err, redis.Nil) {
        return false, nil // 不在黑名单中
    }
    if err != nil {
        return false, err // Redis 错误
    }
    return val == "1", nil
}

// ===== 初始化 Auth =====
auth, _ := Auth.New(
    Auth.DefaultConfig(),
    Auth.WithJWTSecret("my-secret-key"),
)

// 注入存储实现
auth.SetUserStore(&MySQLUserStore{db: db})
auth.SetTokenStore(&RedisTokenStore{redis: rdb})

// ===== 使用 =====
// 登录
tokens, err := auth.Login("admin", "password123")

// 验证
claims, err := auth.Verify(tokens.AccessToken)

// 登出
err = auth.Logout(tokens.AccessToken)
*/
