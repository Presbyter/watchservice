package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
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

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Errorf("NewWatcher 失败. error: %s", err)
		return
	}
	defer watcher.Close()

	//
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				logrus.Debugf("EVENT! %s", event.String())
				// watch for errors
			case err := <-watcher.Errors:
				logrus.Errorf("watcher error: %s", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(folderPath); err != nil {
		logrus.Errorf("监听文件夹失败. error: %s", err)
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	logrus.Infof("收到信号 %v, 进程即将退出.", <-ch)
}
