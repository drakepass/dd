# dd
Daemon服务

# 功能
- 使用非常方便
- 设置后台运行
- 检查子进程
- 程序更新检查

# 使用
``` golang
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
```