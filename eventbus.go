package gobus

import (
	"fmt"
	"sync"
)

type busSubscriber interface {
	subscribe(topic string, receiver BusNotifier)
	subscribeAsync(topic string, receiver BusNotifier, transactional bool)
	subscribeOnce(topic string, receiver BusNotifier)
	subscribeOnceAsync(topic string, receiver BusNotifier)
	unsubscribe(topic string, receiver BusNotifier) error
}
type BusNotifier interface {
	BusNotification(data interface{})
	BusType() string
}
type busPublisher interface {
	publish(topic string, data interface{})
}
type busController interface {
	waitAsync()
}

type Bus interface {
	busController
	busSubscriber
	busPublisher
}

type EventBus struct {
	handlers map[string]map[string]*eventHandler
	lock     sync.Mutex // a lock for the map
	wg       sync.WaitGroup
}

type eventHandler struct {
	receiver      BusNotifier
	flagOnce      bool
	async         bool
	transactional bool
	sync.Mutex    // lock for an event handler - useful for running async callbacks serially
}

func New() *EventBus {
	return &EventBus{
		make(map[string]map[string]*eventHandler),
		sync.Mutex{},
		sync.WaitGroup{},
	}
}
func (bus *EventBus) Subscribe(topic string, receiver BusNotifier) {
	bus.subscribe(topic, receiver)
}
func (bus *EventBus) SubscribeAsync(topic string, receiver BusNotifier, transactional bool) {
	bus.subscribeAsync(topic, receiver, transactional)
}
func (bus *EventBus) SubscribeOnce(topic string, receiver BusNotifier) {
	bus.subscribeOnce(topic, receiver)
}
func (bus *EventBus) SubscribeOnceAsync(topic string, receiver BusNotifier) {
	bus.subscribeOnceAsync(topic, receiver)
}
func (bus *EventBus) Unsubscribe(topic string, receiver BusNotifier) error {
	return bus.unsubscribe(topic, receiver)
}
func (bus *EventBus) doSubscribe(topic, busType string, handler *eventHandler) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if bus.handlers[topic] == nil {
		bus.handlers[topic] = make(map[string]*eventHandler)
	}
	bus.handlers[topic][busType] = handler
}

//实现busSubscriber接口
func (bus *EventBus) subscribe(topic string, receiver BusNotifier) {
	bus.doSubscribe(topic, receiver.BusType(), &eventHandler{
		receiver:      receiver,
		flagOnce:      false,
		async:         false,
		transactional: false,
		Mutex:         sync.Mutex{},
	})
}
func (bus *EventBus) subscribeAsync(topic string, receiver BusNotifier, transactional bool) {
	bus.doSubscribe(topic, receiver.BusType(), &eventHandler{
		receiver:      receiver,
		flagOnce:      false,
		async:         true,
		transactional: transactional,
		Mutex:         sync.Mutex{},
	})
}
func (bus *EventBus) subscribeOnce(topic string, receiver BusNotifier) {
	bus.doSubscribe(topic, receiver.BusType(), &eventHandler{
		receiver:      receiver,
		flagOnce:      true,
		async:         false,
		transactional: false,
		Mutex:         sync.Mutex{},
	})
}
func (bus *EventBus) subscribeOnceAsync(topic string, receiver BusNotifier) {
	bus.doSubscribe(topic, receiver.BusType(), &eventHandler{
		receiver:      receiver,
		flagOnce:      true,
		async:         true,
		transactional: false,
		Mutex:         sync.Mutex{},
	})
}
func (bus *EventBus) unsubscribe(topic string, receiver BusNotifier) error {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	return bus.removeHandler(topic, receiver.BusType())
}
func (bus *EventBus) removeHandler(topic, busType string) error {
	if _, ok := bus.handlers[topic][busType]; ok {
		delete(bus.handlers[topic], busType)
		return nil
	}
	return fmt.Errorf("topic %s doesn't exist", topic)
}

//实现busController 接口
func (bus *EventBus) WaitAsync() {
	bus.waitAsync()
}
func (bus *EventBus) waitAsync() {
	bus.wg.Wait()
}
func (bus *EventBus) Publish(topic string, data interface{}) {
	bus.publish(topic, data)
}

//实现busPublisher接口
func (bus *EventBus) publish(topic string, data interface{}) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if handlers, ok := bus.handlers[topic]; ok {
		copyHandlers := copy(handlers)
		for key, value := range copyHandlers {
			if value.flagOnce {
				bus.removeHandler(topic, key)
			}
			if !value.async {
				bus.doPublish(value, data)
			} else {
				bus.wg.Add(1)
				if value.transactional {
					bus.lock.Unlock()
					value.Lock()
					bus.lock.Lock()
				}
				go bus.doPublishAsync(value, data)
			}
		}
	}
}
func (bus *EventBus) doPublishAsync(handler *eventHandler, data interface{}) {
	defer bus.wg.Done()
	if handler.transactional {
		defer handler.Unlock()
	}
	bus.doPublish(handler, data)
}
func (bus *EventBus) doPublish(handler *eventHandler, data interface{}) {
	handler.receiver.BusNotification(data)
}
func copy(source map[string]*eventHandler) map[string]*eventHandler {
	copyMap := make(map[string]*eventHandler)
	for key, value := range source {
		copyHandler := &eventHandler{
			receiver:      value.receiver,
			flagOnce:      value.flagOnce,
			async:         value.async,
			transactional: value.transactional,
			Mutex:         sync.Mutex{},
		}
		copyMap[key] = copyHandler
	}
	return copyMap
}
