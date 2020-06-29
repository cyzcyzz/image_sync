package main

import (
	"flag"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	glog.Info("程序启动了")
	ParseConfig()
	redisCli := NewClient(RsConfig{Addr: Cfg.Redis.Addr, Db: Cfg.Redis.Db})
	if Cfg.Production.StartOver {
		glog.V(4).Info("启动模式为生产者")
		S := SvrConfig{Listen: Cfg.Production.Listen, Channel: redisCli}
		go S.Server()
		go S.Sentinel()
	}
	if Cfg.Consumer.StartOver {
		glog.V(4).Info("启动模式为消费者")
		client, err := NewDockerClient(Cfg.Consumer.DockerEndpoint)
		if err != nil {
			glog.V(4).Info("获取client失败", err)
		}
		glog.V(4).Info("获取到client", client.Client)
		go client.SyncImage(Cfg.Consumer.Target, redisCli)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalChan:
		glog.V(4).Info("程序退出了")
		os.Exit(0)
	}
}
