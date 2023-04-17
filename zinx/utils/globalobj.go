package utils

import (
	"encoding/json"
	"fmt"
	"github.com/jian/Zinx/zinx/ziface"
	"os"
	"path/filepath"
)

type GlobalObj struct {
	TcpServer ziface.IServer // 当前Zinx	的全局Server对象
	Host      string         // 当前服务器主机Ip
	TcpPort   int            // 当前服务器主机监听端口号
	Name      string         // 当前服务器名称
	Version   string         // 当前Zinx版本号

	MaxPacketSize uint32 //需要数据包的最大值
	MaxConn       int    // 当前服务器主机允许的最大链接个数
}

// 定义一个全局的对象
var GlobalObject *GlobalObj

/*
   提供init方法， 默认加载
*/

func (g *GlobalObj) Reload() {
	file := "conf/zinx.json"
	path, _ := filepath.Abs(file)
	fmt.Printf("path:%s\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	//将json数据解析到struct 中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}

}

func init() {
	//初始化GlobalObject 变量， 设置一些默认值

	GlobalObject = &GlobalObj{
		Name:          "ZinxServerApp",
		Version:       "0.4",
		TcpPort:       7777,
		Host:          "0.0.0.0",
		MaxConn:       12000,
		MaxPacketSize: 4096,
	}

	GlobalObject.Reload()
}
