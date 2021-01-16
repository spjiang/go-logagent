package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

// kafka日志模块

type logData struct {
	topic string
	data  string
}

var (
	client      sarama.SyncProducer // 声明一个全局的一个kafka生产者客户端
	logDataChan chan *logData
)

// Init 初始化client
func Init(addrs []string, maxSize int) (err error) {
	config := sarama.NewConfig()
	// tailf包使用
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader 和 follow 都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	client, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		fmt.Println("producer closed,err:", err)
		return
	}

	// 初始化数据
	logDataChan = make(chan *logData, maxSize)

	// 开启后台goroutine从通道中获取数据发往kafka
	go sendToKafka()
	return
}

// SendToChan taillog发送数据到此通道中,给外部暴露一个函数，该函数只把日志数据发送到一个内部的channel中
func SendToChan(topic, data string) {
	msg := &logData{
		topic: topic,
		data:  data,
	}
	logDataChan <- msg
}

// sendToKafka 发送数据到kafka
func sendToKafka() {
	for {
		select {
		case ld := <-logDataChan:
			// 构造一个消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = ld.topic
			msg.Value = sarama.StringEncoder(ld.data)
			// 发送消息到kafka
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("Send msg failed,err :", err)
				return
			}
			fmt.Printf("pid:%v offset:%v\n", pid, offset)
		default:
			time.Sleep(time.Millisecond * 50)

		}

	}

}
