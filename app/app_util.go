// /////////////////////////////////////////////////////////////////////////////
// Application 一些辅助函数

package app

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/zpab123/sco/config"     // 配置管理
	"github.com/zpab123/sco/discovery"  // 服务发现
	"github.com/zpab123/sco/netservice" // 网络服务
	"github.com/zpab123/sco/network"    // 网络
	"github.com/zpab123/sco/protocol"   // 消息协议
	"github.com/zpab123/sco/rpc"        // rpc
	"github.com/zpab123/zaplog"         // log
)

// 完成 app 的默认设置
func defaultConfig(app *Application) {
	// 解析启动参数
	parseArgs(app)

	// 获取服务器信息
	getServerJson(app)

	// 设置 log 信息
	configLogger(app)

	// 设置 app 默认参数
	setdDfaultOpt(app)
}

// 解析 命令行参数
func parseArgs(app *Application) {
	// 参数定义
	//serverType := flag.String("type", "serverType", "server type") // 服务器类型，例如 gate connect area ...等
	//gid := flag.Uint("gid", 0, "gid")                                       // 服务器进程id
	name := flag.String("name", "gate_1", "server name") // 服务器名字
	//frontend := flag.Bool("frontend", false, "is frontend server")          // 是否是前端服务器
	//host := flag.String("host", "127.0.0.1", "server host")                 // 服务器IP地址
	//port := flag.Uint("port", 0, "server port")                             // 服务器端口
	//clientHost := flag.String("clientHost", "127.0.0.1", "for client host") // 面向客户端的 IP地址
	//cTcpPort := flag.Uint("cTcpPort", 0, "tcp port")                        // tcp 连接端口
	//cWsPort := flag.Uint("cWsPort", 0, "websocket port")                    // websocket 连接端口

	// 解析参数
	flag.Parse()

	// 赋值
	//cmdParam := &cmd.CmdParam{
	//ServerType: *serverType,
	//Gid:        *gid,
	//Name: *name,
	//Frontend:   *frontend,
	//Host:       *host,
	//Port:       *port,
	//ClientHost: *clientHost,
	//CTcpPort:   *cTcpPort,
	//CWsPort:    *cWsPort,
	//}

	// 设置 app 名字
	app.baseInfo.Name = *name
}

// 获取 server.json 信息
func getServerJson(app *Application) {
	// 根据 AppType 和 Name 获取 服务器配置参数
	appType := app.baseInfo.AppType
	name := app.baseInfo.Name
	list, ok := config.GetServerMap()[appType]

	if nil == list || len(list) <= 0 || !ok {
		zaplog.Fatalf("app 获取 appType 信息失败。 appType=%s", appType)

		os.Exit(1)
	}

	// 获取服务器信息
	for _, info := range list {
		if info.Name == name {
			app.serverInfo = info

			break
		}
	}

	if app.serverInfo == nil {
		zaplog.Fatal("app 获取 server.json 信息失败。 appName=%s", app.baseInfo.Name)

		os.Exit(1)
	}
}

// 设置 log 信息
func configLogger(app *Application) {
	// 模块名字
	zaplog.SetSource(app.baseInfo.Name)

	// 输出等级
	lv := config.GetScoIni().LogLevel
	zaplog.SetLevel(zaplog.ParseLevel(lv))

	// 输出文件
	logFile := fmt.Sprintf("./logs/%s.log", app.baseInfo.Name)
	var outputs []string
	stdErr := config.GetScoIni().LogStderr
	if stdErr {
		outputs = append(outputs, "stderr")
	}
	outputs = append(outputs, logFile)
	zaplog.SetOutput(outputs)
}

// 创建默认组件
func createComponent(app *Application) {
	// 网络服务
	nsOpt := app.Option.NetServiceOpt
	if nil != nsOpt && nsOpt.Enable {
		newNetService(app)
	}

	// 分布式组件
	if app.Option.Cluster {
		// rpcServer
		newRpcServer(app)

		// rpcClient
		newRpcClient(app)

		// 服务发现
		disOpt := app.Option.DiscoveryOpt
		if nil != disOpt && disOpt.Enable {
			newDiscovery(app)
		}
	}
}

// 创建 NetService 组件
func newNetService(app *Application) {
	// 创建地址
	serverInfo := app.serverInfo
	opt := app.Option.NetServiceOpt

	var tcpAddr string = ""
	if opt.ForClient && serverInfo.CTcpPort > 0 {
		tcpAddr = fmt.Sprintf("%s:%d", serverInfo.ClientHost, serverInfo.CTcpPort) // 面向客户端的 tcp 地址
	} else if serverInfo.Port > 0 {
		tcpAddr = fmt.Sprintf("%s:%d", serverInfo.Host, serverInfo.Port) // 面向服务器的 tcp 地址
	}

	var wsAddr string = ""
	if opt.ForClient && serverInfo.CWsPort > 0 {
		wsAddr = fmt.Sprintf("%s:%d", serverInfo.ClientHost, serverInfo.CWsPort) // 面向客户端的 websocket 地址
	} else if serverInfo.Port > 0 {
		wsAddr = fmt.Sprintf("%s:%d", serverInfo.Host, serverInfo.Port) // 面向服务器的 websocket 地址
	}

	laddr := &network.TLaddr{
		TcpAddr: tcpAddr,
		WsAddr:  wsAddr,
	}

	// 创建 NetServer
	ns, err := netservice.NewNetService(laddr, app, opt)
	if nil != err {
		zaplog.Fatal("app 创建 NetService 失败: %s", err.Error())

		os.Exit(1)
	}

	app.componentMgr.Add(ns)
}

// 创建 rpcClient
func newRpcClient(app *Application) {
	gc := rpc.NewGrpcClient()
	app.componentMgr.Add(gc)
	app.rpcClient = gc
}

// 创建服务发现组件
func newDiscovery(app *Application) {
	disMap := config.GetDiscoveryMap()
	endpoints := make([]string, 0)
	if _, ok := disMap["etcd"]; ok {
		endpoints = disMap["etcd"]
	} else {
		zaplog.Warnf("获取服务发现配置信息失败")

		return
	}

	// 服务描述
	svcDesc := &discovery.ServiceDesc{
		Type: app.baseInfo.AppType,
		Name: app.baseInfo.Name,
		Host: app.serverInfo.Host,
		Port: app.serverInfo.Port,
	}

	dc, _ := discovery.NewEtcdDiscovery(endpoints)
	dc.SetService(svcDesc)

	app.componentMgr.Add(dc)

	if nil != app.rpcClient {
		dc.AddListener(app.rpcClient)
	}
}

type Remote struct {
}

func (this *Remote) Call(ctx context.Context, rq *protocol.GrpcRequest) (*protocol.GrpcResponse, error) {
	return nil, nil
}

// 创建 rpcserver 组件
func newRpcServer(app *Application) {
	var laddr string = ""
	laddr = fmt.Sprintf("%s:%d", app.serverInfo.Host, app.serverInfo.Port)

	rs := rpc.NewGrpcServer(laddr)

	// 测试
	re := &Remote{}
	rs.SetRpcService(re)
	// 测试 end

	app.componentMgr.Add(rs)
}
