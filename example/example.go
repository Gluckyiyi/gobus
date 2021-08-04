package main

import (
	"fmt"
	"github.com/Gluckyiyi/gobus"
	"time"
)

type Net struct {
}
type OnceNet struct {
}
type AsyncNet struct {
}

func (net *AsyncNet) BusType() string {
	return "AsyncNet"
}
func (net *AsyncNet) BusNotification(topic string, data interface{}) {
	fmt.Println(fmt.Sprintf("收到异步订阅(%s)的消息:%v", topic, data))
}
func (net *OnceNet) BusType() string {
	return "onceNet"
}
func (net *OnceNet) BusNotification(topic string, data interface{}) {
	fmt.Println(fmt.Sprintf("收到一次性订阅(%s)的消息:%v", topic, data))
}
func (net *Net) BusType() string {
	return "net"
}
func (net *Net) BusNotification(topic string, data interface{}) {
	fmt.Println(fmt.Sprintf("收到订阅(%s)的消息:%v", topic,data))
}

func main() {
	bus := gobus.New()
	net := &Net{}
	bus.Subscribe("test", net)
	bus.SubscribeOnce("test", &OnceNet{})
	bus.SubscribeAsync("test", &AsyncNet{}, false)
	for i := 0; i < 8; i++ {
		time.Sleep(1 * time.Second)
		bus.Publish("test", i)
	}
	bus.WaitAsync()
}
