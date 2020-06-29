package main

import (
	"encoding/json"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"
	"github.com/heroku/docker-registry-client/registry"
	"time"
)

type Image struct {
	Host        string `json:"host"`
	Repo        string `json:"repo"`
	LongFmtRepo string `json:"long_fmt_repo"`
	Tag         string `json:"tag"`
	Digest      string `json:"digest"`
	Time        string `json:"time"`
}

//type Info RegistryInfo

type Docker struct {
	Client *docker.Client
}

type result = []map[string]bool

//func newInfo()
func (d *Docker) Pull(host, repo, tag string) error {
	err := d.Client.PullImage(docker.PullImageOptions{
		Repository: host,
		Tag:        tag,
		Registry:   repo,
	}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}
	return nil
}

func (d *Docker) ReTag(newName, oldName string, opts Image) error {
	err := d.Client.TagImage(newName, docker.TagImageOptions{
		Repo:    oldName,
		Tag:     opts.Tag,
		Force:   false,
		Context: nil,
	})
	return err
}

func (d *Docker) Push(targets []string, opts Image) error {
	for _, target := range targets {
		err := d.Client.PushImage(docker.PushImageOptions{
			Name: target,
			Tag:  opts.Tag,
		}, docker.AuthConfiguration{
			Username: "123",
			Password: "123",
		}) //不添加用户名和密码会报错，随便添加了一个
		if err != nil {
			glog.V(4).Infof("上传镜像失败，目标：%s，标签：%s，错误：%s", target, opts.Tag, err)
			continue
		} else {
			glog.V(4).Infof("上传成功，目标：%s，标签：%s", target, opts.Tag)
		}
	}
	return nil
}

func (d *Docker) DeleteLocalImage(repo []string, tag string) (bool, error) {
	var ident = false
	for _, target := range repo {
		err := d.Client.RemoveImage(target + ":" + tag)
		if err != nil {
			glog.V(4).Infof("删除镜像失败，目标：%s，错误：%s", target, err)
			continue
		}
		ident = true
	}
	return ident, nil
}

func (d *Docker) PruneImage() (bool, error) {
	var ident bool
	_, err := d.Client.PruneImages(docker.PruneImagesOptions{
		Filters: nil,
		Context: nil,
	})
	if err != nil {
		glog.V(4).Infof("清除本地空镜像，错误：%s", err)
	}
	ident = true
	return ident, nil
}

func (r *RegistryInfo) getSuccess(repo, tag string) (result, error) {
	tmpResult := make(map[string]bool)
	var result result
	for _, url := range r.Url {
		hub, errH := registry.New(url, r.UserName, r.Password)
		if errH != nil {
			glog.Errorf("生成客户端失败，错误：%s", errH)
			continue
		}
		tagList, errT := hub.Tags(repo)
		if errT != nil {
			glog.V(4).Infof("获取tag列表失败，错误：%s", errT)
			continue
		}
		for _, tags := range tagList {
			if tags == tag {
				tmpResult[url] = true
				result = append(result, tmpResult)
				break
			}
		}
	}
	return result, nil
}

func (d *Docker) SyncImage(targets []string, chanel ClientInterface) {
	for {
		event, err := chanel.BLPop("event")
		if err != nil {
			glog.V(4).Info("弹出队列出错：", err)
		} else {
			glog.V(4).Info("弹出队列:", event)
			image := &Image{}
			errJ := json.Unmarshal([]byte(event), image)
			if errJ != nil {
				glog.V(4).Info("反序列化失败", errJ)
			}
			errPull := d.Pull(image.LongFmtRepo, image.Repo, image.Tag)
			if errPull != nil {
				glog.V(4).Info("拉取镜像失败", errPull)
				continue
			}
			var newRepoName []string
			oldName := image.Host + "/" + image.Repo + ":" + image.Tag
			for _, target := range targets {
				newName := target + "/" + image.Repo
				glog.V(4).Info("新名字：", newName)
				errReTag := d.ReTag(oldName, newName, *image)
				if errReTag != nil {
					glog.V(4).Info("重新打标失败", errReTag)
					continue
				}
				newRepoName = append(newRepoName, newName)
			}
			errPush := d.Push(newRepoName, *image)
			if errPush != nil {
				panic(errPush) //始终是空，未处理
			}
			r := &RegistryInfo{
				Password: Cfg.Registry.Password,
				UserName: Cfg.Registry.UserName,
				Url: Cfg.Registry.Url,
			}
			res, errG := r.getSuccess(image.Repo, image.Tag)
			if errG != nil {
				glog.V(4).Info("获取结果错误", errG)
			}
			for _, state := range res {
				for k, v := range state {
					if v {
						//成功后回调redis，删除备份集合里的数据，防止再次同步
						errRem := chanel.SRem("eventBackup", event)
						if errRem != nil {
							glog.V(4).Info("删除成员错误", errRem)
						}
					} else {
						glog.V(4).Info("目标库没查询到标签，镜像：", k)
					}
				}
			}
			oldName = image.Host + "/" + image.Repo
			newRepoName = append(newRepoName, oldName)
			resD, err := d.DeleteLocalImage(newRepoName, image.Tag)
			if err != nil && !resD {
				glog.V(4).Infof("删除镜像失败，错误%s", err)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func NewDockerClient(endpoint string) (Docker, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		glog.V(4).Info(endpoint)
		return Docker{Client: nil}, err
	}
	return Docker{Client: client}, err
}
