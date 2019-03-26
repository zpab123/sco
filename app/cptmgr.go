// /////////////////////////////////////////////////////////////////////////////
// app 组件管理

package app

import (
	"github.com/zpab123/sco/model"   // 全局模型
	"github.com/zpab123/sco/network" // 网络
	"github.com/zpab123/zaplog"      // log
)

// /////////////////////////////////////////////////////////////////////////////
// 包 初始化

// /////////////////////////////////////////////////////////////////////////////
// ComponentManager 对象

// app 组件管理
type ComponentManager struct {
	componentMap map[string]model.IComponent // 名字-> 组件 集合
	netServerOpt *network.TNetServerOpt      // server 组件配置参数
}

// 新建1个 ComponentManager
func NewComponentManager() *ComponentManager {
	// 组件
	cptMgr := &ComponentManager{
		componentMap: map[string]model.IComponent{},
	}

	// 返回
	return cptMgr
}

// 添加1个 Component 组件
func (this *ComponentManager) AddComponent(cmpt model.IComponent) {
	// 获取名字
	name := cmpt.Name()

	// 组件已经存在
	if _, ok := this.componentMap[name]; ok {
		zaplog.Warnf("组件[*s]重复添加，新组件将覆盖旧组件", name)
	}

	// 保存组件
	this.componentMap[name] = cmpt
}

// 根据名字获取组件
func (this *ComponentManager) GetComponentByName(name string) model.IComponent {
	cpt, ok := this.componentMap[name]

	if ok {
		return cpt
	} else {
		return nil
	}
}

// 获取 netServer 组件参数
func (this *ComponentManager) GetNetServerOpt() *network.TNetServerOpt {
	return this.netServerOpt
}

// 设置 netServer 组件参数
func (this *ComponentManager) SetNetServerOpt(opt *network.TNetServerOpt) {
	this.netServerOpt = opt
}
