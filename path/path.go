// /////////////////////////////////////////////////////////////////////////////
// 路径

package path

import (
	"os"
	"path/filepath"

	"github.com/zpab123/zaplog" // log 库
)

// /////////////////////////////////////////////////////////////////////////////
// 路径

// 获取主程序所在绝对路径 例如：E:\code\go\go-project\src\test
func GetMainPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		zaplog.Error("获取主程序绝对路径失败")

		return "", err
	}

	//strings.Replace(dir, "\\", "/", -1)
	return dir, nil
}
