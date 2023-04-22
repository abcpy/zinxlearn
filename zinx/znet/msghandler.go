package znet

import (
	"fmt"
	"github.com/jian/Zinx/zinx/utils"
	"github.com/jian/Zinx/zinx/ziface"
	"strconv"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter // 存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                    // 业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    // Worker 负责取任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		//一个Worker对应一个queue
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen),
	}
}

func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgId(), "is not FOUND!")
		return
	}

	// 执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//1. 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api, msgId = " + strconv.Itoa(int(msgId)))
	}
	//2. 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

// 启动一个Worker 工作流程
func (mh *MsgHandle) StartOneWorker(workerId int, taskQueue chan ziface.IRequest) {
	fmt.Println("worker ID = ", workerId, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		//有消息则取出队列的Request， 并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}

}

func (mh *MsgHandle) StartWokerPool() {
	//遍历需要启动worker的数量， 依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个worker被启动
		// 给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		// 启动当前worker, 阻塞的等待对应队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])

	}
}

// 将消息交给TaskQUEUE, 由worker 进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {

	// 根据ConnID 来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(), " request msgId=", request.GetMsgId(), "to workerId=", workerID)

	//将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}
