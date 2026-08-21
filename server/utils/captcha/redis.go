package captcha

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/WangMi2022/mit-assets-admin/server/global"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	DefaultExpiration    = 3 * time.Minute
	consumeCaptchaScript = "local value = redis.call('GET', KEYS[1]); if value then redis.call('DEL', KEYS[1]); end; return value"
)

func NewDefaultRedisStore() *RedisStore {
	return &RedisStore{
		Expiration: DefaultExpiration,
		PreKey:     "CAPTCHA_",
		Context:    context.TODO(),
	}
}

type RedisStore struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
}

func (rs *RedisStore) UseWithCtx(ctx context.Context) *RedisStore {
	if ctx != nil {
		rs.Context = ctx
	}
	return rs
}

func (rs *RedisStore) key(id string) string {
	return rs.PreKey + id
}

func (rs *RedisStore) Set(id string, value string) error {
	if global.GVA_REDIS == nil {
		return fmt.Errorf("redis captcha store is unavailable")
	}
	err := global.GVA_REDIS.Set(rs.Context, rs.key(id), value, rs.Expiration).Err()
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("RedisStoreSetError!", zap.Error(err))
		}
		return err
	}
	return nil
}

func (rs *RedisStore) Get(key string, clear bool) string {
	if global.GVA_REDIS == nil {
		return ""
	}
	redisKey := rs.key(key)
	var val string
	var err error
	if clear {
		var result interface{}
		result, err = global.GVA_REDIS.Eval(rs.Context, consumeCaptchaScript, []string{redisKey}).Result()
		if result != nil {
			val, _ = result.(string)
		}
	} else {
		val, err = global.GVA_REDIS.Get(rs.Context, redisKey).Result()
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("RedisStoreGetError!", zap.Error(err))
		}
		return ""
	}
	return val
}

func (rs *RedisStore) Verify(id, answer string, clear bool) bool {
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if id == "" || answer == "" {
		return false
	}
	v := rs.Get(id, clear)
	return strings.EqualFold(strings.TrimSpace(v), answer)
}
