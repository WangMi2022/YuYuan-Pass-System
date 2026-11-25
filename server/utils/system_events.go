package utils

import (
	"errors"
	"sync"
)

var ErrSystemReloadInProgress = errors.New("系统配置正在重载，请稍后再试")

// SystemEvents 定义系统级事件处理
type SystemEvents struct {
	reloadHandlers []func() error
	mu             sync.RWMutex
	reloadMu       sync.Mutex
}

// 全局事件管理器
var GlobalSystemEvents = &SystemEvents{}

// RegisterReloadHandler 注册系统重载处理函数
func (e *SystemEvents) RegisterReloadHandler(handler func() error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.reloadHandlers = append(e.reloadHandlers, handler)
}

// TriggerReload 触发所有注册的重载处理函数
func (e *SystemEvents) TriggerReload() error {
	if !e.reloadMu.TryLock() {
		return ErrSystemReloadInProgress
	}
	defer e.reloadMu.Unlock()

	e.mu.RLock()
	handlers := append([]func() error(nil), e.reloadHandlers...)
	e.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(); err != nil {
			return err
		}
	}
	return nil
}
