package xpool

import (
	"errors"
	"sync"
	"time"
)

var (
	configNilError        = errors.New("配置项不能为空")
	configParamError      = errors.New("配置项错误")
	waitListOverflowError = errors.New("等待队列溢出")
	waitTimeOutError      = errors.New("等待超时")
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
	factory       func() (interface{}, error)
	close         func(interface{}) error
	//资源池
	ch chan *conn
	sync.Mutex
	//是否已关闭，如果是。则在放回资源到资源池时直接close。防止塞会chan时chan已经close
	shutdown bool
}

type conn struct {
	//持有的资源
	c interface{}
	//最新活跃的时间
	time time.Time
}

func (m *xpool) New() (interface{}, error) {
reset:
	select {
	case conn := <-m.ch:
		//如果这个资源的上次活跃时间+设置的最大空闲时间小于当前时间，则关闭这个资源。重新获取
		if conn.time.Add(m.idleTimeOut).Before(time.Now()) {
			m.close(conn.c)
			m.currentActive--
			goto reset
		}
		return conn.c, nil
	default:
	}
	m.Lock()
	//如果当前资源数小于最大资源数。则直接创建新的资源即可
	if m.currentActive < m.maxActive {
		c, err := m.factory()
		if err != nil {
			return nil, err
		}
		conn := conn{c: c, time: time.Now()}
		m.currentActive++
		m.Unlock()
		return conn, nil
	}
	//如果当前资源数已经饱和。则判断当前等待队列是否饱和。如果饱和则返回错误。如果不饱和则排队
	if m.currentWait > m.maxActive {
		m.Unlock()
		return nil, waitListOverflowError
	} else {
		m.currentWait++
		m.Unlock()
		select {
		//因为从排队队列中拿到的只会是刚丢回来的资源。所以不判断是否过期
		case conn := <-m.ch:
			m.currentWait--
			return conn, nil
		case <-time.After(m.maxWaitTimeOut):
			m.currentWait--
			return nil, waitTimeOutError
		}
	}

}

func (m *xpool) Release(c interface{}) error {
	//如果资源池已关闭。则直接close
	if m.shutdown {
		m.close(c)
		m.currentActive--
		return nil
	}
	m.Lock()
	defer m.Unlock()
	//如果当前资源数小于最小激活数。则直接将资源放回池子里
	if m.currentActive < m.minActive {
		m.ch <- &conn{c: c, time: time.Now()}
		return nil
	}
	//如果当前资源数大于最小激活数但空闲资源数小于设置的可空闲资源数，则放回资源池
	if len(m.ch) < m.maxIdle {
		m.ch <- &conn{c: c, time: time.Now()}
		return nil
	}
	//当资源数大于最小激活数，且空闲资源数大于可空闲资源数时。则直接关闭此资源
	err := m.close(c)
	if err != nil {
		return err
	}
	m.currentActive--
	return nil
}

func (m *xpool) Close(c interface{}) error {
	m.currentActive--
	err := m.close(c)
	if err != nil {
		return err
	}
	return nil
}

func (m *xpool) ShutDown() error {
	m.shutdown = true
	for c := range m.ch {
		m.close(c.c)
		m.currentActive--
	}
	return nil
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
	if configs.Factory == nil {
		return nil, configParamError
	}
	if configs.Close == nil {
		return nil, configParamError
	}
	p := xpool{
		maxActive:      configs.MaxActive,
		minActive:      configs.MinActive,
		maxWait:        configs.MaxWait,
		maxIdle:        configs.MaxIdle,
		maxWaitTimeOut: configs.MaxWaitTime,
		idleTimeOut:    configs.IdleTimeOut,
		factory:        configs.Factory,
		close:          configs.Close,
	}
	for i := 1; i < p.minActive; i++ {
		c, err := p.factory()
		if err != nil {
			return nil, err
		}
		p.ch <- &conn{c: c, time: time.Now()}
		p.currentActive++
	}
	return &p, nil
}
