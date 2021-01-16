package taillog

import (
	"fmt"
	"logagent/etcd"
	"time"
)

var tskMgr *taillogMgr

// taillogMgr tailTask管理者
type taillogMgr struct {
	logEntry    []*etcd.LogEntry
	tskMap      map[string]*TailTask
	newConfChan chan []*etcd.LogEntry
}

func Init(logEntryConf []*etcd.LogEntry) {
	tskMgr = &taillogMgr{
		logEntry:    logEntryConf,
		tskMap:      make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry), // 无缓冲区的通道
	}
	// 根据配置项目启动监听任务处理日志
	for _, logEntry := range tskMgr.logEntry {
		tailObj := NewTailTask(logEntry.Path, logEntry.Topic)
		key := fmt.Sprintf("%s_%s", logEntry.Path, logEntry.Topic)
		tskMgr.tskMap[key] = tailObj
	}
	go tskMgr.run()
}

// run 监听自己的newConfChan，有了新的配置来之后就做对应的处理
func (t *taillogMgr) run() {
	for {
		select {
		case newConf := <-t.newConfChan:
			for _, conf := range newConf {
				key := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
				if _, ok := t.tskMap[key]; ok {
					continue
				} else {
					// 新增
					tailObj := NewTailTask(conf.Path, conf.Topic)
					t.tskMap[key] = tailObj
					continue
				}
			}
			// 找出原来的logEntry有，但是NewConf中没有的，进行删除
			for _, c1 := range t.logEntry {
				isDelete := true
				for _, c2 := range newConf {
					if c1.Path == c2.Path && c1.Topic == c2.Topic {
						isDelete = false
						continue
					}
				}
				if isDelete {
					// 把c1对应的这个tailObj给停掉
					key := fmt.Sprintf("%s_%s", c1.Path, c1.Topic)
					// 退出对应协程
					t.tskMap[key].cancelFunc()
				}
			}
			fmt.Printf("新的配置来了%v\n", newConf)
		default:
			time.Sleep(time.Second)
		}

	}
}

// 向外暴露一个函数，用于向taillogMgr结构体中的newConfChan通道newConfChan写入数据
func NewConfChan() chan<- []*etcd.LogEntry {
	return tskMgr.newConfChan
}
