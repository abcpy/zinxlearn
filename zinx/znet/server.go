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
	//Router ziface.IRouter
	msgHandle ziface.IMsgHandle

	// 当前server 的连接管理器
	connMgr ziface.IConnManager

	// 设置该Server的连接创建时Hook函数
	OnConnStart (func(conn ziface.IConnection))

	// 设置该Server的连接断开时Hook函数
	OnConnStop (func(conn ziface.IConnection))
}

// 实现 ziface.Iserver 里的全部接口方法

// 设置该Server的连接创建时的Hook函数
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 调用连接OnConnStat Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart...")
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStop...")
		s.OnConnStop(conn)
	}
}

func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s, listenner at IP: %s, Port: %d, is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("【Zinx】version: %s, Maxconn: %d, MaxPacketSize: %d\n", utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)

	// 开启一个go去做服务端Listener业务
	go func() {
		//0 启动worker 工作池机制
		s.msgHandle.StartWokerPool()

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

			//3.2 设置服务器最大连接控制， 如果超过最大连接， 那么九关闭新的连接
			if s.connMgr.Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//3.3 处理该连接的请求的业务方法
			dealConn := NewConnection(s, conn, cid, s.msgHandle)
			cid++

			//3.4 启动当前连接的处理业务
			go dealConn.Start()

		}
	}()

}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)

	//将其他需要清理的连接信息

	s.connMgr.ClearConn()
}

func (s *Server) Server() {
	s.Start()

	//阻塞， 否则主Go退出， listenner的go 将会退出
	for {
		time.Sleep(10 * time.Second)

	}
}

// 路由功能： 给当前服务注册一个路由业务方法， 供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandle.AddRouter(msgId, router)

	fmt.Println("Add Router succ!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.connMgr
}

func NewServer() ziface.IServer {

	// 先初始化全局配置文件
	utils.GlobalObject.Reload()

	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		msgHandle: NewMsgHandle(),
		connMgr:   NewConnManager(),
	}

	return s

}
