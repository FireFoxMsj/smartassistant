package pubsub

import (
	"sync"
)

type Pubsub struct {
	mu   sync.RWMutex
	subs map[string][]chan interface{}
}

func NewPubsub() *Pubsub {
	return &Pubsub{
		subs: make(map[string][]chan interface{}),
	}
}

func (ps *Pubsub) Subscribe(topic string, ch chan interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.subs[topic] = append(ps.subs[topic], ch)
}

func (ps *Pubsub) Publish(topic string, msg interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, sub := range ps.subs[topic] {
		select {
		case sub <- msg:
		default:
		}
	}
}

func (ps *Pubsub) UnSubscribe(topic string, ch chan interface{}) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs, ok := ps.subs[topic]
	if !ok {
		return
	}
	for i := 0; i < len(subs); i++ {
		if subs[i] == ch {
			ps.subs[topic] = append(subs[:i], subs[i+1:]...)
			close(ch)
		}
	}
}
