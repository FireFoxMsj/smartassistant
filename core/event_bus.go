package core

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
)

// TODO 细节完善
// 1）退出时要做waitegroup处理

type Event struct {
	EventType string `json:"event_type"`
	Data      M      `json:"data"`
	Origin    string `json:"origin"` // 事件来源 todo
}

type listenFunc func(event Event) error

type listener map[string][]listenFunc

type EventBus struct {
	listeners listener
	eventChan []chan Event
	lock      sync.Mutex // map lock
}

// NewEventBus
func NewEventBus() *EventBus {
	eb := EventBus{
		listeners: make(listener),
		eventChan: make([]chan Event, runtime.NumCPU()),
		lock:      sync.Mutex{},
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		eb.eventChan[i] = make(chan Event, 1)
		go eb.start(eb.eventChan[i])
	}

	return &eb
}

func (bus *EventBus) start(eventChan chan Event) {
	defer func() {
		// TODO 这个错误需要反馈给用户？
		if r := recover(); r != nil {
			log.Printf("[EventBus]系统出错 %v", r)
		}
	}()

	for {
		select {
		case event, ok := <-eventChan:
			if ok {
				go bus.asynFire(event)
			}
		}
	}
}

// asynFire 执行监听逻辑，
// TODO 错误信息应该通过queue+websocket给用户
func (bus *EventBus) asynFire(event Event) {
	listeners, _ := bus.listeners[event.EventType]
	for _, cb := range listeners {
		if err := cb(event); err != nil {
			log.Println("asynFire err", err.Error())
		}
	}
}

// Fire 触发事情机制
func (bus *EventBus) Fire(eventType string, eventData M) error {
	if ok := bus.hasEvent(eventType); !ok {
		return fmt.Errorf("没有此事件 %s", eventType)
	}

	event := Event{
		EventType: eventType,
		Data:      eventData,
	}

	i := rand.Intn(runtime.NumCPU())
	bus.eventChan[i] <- event
	return nil
}

// SyncFire 同步任务执行
func (bus *EventBus) SyncFire(eventType string, eventData M) error {
	if _, ok := bus.listeners[eventType]; !ok {
		return fmt.Errorf("没有此事件 %s", eventType)
	}

	event := Event{
		EventType: eventType,
		Data:      eventData,
	}

	listeners, _ := bus.listeners[event.EventType]
	for _, cb := range listeners {
		if err := cb(event); err != nil {
			return err
		}
	}
	return nil
}

func (bus *EventBus) hasEvent(eventType string) bool {
	_, ok := bus.listeners[eventType]
	return ok
}

// Listen 注册事件与事件回调
func (bus *EventBus) Listen(eventType string, listeners ...listenFunc) {
	bus.lock.Lock()
	defer bus.lock.Unlock()
	if bus.hasEvent(eventType) {
		bus.listeners[eventType] = append(bus.listeners[eventType], listeners...)
	} else {
		bus.listeners[eventType] = listeners
	}
}
