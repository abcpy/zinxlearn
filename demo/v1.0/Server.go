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
	err := request.GetConnection().SendMsg(0, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}
}

type HelloZinxRouter struct {
	znet.BaseRouter
}

func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")

	// 先读取客户端的数据， 再回写 ping...ping...ping
	fmt.Println("recv from client: msgId=", request.GetMsgId(), ", data=", string(request.GetData()))

	// 回写数据
	err := request.GetConnection().SenBufdMsg(1, []byte("Hello Zinx Router v0.8"))
	if err != nil {
		fmt.Println("call back ping ping error")
	}

}

// 创建连接的时候执行
func DoConnectionBegin(conn ziface.IConnection) {
	fmt.Println("DoconnectionBegin is Called...")

	// 设置连接属性
	fmt.Println("Set conn Name, home done!")
	conn.SetProperty("Name", "Jian")
	conn.SetProperty("Home", "https//www.abdi")
	err := conn.SendMsg(2, []byte("DoConnectin BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}
}

// 连接断开的时候执行
func DoConnectionLost(conn ziface.IConnection) {
	fmt.Println("DoconnectionLost is Called...")

	if name, err := conn.GetProperty("Name"); err != nil {
		fmt.Println("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err != nil {
		fmt.Println("Conn Property Name = ", home)
	}

}

// Server 模块的测试函数
func main() {
	s := znet.NewServer()

	// 注册连接Hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})
	s.Server()
}
