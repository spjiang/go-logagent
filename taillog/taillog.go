package taillog

import (
	"context"
	"fmt"
	"github.com/hpcloud/tail"
	"logagent/kafka"
	"time"
)

// TailTask
type TailTask struct {
	path     string
	topic    string
	instance *tail.Tail
	// 为了实现推出t.run()
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	tailObj.init() // 根据路径去打印对应的日志
	return
}

// init 初始化连接对象，并调用发往kafka函数
func (t *TailTask) init() {
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 是否跟随
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // 从文件的那个地方开始读
		MustExist: false,
		Poll:      true,
	}
	var err error
	t.instance, err = tail.TailFile(t.path, config)
	if err != nil {
		fmt.Println("tail file failed,err:", err)
		return
	}
	// 当goroutine执行函数退出的时候，goroutine就结束了
	go t.run() // 直接去采集日志发送到kafka
	return
}

// run 直接区采集发往kafka
func (t *TailTask) run() {
	// 1、读取日志
	for {
		select {
		case <-t.ctx.Done():
			fmt.Printf("tail task:%s_%s 结束了...", t.path, t.topic)
			return
		case line := <-t.instance.Lines: // 从tailObj的通道中一行一行的读取日志数据
			// 2、发送到kafka
			fmt.Printf("++++++++++log body: 【%s】++++++++++++\n", line.Text)
			// 发往kafka
			kafka.SendToChan(t.topic, line.Text) // 避免函数调用函数（同步），需要改成异步处理，存放在通道中，速度很快直接返回
		default:
			time.Sleep(time.Second)
		}
	}
}
