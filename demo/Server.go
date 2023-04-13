package main

import "github.com/jian/Zinx/zinx/znet"

// Server 模块的测试函数
func main() {
	s := znet.NewServer("zinx v0.1")
	s.Server()
}
