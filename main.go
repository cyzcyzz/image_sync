package main

import "C"
import (
	"flag"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

var conf string

func init() {
	flag.StringVar(&conf,"conf", "./config.toml", "config file")
}


func main() {
	flag.Parse()
	defer glog.Flush()
	glog.Info("程序启动了")
	err := LoadConfig(conf)
	if err != nil {
		glog.Error("解析配置文件错误")
	}
	redisCli := NewClient(RsConfig{Addr: config.Redis.Addr, Db: config.Redis.Db})
	if config.Global.Mode == "production" {
		glog.V(4).Info("启动模式为生产者")
		S := SvrConfig{Listen: config.Global.Listen, Channel: redisCli}
		go S.Server()
		go S.Sentinel()
	}
	if config.Global.Mode == "consumer" {
		glog.V(4).Info("启动模式为消费者")
		client, err := NewDockerClient(config.Global.DockerEndpoint)
		if err != nil {
			glog.V(4).Info("获取client失败", err)
		}
		glog.V(4).Info("获取到client", client.Client)
		go client.SyncImage(config.Global.Target, redisCli)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalChan:
		glog.V(4).Info("程序退出了")
		os.Exit(0)
	}
}
