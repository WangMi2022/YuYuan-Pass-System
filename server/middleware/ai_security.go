package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/WangMi2022/mit-assets-admin/server/model/system"
	"github.com/WangMi2022/mit-assets-admin/server/utils"
	"github.com/gin-gonic/gin"
)

// AIRequestPolicy constrains one class of AI proxy request.
type AIRequestPolicy struct {
	Window      time.Duration
	MaxRequests int
	BodyLimit   int64
	Timeout     time.Duration
}

type aiRateEntry struct {
	windowStart time.Time
	requests    int
}

var aiRateLimiter = struct {
	sync.Mutex
	entries map[string]aiRateEntry
}{
	entries: make(map[string]aiRateEntry),
}

var aiSSELimiter = struct {
	sync.Mutex
	entries map[string]int
}{
	entries: make(map[string]int),
}

// AISecurity limits AI proxy resource use without recording prompt or response content.
func AISecurity(policy AIRequestPolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !validAIRequestPolicy(policy) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": 7, "msg": "AI 接口安全策略配置无效"})
			return
		}
		if c.Request.ContentLength > policy.BodyLimit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"code": 7, "msg": "AI 请求体过大"})
			return
		}
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, policy.BodyLimit)
		}
		if !allowAIRequest(aiClientKey(c), policy.Window, policy.MaxRequests, time.Now()) {
			c.Header("Retry-After", strconv.FormatInt(int64(policy.Window.Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 7, "msg": "AI 请求过于频繁，请稍后再试"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), policy.Timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// AISSEConcurrency limits long-lived SSE connections per authenticated user.
func AISSEConcurrency(maxConcurrent int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxConcurrent <= 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": 7, "msg": "AI 流式并发策略配置无效"})
			return
		}
		key := aiClientKey(c)
		if !acquireAISSE(key, maxConcurrent) {
			c.Header("Retry-After", "30")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 7, "msg": "AI 流式请求并发数已达上限"})
			return
		}
		defer releaseAISSE(key)
		c.Next()
	}
}

// AIOperationRecord persists AI invocation metadata while deliberately withholding model input and output.
func AIOperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		if global.GVA_DB == nil {
			return
		}
		record := system.SysOperationRecord{
			Ip:      c.ClientIP(),
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Agent:   c.Request.UserAgent(),
			UserID:  int(utils.GetUserID(c)),
			Status:  c.Writer.Status(),
			Latency: time.Since(startedAt),
			Body:    "[AI 请求内容不记录]",
			Resp:    "[AI 响应内容不记录]",
		}
		_ = global.GVA_DB.Create(&record).Error
	}
}

// RequireSuperAdmin blocks source-initialization operations for all non-super-administrator roles.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if utils.GetUserAuthorityId(c) != 888 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 7, "msg": "仅超级管理员可执行初始化文件变更"})
			return
		}
		c.Next()
	}
}

func validAIRequestPolicy(policy AIRequestPolicy) bool {
	return policy.Window > 0 && policy.MaxRequests > 0 && policy.BodyLimit > 0 && policy.Timeout > 0
}

func aiClientKey(c *gin.Context) string {
	userID := utils.GetUserID(c)
	if userID != 0 {
		return "user:" + strconv.FormatUint(uint64(userID), 10)
	}
	return "ip:" + c.ClientIP()
}

func allowAIRequest(key string, window time.Duration, maxRequests int, now time.Time) bool {
	aiRateLimiter.Lock()
	defer aiRateLimiter.Unlock()

	entry := aiRateLimiter.entries[key]
	if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= window {
		aiRateLimiter.entries[key] = aiRateEntry{windowStart: now, requests: 1}
		return true
	}
	if entry.requests >= maxRequests {
		return false
	}
	entry.requests++
	aiRateLimiter.entries[key] = entry
	return true
}

func acquireAISSE(key string, maxConcurrent int) bool {
	aiSSELimiter.Lock()
	defer aiSSELimiter.Unlock()
	if aiSSELimiter.entries[key] >= maxConcurrent {
		return false
	}
	aiSSELimiter.entries[key]++
	return true
}

func releaseAISSE(key string) {
	aiSSELimiter.Lock()
	defer aiSSELimiter.Unlock()
	if aiSSELimiter.entries[key] <= 1 {
		delete(aiSSELimiter.entries, key)
		return
	}
	aiSSELimiter.entries[key]--
}
