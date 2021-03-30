////////////////////////////////////////////////////////////////////////////////
// 根据 "go.uber.org/zap" 封装的1个log 库

package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

////////////////////////////////////////////////////////////////////////////////
// 初始化

var (
	Logger       *zap.Logger        // 标准 logger
	Sugar        *zap.SugaredLogger // 语法糖
	cfg          zap.Config         // log 配置
	source       string             // source 信息
	currentLevel Level              // 当前信息等级
)

func init() {
	ec := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "lv",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	cfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		DisableCaller:    true, // 是否打印行号
		Encoding:         "console",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    ec,
	}

	currentLevel = DebugLevel
	reBuildByCfg()
}

////////////////////////////////////////////////////////////////////////////////
// public api

// SetSource sets the component name (dispatcher/gate/game) of gwlog module
func SetSource(s string) {
	source = s
	reBuildByCfg()
}

// 设置 log 输出等级
func SetLevel(lv Level) {
	currentLevel = lv
	cfg.Level.SetLevel(lv)
}

// 获取当前 log 等级
func GetLevel() Level {
	return currentLevel
}

// 添加 log 输出文件信息
func SetOutput(path []string) {
	cfg.OutputPaths = path
	reBuildByCfg()
}

// 设置输出格式
func SetEncoding(e string) {
	if cfg.Encoding == e {
		return
	}

	if e == "json" {
		cfg.Encoding = "json"
	} else {
		cfg.Encoding = "console"
	}

	reBuildByCfg()
}

// 设置是否为开发模式
func SetDevelopment(dev bool) {
	cfg.Development = dev
	reBuildByCfg()
}

// 格式化1个 int8 字段
func Int8(key string, v int8) zap.Field {
	return zap.Int8(key, v)
}

// 格式化1个 uint8 字段
func Uint8(key string, v uint8) zap.Field {
	return zap.Uint8(key, v)
}

// 格式化1个 int16 字段
func Int16(key string, v int16) zap.Field {
	return zap.Int16(key, v)
}

// 格式化1个 uint16 字段
func Uint16(key string, v uint16) zap.Field {
	return zap.Uint16(key, v)
}

// 格式化1个 int32 字段
func Int32(key string, v int32) zap.Field {
	return zap.Int32(key, v)
}

// 格式化1个 uint32 字段
func Uint32(key string, v uint32) zap.Field {
	return zap.Uint32(key, v)
}

// 格式化1个 int64 字段
func Int64(key string, v int64) zap.Field {
	return zap.Int64(key, v)
}

// 格式化1个 uint64 字段
func Uint64(key string, v uint64) zap.Field {
	return zap.Uint64(key, v)
}

// 格式化1个 float32 字段
func Float32(key string, v float32) zap.Field {
	return zap.Float32(key, v)
}

// 格式化1个 float64 字段
func Float64(key string, v float64) zap.Field {
	return zap.Float64(key, v)
}

// 格式化1个 string 字段
func String(key string, v string) zap.Field {
	return zap.String(key, v)
}

// 格式化1个 Duration 字段
func Duration(key string, v time.Duration) zap.Field {
	return zap.Duration(key, v)
}

////////////////////////////////////////////////////////////////////////////////
// private api

// 根据 onfig 重新编译 Logger
func reBuildByCfg() {
	if newLogger, err := cfg.Build(); nil == err {
		if Logger != nil {
			Logger.Sync()
		}
		Logger = newLogger
		// logger = logger.With(zap.Time("ts", time.Now()))
		if source != "" {
			Logger = Logger.With(zap.String("source", source))
		}
		setSugar(Logger.Sugar())
	} else {
		panic(err)
	}
}

// 设置 Sugar
func setSugar(s *zap.SugaredLogger) {
	if nil != s {
		Sugar = s
	}
}
