package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrClientNotFound      = errors.New("client not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionNotFound     = errors.New("session not found")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrInvalidPermission   = errors.New("invalid permission")
)

// Permission 权限类型
type Permission string

const (
	// PermPluginManage 插件管理权限
	PermPluginManage Permission = "plugin:manage"
	// PermServiceManage 服务管理权限
	PermServiceManage Permission = "service:manage"
	// PermPluginUse 插件使用权限
	PermPluginUse Permission = "plugin:use"
)

// Client 客户端信息
type Client struct {
	ID          string       `json:"id"`
	Secret      string       `json:"secret"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// Session 会话信息
type Session struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthManager 认证管理器
type AuthManager struct {
	clients  map[string]*Client
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewAuthManager 创建认证管理器
func NewAuthManager() *AuthManager {
	return &AuthManager{
		clients:  make(map[string]*Client),
		sessions: make(map[string]*Session),
	}
}

// AddClient 添加客户端
func (am *AuthManager) AddClient(client *Client) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.clients[client.ID]; exists {
		return ErrClientAlreadyExists
	}

	am.clients[client.ID] = client
	return nil
}

// RemoveClient 移除客户端
func (am *AuthManager) RemoveClient(clientID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.clients[clientID]; !exists {
		return ErrClientNotFound
	}

	delete(am.clients, clientID)

	// 移除该客户端的所有会话
	for sessionID, session := range am.sessions {
		if session.ClientID == clientID {
			delete(am.sessions, sessionID)
		}
	}

	return nil
}

// GetClient 获取客户端信息
func (am *AuthManager) GetClient(clientID string) (*Client, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	client, exists := am.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// Authenticate 认证客户端
func (am *AuthManager) Authenticate(clientID, nonce string, timestamp int64, signature string) (string, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 获取客户端信息
	client, exists := am.clients[clientID]
	if !exists {
		return "", ErrClientNotFound
	}

	// 验证签名
	expectedSignature := generateSignature(client.Secret, clientID, nonce, timestamp)
	if signature != expectedSignature {
		return "", ErrInvalidCredentials
	}

	// 检查时间戳是否在合理范围内（5分钟内）
	now := time.Now()
	requestTime := time.Unix(timestamp, 0)
	if now.Sub(requestTime) > 5*time.Minute {
		return "", errors.New("timestamp expired")
	}

	// 创建会话
	sessionID := uuid.New().String()
	session := &Session{
		ID:        sessionID,
		ClientID:  clientID,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour), // 会话有效期24小时
	}

	am.sessions[sessionID] = session

	return sessionID, nil
}

// ValidateSession 验证会话
func (am *AuthManager) ValidateSession(sessionID string) (*Client, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	client, exists := am.clients[session.ClientID]
	if !exists {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// RevokeSession 撤销会话
func (am *AuthManager) RevokeSession(sessionID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.sessions[sessionID]; !exists {
		return ErrSessionNotFound
	}

	delete(am.sessions, sessionID)
	return nil
}

// HasPermission 检查客户端是否有指定权限
func (am *AuthManager) HasPermission(clientID string, perm Permission) (bool, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	client, exists := am.clients[clientID]
	if !exists {
		return false, ErrClientNotFound
	}

	for _, p := range client.Permissions {
		if p == perm {
			return true, nil
		}
	}

	return false, nil
}

// HasPluginPermission 检查客户端是否有使用特定插件的权限
func (am *AuthManager) HasPluginPermission(clientID, pluginID string) (bool, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	client, exists := am.clients[clientID]
	if !exists {
		return false, ErrClientNotFound
	}

	// 检查是否有全局插件使用权限
	for _, p := range client.Permissions {
		if p == PermPluginUse {
			return true, nil
		}
	}

	// 检查是否有特定插件使用权限
	pluginPerm := Permission(fmt.Sprintf("plugin:%s:use", pluginID))
	for _, p := range client.Permissions {
		if p == pluginPerm {
			return true, nil
		}
	}

	return false, nil
}

// generateSignature 生成签名
func generateSignature(secret, clientID, nonce string, timestamp int64) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%s:%s:%d", clientID, nonce, timestamp)))
	return hex.EncodeToString(h.Sum(nil))
}
