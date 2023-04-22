package ziface

import (
	"net"
)

// 定义连接接口
type IConnection interface {
	// 启动连接，让当前连接开始工作
	Start()

	// 停止连接， 结束当前连接状态
	Stop()

	// 从当前连接获取原始的socket TCPConn
	GetTCPConnection() *net.TCPConn

	// 获取当前连接ID
	GetConnID() uint32

	// 获取远程客户端地址信息
	RemoteAddr() net.Addr

	//直接将Message数据发送给远端的TCP客户端 （无缓冲）
	SendMsg(msgId uint32, data []byte) error

	// 直接将Message数据发送给远端的TCP客户端 （有缓冲）
	SenBufdMsg(msgId uint32, data []byte) error

	// 设置链接属性
	SetProperty(key string, value interface{})

	// 获取链接属性
	GetProperty(key string) (interface{}, error)

	// 移除连接
	RemoveProperty(key string)
}

// 定义一个统一处理连接业务的接口
type HandFunc func(*net.TCPConn, []byte, int) error
