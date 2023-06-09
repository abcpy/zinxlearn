package znet

type Message struct {
	Id      uint32 //消息的Id
	DataLen uint32 // 消息的长度
	Data    []byte // 消息的内容
}

// 创建一个Message消息包
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		Id:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}

}

// 获取消息数据段长度
func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

// 获取消息ID
func (m *Message) GetMsgId() uint32 {
	return m.Id
}

// 获取消息内容
func (m *Message) GetData() []byte {
	return m.Data
}

// 设计消息ID
func (m *Message) SetMsgId(msgID uint32) {
	m.Id = msgID
}

// 设计消息内容
func (m *Message) SetData(data []byte) {
	m.Data = data
}

// 设置消息数据长度
func (m *Message) SetDataLen(len uint32) {
	m.DataLen = len
}
