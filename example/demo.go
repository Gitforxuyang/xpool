package main

import (
	"fmt"
	"sync/atomic"
	"time"
	"xpool"
)

type Client struct {
}

var doitCount int32 = 0

func (m *Client) Doit() {
	time.Sleep(time.Millisecond * 800)
	fmt.Println("doit")
	atomic.AddInt32(&doitCount, 1)
}
func (m *Client) Close() {
	fmt.Println("close")
}
func main() {
	config := xpool.Configs{
		MaxActive:   10,
		MinActive:   5,
		MaxWaitTime: time.Second * 1,
		MaxIdle:     2,
		MaxWait:     20,
		IdleTimeOut: time.Second * 10,
		Factory: func() (i interface{}, e error) {
			return &Client{}, nil
		},
		Close: func(i interface{}) error {
			c := i.(*Client)
			c.Close()
			return nil
		},
	}
	pool, err := xpool.NewXPool(&config)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 30; i++ {
		go func() {
			client, err := pool.New()
			if err != nil {
				fmt.Println(err)
				return
			}
			c, _ := client.(*Client)
			c.Doit()
			pool.Release(c)
		}()
	}
	time.Sleep(time.Second * 10)
	fmt.Println(atomic.LoadInt32(&doitCount))
}
