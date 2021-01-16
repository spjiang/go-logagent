package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"logagent/conf"
	"logagent/etcd"
	"logagent/kafka"
	"logagent/taillog"
	"sync"
	"time"
)

var cfg = new(conf.AppConf)

func main() {
	// 0、加载配置文件
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Printf("Init config failed, err:%v\n", err)
		return
	}

	// 1、初始化 kafka连接
	err = kafka.Init([]string{cfg.KafkaConfig.Address}, cfg.KafkaConfig.ChanMaxSize)
	if err != nil {
		fmt.Printf("init Kafka failed,err:%v\n", err)
		return
	}
	fmt.Println("init kafka success.")

	// 2、初始化etcd
	err = etcd.Init(cfg.EtcdConfig.Address, time.Duration(cfg.EtcdConfig.Timeout)*time.Second)
	if err != nil {
		fmt.Printf("init etcd failed,err:%v\n", err)
		return
	}
	fmt.Println("init etcd success.")

	// 2.1 从etcd中获取日志收集项的配置信息
	logEntryConf, err := etcd.GetConf(cfg.EtcdConfig.Key)
	if err != nil {
		fmt.Printf("etcd.GetConf failed,err:%v\n", err)
		return
	}

	// debug 打印配置项目
	for index, value := range logEntryConf {
		fmt.Printf("index:%v value:%v\n", index, value)
	}

	taillog.Init(logEntryConf)

	// 2.2 排一个哨兵监视日志收集项目的变化（有变化及时通知我的logAgent实现热加载配置）
	// 因为NewConfChan访问了tskMgr的newConfChan中，这个channel是在taillog.Init(logEntryConf)执行的初始化
	newConfChan := taillog.NewConfChan() // 从taillog包中获取对外暴露的通道

	var wg sync.WaitGroup
	wg.Add(1)
	etcd.WatchConf(cfg.EtcdConfig.Key, newConfChan) // 哨兵发现最新的的配置信息会通知上面的通道
	wg.Wait()

}
