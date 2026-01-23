package logger

type Field struct {
	Key string
	Value any
}

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}