package main

import (
	"fmt"
	"time"
	"xpool"
)

type Client struct {
}

func (m *Client) Doit() {
	fmt.Println("doit")
}
func (m *Client) Close() {
	fmt.Println("close")
}
func main() {
	config := xpool.Configs{
		MaxActive:   10,
		MinActive:   2,
		MaxWaitTime: time.Second * 1,
		MaxIdle:     2,
		MaxWait:     10,
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
	client, _ := pool.New()
	c, ok := client.(*Client)
	fmt.Println(ok)
	c.Doit()
	pool.Release(client)
	pool.ShutDown()
}
