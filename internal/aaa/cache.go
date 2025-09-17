package aaa

import (
	"sync"
	"time"
)

// PermissionCache provides TTL-based caching for user permissions
type PermissionCache struct {
	cache map[string]*CachedUser
	mutex sync.RWMutex
	ttl   time.Duration
}

// CachedUser represents cached user data with permissions
type CachedUser struct {
	UserID      string
	Username    string
	Roles       []AAARole
	Permissions []string
	ExpiresAt   time.Time
}

// NewPermissionCache creates a new permission cache with TTL
func NewPermissionCache(ttlMinutes int) *PermissionCache {
	cache := &PermissionCache{
		cache: make(map[string]*CachedUser),
		ttl:   time.Duration(ttlMinutes) * time.Minute,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves cached user data if not expired
func (pc *PermissionCache) Get(userID string) (*CachedUser, bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	if cached, exists := pc.cache[userID]; exists {
		if time.Now().Before(cached.ExpiresAt) {
			return cached, true
		}
		// Expired, remove it
		delete(pc.cache, userID)
	}

	return nil, false
}

// Set stores user data in cache with expiration
func (pc *PermissionCache) Set(userID string, user *CachedUser) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	pc.cache[userID] = user
}

// Clear removes a specific user from cache
func (pc *PermissionCache) Clear(userID string) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	delete(pc.cache, userID)
}

// cleanup runs periodically to remove expired entries
func (pc *PermissionCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pc.mutex.Lock()
		now := time.Now()
		for userID, cached := range pc.cache {
			if now.After(cached.ExpiresAt) {
				delete(pc.cache, userID)
			}
		}
		pc.mutex.Unlock()
	}
}

// GetStats returns cache statistics
func (pc *PermissionCache) GetStats() map[string]interface{} {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	return map[string]interface{}{
		"total_users": len(pc.cache),
		"ttl_minutes": int(pc.ttl.Minutes()),
	}
}
