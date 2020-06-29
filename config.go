package main

import (
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

type Config struct {
	Redis      RedisInfo      `json:"redisinfo"`
	Production ProductionInfo `json:"productioninfo"`
	Consumer   ConsumerInfo   `json:"consumerinfo"`
	Registry   RegistryInfo   `json:"registry"`
}

type ProductionInfo struct {
	StartOver bool   `json:"start_over"`
	Listen    string `json:"listen"`
}

type ConsumerInfo struct {
	StartOver      bool     `json:"start_over"`
	DockerEndpoint string   `json:"docker_endpoint"`
	Target         []string `json:"target"`
}

type RedisInfo struct {
	Addr string `json:"addr"`
	Db   int    `json:"db"`
}

type RegistryInfo struct {
	Url      []string `json:"url"`
	UserName string   `json:"user_name"`
	Password string   `json:"password"`
}

var Cfg Config

/*解析配置文件的包，配置文件解析为结构体
 */
func ParseConfig() {
	apps := kingpin.New("image_sync_program", "sync image from master to slave")
	config := apps.Flag("conf", "config file").Default("./config.toml").Short('c').String()
	//和glog包使用的flag包存在冲突，目前没有好的解决方法，只能是使用下面这段代码去除报错信息
	apps.Flag("l", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('l').String()
	apps.Flag("s", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('s').String()
	apps.Flag("v", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('v').String()
	apps.Flag("a", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('a').String()
	cmd, pErr := apps.Parse(os.Args[1:])
	if pErr != nil {
		glog.Errorf("解析命令错误，命令：%s,错误：%s", cmd, pErr)
	}
	_, err := toml.DecodeFile(*config, &Cfg)
	if err != nil {
		glog.Errorf("读取配置错误，请检查tom配置%s", err)
	}
}
