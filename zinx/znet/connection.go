package znet

import (
	"errors"
	"fmt"
	"github.com/jian/Zinx/zinx/utils"
	"github.com/jian/Zinx/zinx/ziface"
	"io"
	"net"
	"sync"
)

type Connection struct {
	TcpServer ziface.IServer // 当前conn属于哪个server, 在conn初始化的时候添加

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

	//无缓冲管道， 用于读。写两个goroutine 之间的消息通信
	msgChan chan []byte

	//有缓冲管道， 用于读。写两个goroutine 之间的消息通信
	msgBuffChan chan []byte

	// 连接属性
	property map[string]interface{}

	// 保护连接属性修改的锁
	propertyLock sync.RWMutex
}

// 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte), //msgChan初始化
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}), // 连接属性map初始化
	}

	//将新创建的Conn添加到连接管理中
	c.TcpServer.GetConnMgr().Add(c)

	return c

}

// 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("No Property found")
	}
}

// 移除连接
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
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
		if utils.GlobalObject.WorkerPoolSize > 0 {
			// 已经启动工作池机制， 将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(&request)
		} else {
			go c.MsgHandler.DoMsgHandler(&request)

		}
	}

}

/*
   写消息Goroutine， 用户将数据发送给给互动
*/

func (c *Connection) StartWriter() {
	fmt.Println("[writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn writer exit]")

	for {
		select {
		case data := <-c.msgChan:
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, "Conn Writer exit")
				return
			}
		// 缓冲channel
		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:, ", err, "Conn Writer exit")
					return
				}
			} else {
				break
				fmt.Println("msgBuffChan is closed")
			}
		case <-c.ExitBuffChan:
			// conn 已经关闭
			return
		}
	}

}

// 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//开启处理该连接读取到客户端数据之后的请求业务
	go c.StartReader()

	// 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()

	// 按照用户传递进来的连接时处理业务
	c.TcpServer.CallOnConnStart(c)

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

	// 如果用户注册了hook 函数
	c.TcpServer.CallOnConnStop(c)

	// 关闭socker 连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true

	// 将连接从连接管理中删除
	c.TcpServer.GetConnMgr().Remove(c)

	//关闭该连接全部管道
	close(c.ExitBuffChan)
	close(c.msgChan)

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
	//if _, err := c.Conn.Write(msg); err != nil {
	//	fmt.Println("Write msg id ", msgId, " error")
	//	c.ExitBuffChan <- true
	//	return errors.New("conn write error")
	//}
	c.msgChan <- msg // 将之前直接回写给conn. Write 的方法 改为 发送给Channel 供Writer读取
	return nil
}

func (c *Connection) SenBufdMsg(msgId uint32, data []byte) error {

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
	//if _, err := c.Conn.Write(msg); err != nil {
	//	fmt.Println("Write msg id ", msgId, " error")
	//	c.ExitBuffChan <- true
	//	return errors.New("conn write error")
	//}
	c.msgBuffChan <- msg // 将之前直接回写给conn. Write 的方法 改为 发送给Channel 供Writer读取
	return nil
}
