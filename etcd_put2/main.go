package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

// main 测试put to etcd
func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.0.101:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	fmt.Println("init etcd success.")
	defer cli.Close()
	// put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, "/logagent/collect_config", `[{"path":"/Users/jiangshengping/wwwroot/spjiang/go/src/go-logagent/redis.log","topic":"redis"},{"path":"/Users/jiangshengping/wwwroot/spjiang/go/src/go-logagent/mysql.log","topic":"mysql"}]`)
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed,%v\n", err)
		return
	}
	return
}
