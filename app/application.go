// /////////////////////////////////////////////////////////////////////////////
// 1个通用服务器对象

package app

import (
	"sync"
)

// /////////////////////////////////////////////////////////////////////////////
// Application 对象

// 1个通用服务器对象
type Application struct {
	stopGroup    sync.WaitGroup    // stop 等待组
	componentMgr *ComponentManager // 组件管理
}

// 创建1个新的 Application 对象
func NewApplication() *Application {

}

// 初始化 Application
func (this *Application) Init() {

}

// 启动 app
func (this *Application) Run() {
	// 启动所有组件
	this.runComponent()

	// 等待停止
	this.stopGroup.Wait()

	// 停止完成
}

// 停止 app
func (this *Application) Stop() {
	// 停止所有组件
	this.stopComponent()
}

// 启动所有组件
func (this *Application) runComponent() {
	for _, cmpt := range this.componentMgr.componentMap {
		this.stopGroup.Add(1)

		go func() {
			defer this.stopGroup.Done()

			cmpt.Run()
		}()
	}
}

// 停止所有组件
func (this *Application) stopComponent() {
	for _, cmpt := range this.componentMgr.componentMap {
		cmpt.Stop()
	}
}
