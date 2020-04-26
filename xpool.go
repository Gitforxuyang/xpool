package xpool

import (
	"errors"
	"time"
)

var (
	configNilError   = errors.New("配置项不能为空")
	configParamError = errors.New("配置项错误")
)

type xpool struct {
	maxActive      int
	minActive      int
	maxIdle        int
	idleTimeOut    time.Duration
	maxWait        int
	maxWaitTimeOut time.Duration
	//当前等待人数
	currentWait int
	//当前资源数
	currentActive int
}

type conn struct {
}

func (xpool) New() (interface{}, error) {
	panic("implement me")
}

func (xpool) Release(interface{}) error {
	panic("implement me")
}

func (xpool) Close(interface{}) error {
	panic("implement me")
}

func (xpool) ShutDown() error {
	panic("implement me")
}

func NewXPool(configs *Configs) (XPool, error) {
	if configs == nil {
		return nil, configNilError
	}
	//最大活跃数不能为空
	if configs.MaxActive == 0 {
		return nil, configParamError
	}
	//如果最大等待数大于0 则必须设置最大等待时间
	if configs.MaxWait > 0 && configs.MaxWaitTime == 0 {
		return nil, configParamError
	}
	//如果最大空闲数大于0  则必须设置空闲超时时间
	if configs.MaxIdle > 0 && configs.IdleTimeOut == 0 {
		return nil, configParamError
	}
	p := xpool{
		maxActive:      configs.MaxActive,
		minActive:      configs.MinActive,
		maxWait:        configs.MaxWait,
		maxIdle:        configs.MaxIdle,
		maxWaitTimeOut: configs.MaxWaitTime,
		idleTimeOut:    configs.IdleTimeOut,
	}
	return &p, nil
}
