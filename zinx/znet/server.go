package znet

import (
	"errors"
	"fmt"
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
}

// 定义当前客户端连接handle api
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	// 回显业务
	fmt.Println("[Conn Handle] CallBackToClient ...")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err", err)
		return errors.New("CallBackToClient error")
	}

	return nil
}

// 实现 ziface.Iserver 里的全部接口方法

func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port: %d, is starting\n", s.IP, s.Port)

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
			dealConn := NewConnection(conn, cid, CallBackToClient)
			cid++

			//3.4 启动当前连接的处理业务
			go dealConn.Start()

			////最大512字节的回显服务
			//go func() {
			//	// 不断的循环从客户端获取数据
			//	for {
			//		buf := make([]byte, 512)
			//		cnt, err := conn.Read(buf)
			//		if err != nil {
			//			fmt.Println("recv buf err ", err)
			//			continue
			//		}
			//		//回显
			//		if _, err := conn.Write(buf[:cnt]); err != nil {
			//			fmt.Println("write back buf err ", err)
			//			continue
			//		}
			//
			//	}
			//}()

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

func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      name,
		IPVersion: "tcp4",
		IP:        "0.0.0.0",
		Port:      7777,
	}

	return s

}
