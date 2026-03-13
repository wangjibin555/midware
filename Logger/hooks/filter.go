package hooks

import "strings"

// FilterHook 敏感信息过滤钩子
// 用于脱敏处理，如密码、token、密钥等
type FilterHook struct {
	sensitiveKeys map[string]bool // 需要过滤的字段名
	maskValue     string          // 替换值，默认 "***FILTERED***"
}

// NewFilterHook 创建过滤钩子
func NewFilterHook(keys ...string) *FilterHook {
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[strings.ToLower(key)] = true
	}

	return &FilterHook{
		sensitiveKeys: keyMap,
		maskValue:     "***FILTERED***",
	}
}

// WithMaskValue 设置自定义的遮蔽值
func (h *FilterHook) WithMaskValue(value string) *FilterHook {
	h.maskValue = value
	return h
}

// Levels 所有级别都过滤
func (h *FilterHook) Levels() []Level {
	return []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
	}
}

// Fire 执行过滤逻辑
func (h *FilterHook) Fire(entry *Entry) error {
	// 遍历字段，过滤敏感信息
	for i := range entry.Fields {
		key := strings.ToLower(entry.Fields[i].Key)
		if h.sensitiveKeys[key] {
			entry.Fields[i].Value = h.maskValue
		}
	}
	return nil
}
