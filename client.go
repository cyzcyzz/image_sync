package main

import (
	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"time"
)

type ClientInterface interface {
	LPop(key string) (string, error)
	RPush(key string, value string) error
	BLPop(key string) (string, error)
	SAdd(key string, mem string) error
	SRem(key string, mem string) error
	SMembers(key string) ([][]byte, error)
	SCard(key string) (int, error)
}

type Client struct {
	Pool *redis.Pool
}

type RsConfig struct {
	Addr        string
	Db          int
	Passwd      string
	MaxIdle     int           // 最多存在的空闲连接数
	MaxActive   int           // 最多激活的连接数
	IdleTimeout time.Duration // 空闲多长时间后连接被关闭
}

func (c *Client) LPop(key string) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()
	value, err := redis.String(conn.Do("LPOP", key))
	if err != nil {
		return "", err
	} else {
		return value, err
	}
}

func (c *Client) RPush(key string, value string) error {
	conn := c.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("LPUSH", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BLPop(key string) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()
	value, err := redis.ByteSlices(conn.Do("BLPOP", key, 5))
	var reValue string
	for _, mem := range value {
		if string(mem) == "event" {
			continue
		}
		reValue = string(mem)
	}
	if err != nil {
		return "", err
	}
	return reValue, err
}

func (c *Client) SAdd(key string, mem string) error {
	conn := c.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("SADD", key, mem)
	if err != nil {
		return err
	}
	return err
}

func (c *Client) SRem(key string, mem string) error {
	conn := c.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("SREM", key, mem)
	if err != nil {
		return err
	}
	return err
}

func (c *Client) SMembers(key string) ([][]byte, error) {
	conn := c.Pool.Get()
	defer conn.Close()
	result, err := redis.ByteSlices(conn.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}
	return result, err
}

func (c *Client) SCard(key string) (int, error) {
	conn := c.Pool.Get()
	defer conn.Close()
	num, err := redis.Int(conn.Do("SCARD", key))
	if err != nil {
		return 0, err
	}
	return num, err
}

func NewClient(config RsConfig) *Client {
	if config.MaxActive == 0 {
		config.MaxActive = 5
	}
	if config.MaxIdle == 0 {
		config.MaxIdle = 10
	}
	return &Client{
		Pool: &redis.Pool{
			Dial: func() (redis.Conn, error) {
				var c, err = redis.Dial("tcp", config.Addr, redis.DialDatabase(config.Db))
				if err != nil {
					glog.Errorf("redis错误%s", err)
					return nil, err
				}
				return c, nil
			},
			MaxIdle:     config.MaxIdle,
			MaxActive:   config.MaxActive,
			IdleTimeout: 5 * time.Second,
		},
	}
}
