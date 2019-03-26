// /////////////////////////////////////////////////////////////////////////////
// 配置文件读取工具

package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/zpab123/sco/path" // 路径库
	"github.com/zpab123/zaplog"   // log 库
)

// /////////////////////////////////////////////////////////////////////////////
// 包 初始化

// 变量
var (
	configMutex sync.Mutex                // 进程互斥锁
	mainPath    string                    // 程序启动目录
	scoIni      *TScoIni     = &TScoIni{} // sco 引擎配置信息
	serverJSon  *TServerJson              // server.json 配置表
	serverMap   TServerMap                // servers.json 中// 服务器 type -> *[]ServerInfo 信息集合
)

// 初始化
func init() {
	// 获取mainPath
	dir, err := path.GetMainPath()
	if nil == err {
		mainPath = dir
	}

	// 读取 sco.ini 配置
	readScoIni()

	// 读取 servers.json 配置信息
	readServerJson()
}

// /////////////////////////////////////////////////////////////////////////////
// 对外 api

// 获取 sco.ini 配置对象
func GetScoIni() *TScoIni {
	return scoIni
}

// 获取 servers.json 配置信息
func GetServerJson() *TServerJson {
	return serverJSon
}

// 获取 当前环境的 服务器信息集合
func GetServerMap() TServerMap {
	return serverMap
}

// /////////////////////////////////////////////////////////////////////////////
// 私有 api

// 读取 servers.json 配置信息
func readServerJson() {
	// 锁住线程
	configMutex.Lock()

	// retun 后，解锁
	defer configMutex.Unlock()

	// 读取文件
	if nil == serverJSon {
		// 获取 main 路径
		if "" == mainPath {
			zaplog.Fatal("读取 main 路径失败")

			os.Exit(1)
		}

		// 创建对象
		serverJSon = &TServerJson{
			Development: TServerMap{},
			Production:  TServerMap{},
		}

		// 加载文件
		fPath := filepath.Join(mainPath, C_PATH_SERVER)
		LoadJsonToMap(fPath, serverJSon)

		// 根据运行环境赋值
		if C_ENV_DEV == scoIni.Env {
			serverMap = serverJSon.Development
		} else {
			serverMap = serverJSon.Production
		}
	}
}
