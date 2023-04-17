package main

import (
	"fmt"
	"github.com/jian/Zinx/zinx/ziface"
	"github.com/jian/Zinx/zinx/znet"
)

// ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Test PreHandle
func (this *PingRouter) PreHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle!")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("before ping ...\n"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle!")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("Ping ping ...\n"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}

// Test Handle
func (this *PingRouter) PostHandle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle!")
	_, err := request.GetConnection().GetTCPConnection().Write([]byte("Post ping ...\n"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}

// Server 模块的测试函数
func main() {
	s := znet.NewServer("zinx v0.1")

	s.AddRouter(&PingRouter{})
	s.Server()
}
