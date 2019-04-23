// /////////////////////////////////////////////////////////////////////////////
// 常量-接口-types

package config

// /////////////////////////////////////////////////////////////////////////////
// 常量

// 运行环境
const (
	C_ENV_DEV = "development" // 默认：开发环境
	C_ENV_PRO = "production"  // 运营环境
)

// 配置文件路径
const (
	C_PATH_SCO_INI = "/config/sco.ini"        // world.ini 配置文件路径
	C_PATH_MASTER  = "/config/master.json"    // master 服务器配置文件路径
	C_PATH_SERVER  = "/config/servers.json"   // servers 服务器配置文件路径
	C_PATH_DIS     = "/config/discovery.json" // discovery.json 服务器配置文件路径
)

// 服务器类型
const (
	C_SERVER_TYPE_MASTER    = "master"    // master 服务器
	C_SERVER_TYPE_GATE      = "gate"      // gate 服务器类型
	C_SERVER_TYPE_CONNECTOR = "connector" // connector 服务器类型
)

// /////////////////////////////////////////////////////////////////////////////
// Sco.ini

// sco 引擎默认配置
type TScoIni struct {
	Env       string // 当前运行环境，production=开发环境；development=运营环境
	LogLevel  string // log 输出等级
	LogStderr bool   // 未知
	ShakeKey  string // 握手密钥
	Acceptor  uint32 // 1=tcp 2=websocket 3=tcp+websocket 4= tcp+websocket
}

// 数据库配置
type TD struct {
	Type string // 数据库类型
	Url  string // 数据库地址
}

// /////////////////////////////////////////////////////////////////////////////
// servers.json 配置文件

// servers.json 的 server 服务器信息
type TServerInfo struct {
	Name       string // 服务器的名字
	Host       string // 服务器的ID地址
	Port       uint   // 服务器端口
	ClientHost string // 面向客户端的 IP地址
	CTcpPort   uint   // 面向客户端的 tcp端口
	CWsPort    uint   // 面向客户端的 websocket端口
}

// 服务器 type -> *[]ServerInfo 信息集合
type TServerMap map[string][]*TServerInfo

// server.json 配置表
type TServerJson struct {
	Development TServerMap // 开发环境 配置信息
	Production  TServerMap // 运营环境 配置信息
}

// /////////////////////////////////////////////////////////////////////////////
// discovery.json 配置文件

type TDiscoveryJson struct {
	Development map[string][]string
	Production  map[string][]string
}
