package adapter

import "Logger"

// NoopAdapter 空操作适配器，不输出任何日志
// 用于测试或需要禁用日志的场景
type NoopAdapter struct{}

// NewNoopAdapter 创建空操作适配器
func NewNoopAdapter() *NoopAdapter {
	return &NoopAdapter{}
}

// Log 实现 Logger.Adapter 接口（不执行任何操作）
func (a *NoopAdapter) Log(level Logger.Level, msg string, fields []Logger.Field) {
	// 不执行任何操作
}

// Sync 实现 Logger.Adapter 接口（不执行任何操作）
func (a *NoopAdapter) Sync() error {
	return nil
}
