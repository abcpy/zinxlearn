package ziface

type IRequest interface {
	GetConnection() IConnection //获取请求连接信息
	GetData() []byte            // 获取请求信息的数据
}
