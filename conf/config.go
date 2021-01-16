package conf

type AppConf struct {
	KafkaConfig `ini:"kafka"`
	EtcdConfig  `ini:"etcd"`
}

type KafkaConfig struct {
	Address     string `ini:"address"`
	ChanMaxSize int    `ini:"chanMaxSize"`
}

type EtcdConfig struct {
	Address string `ini:"address"`
	Timeout int    `ini:"timeout"`
	Key     string `ini:"collect_log_key"`
}
