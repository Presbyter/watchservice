package main

import (
	"context"
	"flag"
	"github.com/Presbyter/watchserver"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	folderPath string
)

func init() {
	flag.StringVar(&folderPath, "f", "", "所需要监听的文件夹")
	flag.Parse()

	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
}

func main() {
	// 参数校验
	if folderPath == "" {
		path, err := os.Executable()
		if err != nil {
			logrus.Errorf("获取当前路径失败. error: %s", err)
			return
		}
		folderPath = filepath.Dir(path)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := watchserver.NewWatcher(folderPath)
	go s.Run(ctx)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	logrus.Infof("收到信号 %v, 进程即将退出.", <-ch)
	cancel()
}
