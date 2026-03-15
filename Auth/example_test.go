package Auth_test

import (
	"Auth"
	"Auth/crypto"
	"fmt"
	"log"
	"time"
)

// ========== 示例：简单的内存用户存储 ==========

type MemoryUserStore struct {
	users map[string]*Auth.User // username -> user
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		users: make(map[string]*Auth.User),
	}
}

func (s *MemoryUserStore) AddUser(user *Auth.User, password string) error {
	// 加密密码
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	// 存储用户（实际应用中应该存入数据库）
	s.users[user.Username] = user
	// 密码哈希应该单独存储，这里为了演示简化处理
	s.users[user.Username+"_hash"] = &Auth.User{ID: hash}
	return nil
}

func (s *MemoryUserStore) GetByID(userID string) (*Auth.User, error) {
	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, Auth.ErrUserNotFound
}

func (s *MemoryUserStore) GetByUsername(username string) (*Auth.User, error) {
	user, ok := s.users[username]
	if !ok {
		return nil, Auth.ErrUserNotFound
	}
	return user, nil
}

func (s *MemoryUserStore) GetByEmail(email string) (*Auth.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, Auth.ErrUserNotFound
}

func (s *MemoryUserStore) ValidateCredentials(username, password string) (*Auth.User, error) {
	user, ok := s.users[username]
	if !ok {
		return nil, Auth.ErrUserNotFound
	}

	// 获取密码哈希
	hashUser, ok := s.users[username+"_hash"]
	if !ok {
		return nil, Auth.ErrInvalidCredentials
	}

	// 验证密码
	valid, err := crypto.VerifyPassword(password, hashUser.ID)
	if err != nil || !valid {
		return nil, Auth.ErrInvalidCredentials
	}

	return user, nil
}

func (s *MemoryUserStore) GetUserRoles(userID string) ([]string, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return user.Roles, nil
}

func (s *MemoryUserStore) GetUserPermissions(userID string) ([]string, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return user.Permissions, nil
}

// ========== 示例：简单的内存 Token 黑名单 ==========

type MemoryTokenStore struct {
	blacklist map[string]time.Time // token -> expireAt
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		blacklist: make(map[string]time.Time),
	}
}

func (s *MemoryTokenStore) AddToBlacklist(token string, expireAt time.Time) error {
	s.blacklist[token] = expireAt
	return nil
}

func (s *MemoryTokenStore) IsInBlacklist(token string) (bool, error) {
	expireAt, ok := s.blacklist[token]
	if !ok {
		return false, nil
	}
	// 检查是否已过期（自动清理）
	if time.Now().After(expireAt) {
		delete(s.blacklist, token)
		return false, nil
	}
	return true, nil
}

func (s *MemoryTokenStore) RemoveFromBlacklist(token string) error {
	delete(s.blacklist, token)
	return nil
}

// ========== 完整使用示例 ==========

func Example() {
	// 1. 创建 Auth 实例
	auth, err := Auth.New(
		Auth.DefaultConfig(),
		Auth.WithJWTSecret("my-super-secret-key-32-bytes!!"),
		Auth.WithAccessTokenExpire(15*time.Minute),
		Auth.WithRefreshTokenExpire(7*24*time.Hour),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 注入存储实现
	userStore := NewMemoryUserStore()
	tokenStore := NewMemoryTokenStore()
	auth.SetUserStore(userStore)
	auth.SetTokenStore(tokenStore)

	// 3. 添加测试用户
	testUser := &Auth.User{
		ID:          "user-001",
		Username:    "admin",
		Email:       "admin@example.com",
		Roles:       []string{"admin"},
		Permissions: []string{"user:read", "user:write", "post:*"},
		Status:      Auth.UserStatusNormal,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := userStore.AddUser(testUser, "password123"); err != nil {
		log.Fatal(err)
	}

	// 4. 用户登录
	tokens, err := auth.Login("admin", "password123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("登录成功！")
	fmt.Printf("Access Token: %s...\n", tokens.AccessToken[:50])
	fmt.Printf("Refresh Token: %s...\n", tokens.RefreshToken[:50])
	fmt.Printf("Expires In: %d seconds\n", tokens.ExpiresIn)

	// 5. 验证 Token
	claims, err := auth.Verify(tokens.AccessToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nToken 验证成功！\n")
	fmt.Printf("User ID: %s\n", claims.UserID)
	fmt.Printf("Username: %s\n", claims.Username)
	fmt.Printf("Roles: %v\n", claims.Roles)
	fmt.Printf("Permissions: %v\n", claims.Permissions)

	// 6. 权限检查
	if auth.CheckPermission(claims, "user:read") {
		fmt.Println("\n✓ 有 user:read 权限")
	}
	if auth.CheckPermission(claims, "post:delete") {
		fmt.Println("✓ 有 post:delete 权限（通配符匹配）")
	}
	if !auth.CheckPermission(claims, "admin:delete") {
		fmt.Println("✗ 没有 admin:delete 权限")
	}

	// 7. 刷新 Token
	newTokens, err := auth.Refresh(tokens.RefreshToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nToken 刷新成功！\n")
	fmt.Printf("New Access Token: %s...\n", newTokens.AccessToken[:50])

	// 8. 登出（将 Token 加入黑名单）
	if err := auth.Logout(tokens.AccessToken); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\n登出成功！")

	// 9. 验证已登出的 Token（应该失败）
	_, err = auth.Verify(tokens.AccessToken)
	if err == Auth.ErrRevokedToken {
		fmt.Println("✓ Token 已被撤销，验证失败（预期行为）")
	}

	// Output:
	// 登录成功！
}
