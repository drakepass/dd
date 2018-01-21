package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/ohko/dd"
)

func main() {
	// 启动Daemon
	dd.Daemon(true, true, time.Second*3)

	// 子进程
	// ...

	// 等待结束信号
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	<-c
}
