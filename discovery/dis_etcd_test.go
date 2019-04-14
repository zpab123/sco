// /////////////////////////////////////////////////////////////////////////////
// etcdDiscovery 测试

package discovery

import (
	"fmt"
	"testing"
)

var (
	endpoints = []string{
		"http://192.168.1.180:2379",
		"http://192.168.1.180:2479",
		"http://192.168.1.180:2579",
	}
)

// 测试运行
func TestRun(t *testing.T) {
	sd, err := NewEtcdDiscovery(endpoints)
	if nil != err {
		fmt.Println(err.Error())
		return
	}

	sd.Run()
}
