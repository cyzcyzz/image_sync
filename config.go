package main

import (
	"github.com/BurntSushi/toml"
)

var config *Config

type Config struct {
	Redis      Redis      `toml:"redis"`
	//Production Production `toml:"production"`
	//Consumer   Consumer   `toml:"consumer"`
	Registry   Registry   `toml:"registry"`
	Global Global  `toml:"global"`
}

/*
type Production struct {
	StartOver bool   `toml:"start_over"`
	Listen    string `toml:"listen"`
}

type Consumer struct {
	StartOver      bool     `toml:"start_over"`
	DockerEndpoint string   `toml:"docker_endpoint"`
	Target         []string `toml:"target"`
}
*/
type Redis struct {
	Addr string `toml:"addr"`
	Db   int    `toml:"db"`
}

type Registry struct {
	Url      []string `toml:"url"`
	UserName string   `toml:"user_name"`
	Password string   `toml:"password"`
}

type Global struct {
	Listen    string `toml:"listen"`
	DockerEndpoint string   `toml:"docker_endpoint"`
	Target         []string `toml:"target"`
	Mode  string `toml:"mode"`
}
//var Cfg Config

//解析配置文件的包，配置文件解析为结构体
//func ParseConfig() {
//	apps := kingpin.New("image_sync_program", "sync image from master to slave")
//	config := apps.Flag("conf", "config file").Default("./config.toml").Short('c').String()
//	//和glog包使用的flag包存在冲突，目前没有好的解决方法，只能是使用下面这段代码去除报错信息
//	//apps.Flag("l", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('l').String()
//	//apps.Flag("s", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('s').String()
//	//apps.Flag("v", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('v').String()
//	//apps.Flag("a", "config file").Default("/Users/oyfacc/go/src/docker_registry_sync_v2/config/config.toml").Short('a').String()
//	cmd, pErr := apps.Parse(os.Args[1:])
//	if pErr != nil {
//		glog.Errorf("解析命令错误，命令：%s,错误：%s", cmd, pErr)
//	}
//	_, err := toml.DecodeFile(*config, &Cfg)
//	if err != nil {
//		glog.Errorf("读取配置错误，请检查tom配置%s", err)
//	}
//}

func LoadConfig(path string) error {
	c, err := ParseToml(path)
	if err !=nil {
		return err
	}
	config = c
	return nil
}

func ParseToml(path string) (*Config, error) {
	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}