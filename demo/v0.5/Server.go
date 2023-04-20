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

// Test Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router PreHandle!")
	// 先读取客户端的数据， 再回写 ping...ping...ping
	fmt.Println("recv from client: msgId=", request.GetMsgId(), ", data=", string(request.GetData()))

	// 回写数据
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}

// Server 模块的测试函数
func main() {
	s := znet.NewServer()

	s.AddRouter(&PingRouter{})
	s.Server()
}
