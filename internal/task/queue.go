package task

import (
	"container/heap"
	"context"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Task

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 时间越大越往后
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Task)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update 修改任务
func (pq *PriorityQueue) update(item *Task, value string, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

const (
	defaultTickTime = 200 * time.Microsecond
	sleepTickTime   = 5 * time.Minute
)

func NewQueueServe() *QueueServe {
	var qs QueueServe
	qs.init()
	return &qs
}

type QueueServe struct {
	mu     sync.Mutex
	ticker *time.Ticker
	pq     PriorityQueue
}

func (qs *QueueServe) init() {
	items := make(map[string]int)
	qs.pq = make(PriorityQueue, len(items))
	heap.Init(&qs.pq)
	qs.ticker = time.NewTicker(defaultTickTime)
}

func (qs *QueueServe) Push(task *Task) {
	qs.push(task)
	qs.ticker.Reset(defaultTickTime)
}

func (qs *QueueServe) push(task *Task) {
	qs.mu.Lock()
	heap.Push(&qs.pq, task)
	qs.mu.Unlock()
}

func (qs *QueueServe) pop() *Task {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return heap.Pop(&qs.pq).(*Task)

}

func (qs *QueueServe) Remove(i int) {
	if i < 0 || i >= qs.pq.Len() {
		return
	}
	qs.mu.Lock()
	heap.Remove(&qs.pq, i)
	qs.mu.Unlock()
}

func (qs *QueueServe) Len() int {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return qs.pq.Len()

}

func (qs *QueueServe) Start(ctx context.Context) {
	ticker := qs.ticker
	defer ticker.Stop() // avoid leak

	for {
		select {
		case ct := <-ticker.C:
			logrus.Debugf("current ticket at: %d:%d:%d", ct.Hour(), ct.Minute(), ct.Second())
			if qs.Len() == 0 {
				ticker.Reset(sleepTickTime)
				continue
			}
			task := qs.pop()
			now := time.Now()
			timeAt := now.Unix()

			if task.Priority > timeAt {
				nextTick := time.Unix(task.Priority, 0).Sub(now)
				ticker.Reset(nextTick)
				qs.push(task)
			} else {
				ticker.Reset(defaultTickTime)
				go task.Run()
			}
		case <-ctx.Done():
			logrus.Info("stopping task queue")
			return
		}
	}
}
