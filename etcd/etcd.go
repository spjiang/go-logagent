package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type LogEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

var (
	cli *clientv3.Client
)

// Init 初始化etcd
func Init(addr string, timeout time.Duration) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{addr},
		DialTimeout: timeout,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	return
}

// GetConf 获取日志信息
func GetConf(key string) (logEntryConf []*LogEntry, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, key)
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err %v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("resp.Kvs, body %v\n", string(ev.Value))
		err = json.Unmarshal(ev.Value, &logEntryConf)
		if err != nil {
			fmt.Printf("Unmarshal etcd  value  failed, err %v\n", err)
			return
		}
	}
	return
}

// WatchConf etcd watch 观察etcd数据变化
func WatchConf(key string, newConfChan chan<- []*LogEntry) {
	ch := cli.Watch(context.Background(), key)
	for wresp := range ch {
		for _, evt := range wresp.Events {
			fmt.Printf("Type:%v key:%v value:%v", evt.Type, string(evt.Kv.Key), string(evt.Kv.Key))
			// 通知别人
			var newConf []*LogEntry

			if evt.Type != clientv3.EventTypeDelete {
				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					fmt.Printf("Unmarshal failed, err:%v\n", err)
					continue
				}
			}
			fmt.Printf("get new conf :%v\n", newConf)
			newConfChan <- newConf
		}
	}
}
