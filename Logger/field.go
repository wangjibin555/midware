package Logger

import "time"

const (
	StringType FieldType = iota
	IntType
	Int64Type
	FloatType
	BoolType
	TimeType
	DurationType
	ErrorType
	AnyType
)

type FieldType int8
type Field struct {
	Key   string
	Type  FieldType
	Value interface{}
}

type Fields map[string]*Field

// 便捷构建，让创建日志字段类型安全，字段简单，性能更好
func String(key, val string) Field {
	return Field{
		Key:   key,
		Type:  StringType,
		Value: val,
	}
}

func Int(key string, val int) Field {
	return Field{
		Key:   key,
		Type:  IntType,
		Value: val,
	}
}

func Int64(key string, val int64) Field {
	return Field{
		Key:   key,
		Type:  Int64Type,
		Value: val,
	}
}

func Float(key string, val float64) Field {
	return Field{
		Key:   key,
		Type:  FloatType,
		Value: val,
	}
}

func Bool(key string, val bool) Field {
	return Field{
		Key:   key,
		Type:  BoolType,
		Value: val,
	}
}

func Time(key string, val time.Time) Field {
	return Field{
		Key:   key,
		Type:  TimeType,
		Value: val,
	}
}

func Duration(key string, val time.Duration) Field {
	return Field{
		Key:   key,
		Type:  DurationType,
		Value: val,
	}
}

// Err 创建错误类型字段
// 注意：使用 Err 而非 Error 以避免与全局日志方法 Error() 冲突
func Err(key string, val error) Field {
	return Field{
		Key:   key,
		Type:  ErrorType,
		Value: val,
	}
}

func Any(key string, val interface{}) Field {
	return Field{
		Key:   key,
		Type:  AnyType,
		Value: val,
	}
}
