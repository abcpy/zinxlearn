package znet

import (
	"errors"
	"fmt"
	"github.com/jian/Zinx/zinx/ziface"
	"io"
	"net"
)

type Connection struct {
	//当前连接的socker TCP 套接字
	Conn *net.TCPConn

	// 当前连接的ID， 也可以称作为SessionID, ID 全局唯一
	ConnID uint32

	// 当前连接的关闭状态
	isClosed bool

	//该连接的处理方法api
	//handleAPI ziface.HandFunc

	// 该链接的处理方法router
	MsgHandler ziface.IMsgHandle

	//告知该连接已经退出/停止的channel
	ExitBuffChan chan bool
}

// 创建连接的方法
func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
	}

	return c

}

// 处理conn读数据的Goroutine
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		// 读取我们最大的数据buf中
		//buf := make([]byte, 512)
		//_, err := c.Conn.Read(buf)
		//if err != nil {
		//	fmt.Println("recv buf err ", err)
		//	c.ExitBuffChan <- true
		//	continue
		//}

		// 调用当前连接业务(这里执行的是当前conn的绑定的handle方法)
		//if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
		//	fmt.Println("connID ,", c.ConnID, " handle is error")
		//	c.ExitBuffChan <- true
		//	return
		//}

		//创建拆包对象
		dp := NewDataPack()

		//获取客户端的MSg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			c.ExitBuffChan <- true
			continue
		}

		//拆包， 得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			c.ExitBuffChan <- true
			continue
		}

		// 根据datalen 读取data, 放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				c.ExitBuffChan <- true
				continue
			}
		}

		msg.SetData(data)

		//得到当前客户端请求的request 数据
		request := Request{conn: c, msg: msg}

		//// 从路由router 中找到注册绑定Conn的对应的Handle
		//go func(request ziface.IRequest) {
		//	c.Router.PreHandle(request)
		//	c.Router.Handle(request)
		//	c.Router.PostHandle(request)
		//}(&request)

		// 绑定好的消息和对应的处理方法中执行对应的Handle方法
		go c.MsgHandler.DoMsgHandler(&request)
	}

}

// 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//开启处理该连接读取到客户端数据之后的请求业务
	go c.StartReader()

	for {
		select {
		case <-c.ExitBuffChan:
			//得到退出消息， 不再阻塞
			return
		}
	}
}

// 停止连接， 结束当前连接状态
func (c *Connection) Stop() {
	// 1. 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}

	c.isClosed = true

	// 关闭socker 连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true

	//关闭该连接全部管道
	close(c.ExitBuffChan)

}

// 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID

}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 直接将Message 数据发送给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {

	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	//将data封包， 并且发送
	dp := NewDataPack()

	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}

	// 写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("Write msg id ", msgId, " error")
		c.ExitBuffChan <- true
		return errors.New("conn write error")
	}
	return nil
}
