package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"time"
)

type SvrConfig struct {
	Listen  string
	Channel ClientInterface
}

//type Images struct {
//	Id int64
//	RegistryName string
//	Image string
//	Tag string
//	Timestamp int64
//	Digest string
//	Description string
//}

func getData(data gjson.Result, key string) string {
	if key == "host" {
		return gjson.Get(gjson.Get(data.String(), "request").String(), key).String()
	} else {
		return gjson.Get(gjson.Get(data.String(), "target").String(), key).String()
	}
}

func (config *SvrConfig) registry(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Printf("%T", body)
	events := gjson.Get(string(body), "events")
	fmt.Printf("%s", events)
	for _, event := range events.Array() {
		action := gjson.Get(event.String(), "action").String()
		if action == "push" {
			host := getData(event, "host")
			tag := getData(event, "tag")
			repo := getData(event, "repository")
			longFmtRepo := host + "/" + repo
			digest := getData(event, "digest")
			nowTime := time.Now().Format("2006-01-02 15:04:05")
			if len(tag) >= 1 {
				redisValue, _ := json.Marshal(Image{
					Host:        host,
					Repo:        repo,
					LongFmtRepo: longFmtRepo,
					Tag:         tag,
					Digest:      digest,
					Time:        nowTime,
				})
				pushErr := config.Channel.RPush("event", string(redisValue))
				if pushErr != nil {
					glog.V(4).Info("写入redis队列错误：", pushErr)
				}
				sAddErr := config.Channel.SAdd("eventBackup", string(redisValue))
				if sAddErr != nil {
					glog.V(4).Info("写入集合错误：", sAddErr)
				}
			}
		}
	}

}

func (config *SvrConfig) Server() {
	route := gin.Default()
	route.POST("/registry", config.registry)
	err := route.Run(config.Listen)
	if err != nil {
		panic(err)
	}
}

func (config *SvrConfig) Sentinel() {
	for {
		num, cErr := config.Channel.SCard("eventBackup")
		if cErr != nil {
			glog.V(4).Info("统计集合元素数错误：", cErr)
		}
		if num == 0 {
			//等于零说明没有没成功的元素
			glog.V(4).Info("集合元素数量为0,全部成功！")
		} else {
			//存在没成功的元素
			mem, mErr := config.Channel.SMembers("eventBackup")
			if mErr != nil {
				glog.V(4).Info("获取集合成员错误：", mErr)
			}
			for _, memB := range mem {
				strMem := string(memB)
				image := &Image{}
				errJ := json.Unmarshal([]byte(strMem), image)
				if errJ != nil {
					glog.V(4).Info("反序列化错误：", errJ)
				}
				loc, _ := time.LoadLocation("Local")
				tm, _ := time.ParseInLocation("2006-01-02 15:04:05", image.Time, loc)
				now := time.Now()
				sub := now.Sub(tm)
				if sub.Seconds() >= 10 {
					//重新添加到任务队列
					errR := config.Channel.SRem("eventBackup", strMem)
					if errR != nil {
						glog.V(4).Info("移除元素错误：", errR)
					}
					image.Time = time.Now().Format("2006-01-02 15:04:05") //重新更新时间戳
					redisValue, _ := json.Marshal(image)
					pushErr := config.Channel.RPush("event", string(redisValue))
					if pushErr != nil {
						glog.V(4).Info("写入redis队列错误：", pushErr)
					}
					sAddErr := config.Channel.SAdd("eventBackup", string(redisValue))
					if sAddErr != nil {
						glog.V(4).Info("写入集合错误：", sAddErr)
					}
				}

			}
		}
		time.Sleep(600 * time.Second)
	}
}
