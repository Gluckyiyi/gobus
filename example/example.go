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
func (net *AsyncNet)BusType()string  {
	return "AsyncNet"
}
func (net *AsyncNet)BusNotification(data interface{})  {
	fmt.Println(fmt.Sprintf("收到异步订阅的消息:%v",data))
}
func (net *OnceNet)BusType()string  {
	return "onceNet"
}
func (net *OnceNet)BusNotification(data interface{})  {
	fmt.Println(fmt.Sprintf("收到一次性订阅的消息:%v",data))
}
func (net *Net)BusType()string  {
	return "net"
}
func (net *Net)BusNotification(data interface{})  {
	fmt.Println(fmt.Sprintf("收到订阅的消息:%v",data))
}

func main() {
	bus := gobus.New()
	net := &Net{}
	bus.Subscribe("test",net)
	bus.SubscribeOnce("test",&OnceNet{})
	bus.SubscribeAsync("test",&AsyncNet{},false)
	for i := 0 ;i < 8;i++ {
		time.Sleep(1 *time.Second)
		bus.Publish("test",i)
	}
	bus.WaitAsync()
}


