package main

import (
	"flag"
	"github.com/guotie/config"
	"github.com/guotie/deferinit"
	"runtime"
	"github.com/smtc/glog"
	"github.com/gin-gonic/gin"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"strings"
)

var (
	configFn                                    = flag.String("config", "./config.json", "config file path")
	debugFlag                                   = flag.Bool("d", false, "debug mode")
	rt                                   *gin.Engine
	rootPrefix                           string
)

func serverRun(cfn string, debug bool) {
	config.ReadCfg(cfn)
	logInit(debug)

	rootPrefix = strings.TrimSpace(config.GetStringDefault("rootprefix", ""))

	if len(rootPrefix) != 0 {
		if !strings.HasPrefix(rootPrefix, "/") {
			rootPrefix = "/" + rootPrefix
		}
		if strings.HasSuffix(rootPrefix, "/") {
			rootPrefix = rootPrefix[0 : len(rootPrefix)-1]
		}
	}

	// 初始化
	deferinit.InitAll()
	glog.Info("init all module successfully.\n")

	// 设置多cpu运行
	runtime.GOMAXPROCS(runtime.NumCPU())

	deferinit.RunRoutines()
	glog.Info("run routines successfully.\n")
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}
	rt = gin.Default()
	router(rt)
	go rt.Run(fmt.Sprintf(":%d", port))
}

// 结束进程
func serverExit() {
	// 结束所有go routine
	deferinit.StopRoutines()
	glog.Info("stop routine successfully.\n")

	deferinit.FiniAll()
	glog.Info("fini all modules successfully.\n")
}

func main() {
	//判断进程是否存在
	if checkPid() {
		return
	}

	flag.Parse()

	serverRun(*configFn, *debugFlag)

	c := make(chan os.Signal, 1)
	writePid()
	// 信号处理
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	// 等待信号
	<-c

	serverExit()
	rmPidFile()
	glog.Close()
	os.Exit(0)
}

func router(r *gin.Engine) {

	g := &r.RouterGroup
	if rootPrefix != "" {
		g = r.Group(rootPrefix)
	}
	{
		g.GET("/", func(c *gin.Context) {
			c.String(200, "ok")
		})
	}
}


// 报告用户请求相关信息
func userReqInfo(req *http.Request) (info string) {
	info += fmt.Sprintf("ipaddr: %s user-agent: %s referer: %s",
		req.RemoteAddr, req.UserAgent(), req.Referer())
	return info
}
