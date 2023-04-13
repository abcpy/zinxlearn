package ziface

import "net"

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
}

// 定义一个统一处理连接业务的接口
type HandFunc func(*net.TCPConn, []byte, int) error
