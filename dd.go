package dd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var gForce = true                      // 子进程退出，无论什么情况都重启子进程
var gCheckUpdateInterval time.Duration // 检查文件更新频率
var gLastModTime time.Time             // 文件最后更新时间
var gQuit = false                      // 程序将退出，不在创建子进程
var gChild *os.Process                 // 子进程

// Daemon 启动守护进程
// daemon: 后台运行
// force: 无论child是否正常退出都重启，否则仅在程序异常退出时才重启
// interval: 检查程序更新频率，0=不检查
func Daemon(daemon, force bool, interval time.Duration) {
	gForce = force
	gCheckUpdateInterval = interval

	if isChild() { // 如果是子进程，直接返回
		return
	}

	if daemon && !isDaemon() { // 后台运行
		fork(true, true)
		os.Exit(0)
	}

	// 启动守护进程
	parent()
}

func parent() {
	// 创建进程
	fork(false, false)

	// 检查本程序是否更新，如果更新，重新启动
	if gCheckUpdateInterval != 0 {
		go update()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c

	// 关闭子进程
	gChild.Signal(syscall.SIGQUIT)
	os.Exit(0)
}

func wait() {
	state, err := gChild.Wait()
	if err != nil {
		log.Println(err)
	}

	if gForce && !gQuit { // 是否始终重启
		fork(false, false)
		return
	}

	if !state.Success() && !gQuit { // 异常退出
		fork(false, false)
		return
	}
}

func isDaemon() bool {
	if os.Getenv("d1/"+os.Args[0]) != "" {
		return true
	}
	return false
}

func isChild() bool {
	if os.Getenv("d2/"+os.Args[0]) != "" {
		return true
	}
	return false
}

func fork(daemon, isUpdate bool) {
	// 执行程序本身
	name := os.Args[0]
	// 执行路径
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	cmd := exec.Command(name)
	cmd.Args = os.Args
	cmd.Dir = pwd
	cmd.Env = os.Environ()
	if daemon {
		cmd.Env = append(cmd.Env, fmt.Sprintf("d1/%s=1", name))
	}
	if !isUpdate {
		cmd.Env = append(cmd.Env, fmt.Sprintf("d2/%s=1", name))
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalln(err)
	}

	gChild = cmd.Process
	// 检查子进程，退出后重启
	go wait()
}

func update() {
	// 执行程序本身
	name, _ := exec.LookPath(os.Args[0])
	for {
		time.Sleep(gCheckUpdateInterval)
		fi, err := os.Stat(name)
		if err != nil {
			log.Println(err)
			continue
		}

		// 第一次读取原程序时间
		if gLastModTime.IsZero() {
			gLastModTime = fi.ModTime()
		}

		// 原程序时间与当前程序时间不同
		if gLastModTime.Sub(fi.ModTime()) != 0 {
			// 结束子进程
			gQuit = true
			if runtime.GOOS == "windows" {
				gChild.Kill()
			} else {
				gChild.Signal(syscall.SIGTERM)
			}
			for { // 等待子进程完全退出
				time.Sleep(time.Second)
				if runtime.GOOS == "windows" {
					if _, err := os.FindProcess(gChild.Pid); err != nil {
						break
					}
				} else {
					// if syscall.Kill(gChild.Pid, 0) != nil {
					// 	break
					// }
					if gChild.Signal(syscall.Signal(0)) != nil {
						break
					}
				}

			}
			// 启动全新程序
			fork(true, true)
			os.Exit(0)
		}
	}
}
