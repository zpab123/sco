// /////////////////////////////////////////////////////////////////////////////
// ini 配置文件读取对象

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"     // ini 库
	"github.com/zpab123/zaplog" // log 工具
)

// /////////////////////////////////////////////////////////////////////////////
// 私有 api

// 读取 sco.ini 配置文件
func readScoIni() {
	// 锁住线程
	configMutex.Lock()

	// retun 后，解锁
	defer configMutex.Unlock()

	// 读取配置文件
	fPath := filepath.Join(mainPath, C_PATH_SCO_INI)
	zaplog.Debugf("读取 sco.ini 配置文件，路径=%s", fPath)
	iniFile, err := ini.Load(fPath)

	// 错误检查
	checkConfigError(err, "")

	// 获取配置
	for _, sec := range iniFile.Sections() {
		// 配置名字
		secName := sec.Name()
		secName = strings.ToLower(secName) // 转化成小写

		// 读取配置
		if secName == "sco" {
			readSco(sec, scoIni)
		}
	}
}

// 检查 ini 读取错误
func checkConfigError(err error, msg string) {
	if err != nil {
		if msg == "" {
			msg = err.Error()
		}

		zaplog.Fatalf("读取 sco.ini 出现错误。 err=%s", msg)

		os.Exit(1)
	}
}

// 读取 sco 配置数据
func readSco(sec *ini.Section, conf *TScoIni) {
	// 设置默认
	conf.Env = C_ENV_DEV

	// 读取属性
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name()) //转化成小写
		if "env" == name {
			conf.Env = key.MustString(conf.Env)
			if conf.Env != C_ENV_DEV && conf.Env != C_ENV_PRO {
				conf.Env = C_ENV_DEV
				zaplog.Fatal("sco.ini 中 [sco]-env 参数配置错误。修改为 development")
			}
		} else if "log_level" == name {
			conf.LogLevel = key.MustString(conf.LogLevel)
		} else if "log_stderr" == name {
			conf.LogStderr = key.MustBool(conf.LogStderr)
		} else if "shake_key" == name {
			conf.ShakeKey = key.MustString(conf.ShakeKey)
		} else if "acceptor" == name {
			var a int = 0
			a = key.MustInt(a)
			conf.Acceptor = uint32(a)
		}
	}
}
