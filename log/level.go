////////////////////////////////////////////////////////////////////////////////
// log 等级

package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

////////////////////////////////////////////////////////////////////////////////
// 初始化

var (
	// DebugLevel level
	DebugLevel Level = Level(zap.DebugLevel)
	// InfoLevel level
	InfoLevel Level = Level(zap.InfoLevel)
	// WarnLevel level
	WarnLevel Level = Level(zap.WarnLevel)
	// ErrorLevel level
	ErrorLevel Level = Level(zap.ErrorLevel)
	// PanicLevel level
	PanicLevel Level = Level(zap.PanicLevel)
	// FatalLevel level
	FatalLevel Level = Level(zap.FatalLevel)
)

////////////////////////////////////////////////////////////////////////////////
// public api

// 根据 s 解析正确 Level
func ParseLevel(s string) Level {
	if strings.ToLower(s) == "debug" {
		return DebugLevel
	} else if strings.ToLower(s) == "info" {
		return InfoLevel
	} else if strings.ToLower(s) == "warn" || strings.ToLower(s) == "warning" {
		return WarnLevel
	} else if strings.ToLower(s) == "error" {
		return ErrorLevel
	} else if strings.ToLower(s) == "panic" {
		return PanicLevel
	} else if strings.ToLower(s) == "fatal" {
		return FatalLevel
	}

	//Errorf("设置 LogLevel 出错。 错误的 LogLevel=%s", s)

	return DebugLevel
}

////////////////////////////////////////////////////////////////////////////////
// Level

// log 等级
type Level = zapcore.Level
