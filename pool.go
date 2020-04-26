package xpool

import "time"

type XPool interface {
	//获取一个新的资源
	New() (interface{}, error)
	//释放一个资源
	Release(interface{}) error
	//关闭一个资源
	Close(interface{}) error
	//程序关闭
	ShutDown() error
}

type Configs struct {
	//最大资源数
	MaxActive int
	//最小资源数。 在初始化的时候就完成加载。加载失败无法启动
	MinActive int
	//最大空闲数。应该要小于等于最大资源数。
	MaxIdle int
	//最大等待数。如果发生资源不够排队等待时，当派对的数量大于这个数则会报错
	MaxWait int
	//最大等待时间。等待时的超时时间
	MaxWaitTime time.Duration
	//空闲超时时间。当某个资源空闲时间超过设置阈值。则下次获取到时主动关闭它。防止已关闭的资源被使用
	IdleTimeOut time.Duration
	//创建资源的函数
	Factory func() (interface{}, error)
	//关闭一个资源的函数
	Close func(interface{}) error
}
