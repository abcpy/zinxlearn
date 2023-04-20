package znet

import (
	"fmt"
	"github.com/jian/Zinx/zinx/utils"
	"github.com/jian/Zinx/zinx/ziface"
	"net"
	"time"
)

// iserver 接口实现，定义一个Server服务类
type Server struct {
	// 服务器的名称
	Name string
	// tcp4 or other
	IPVersion string
	//IP地址
	IP string
	// 端口
	Port int

	// 当前Server由用户绑定的回调router, 也就是Server注册的链接对应的处理业务员
	Router ziface.IRouter
}

// 实现 ziface.Iserver 里的全部接口方法

func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s, listenner at IP: %s, Port: %d, is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("【Zinx】version: %s, Maxconn: %d, MaxPacketSize: %d\n", utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	// 开启一个go去做服务端Listener业务
	go func() {
		//1. 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err:", err)
			return
		}

		fmt.Printf("addr: %v\n", addr)

		//2. 监听服务器地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen:", s.IPVersion, "err:", err)
			return
		}
		//已经监听成功
		fmt.Println("start Zinx server ", s.Name, " succ, now listening...")

		var cid uint32
		cid = 0

		//3. 启动server网络连接业务
		for {
			// 3.1 阻塞等待客户端建立连接结束
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err:", err)
				continue
			}

			//3.3 处理该连接的请求的业务方法
			dealConn := NewConnection(conn, cid, s.Router)
			cid++

			//3.4 启动当前连接的处理业务
			go dealConn.Start()

		}
	}()

}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)
}

func (s *Server) Server() {
	s.Start()

	//阻塞， 否则主Go退出， listenner的go 将会退出
	for {
		time.Sleep(10 * time.Second)

	}
}

// 路由功能： 给当前服务注册一个路由业务方法， 供客户端链接处理使用
func (s *Server) AddRouter(router ziface.IRouter) {
	s.Router = router

	fmt.Println("Add Router succ!")
}

func NewServer() ziface.IServer {

	// 先初始化全局配置文件
	utils.GlobalObject.Reload()

	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		Router:    nil,
	}

	return s

}
