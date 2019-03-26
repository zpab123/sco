// /////////////////////////////////////////////////////////////////////////////
// ini 配置文件读取对象

package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/zpab123/zaplog" // log库
)

// ////////////////////////////////////////////////////////////////////////////////
// 对外 api

// 加载 JSON 配置文件，并转化为结构体
//
//filepath=文件路径，v=需要写入的结构体指针 ）
func LoadJsonToSruct(filepath string, v interface{}) error {
	// 读取文件
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		zaplog.Fatalf("读取 Json 文件失败: %s", err.Error())
		return err
	}

	// 转化文件
	jsonErr := json.Unmarshal(bytes, v)
	if jsonErr != nil {
		zaplog.Fatalf("解析 Json 文件失败: %s", jsonErr.Error())
		return jsonErr
	}

	return nil
}

// 加载 JSON 配置文件，并转化为 map
//
// filepath=文件路径，v=需要写入的结构体指针
func LoadJsonToMap(filepath string, v interface{}) error {
	// 读取文件
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		zaplog.Fatalf("读取 Json 文件失败: %s", err.Error())
		return err
	}

	// 转化文件
	jsonErr := json.Unmarshal(bytes, v)
	if jsonErr != nil {
		zaplog.Fatalf("解析 Json 文件失败: %s", jsonErr.Error())
		return jsonErr
	}

	return nil
}
