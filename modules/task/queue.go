package task

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// A priorityQueue implements heap.Interface and holds Items.
type priorityQueue []*Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// 时间越大越往后
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Task)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update 修改任务
func (pq *priorityQueue) update(item *Task, value string, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

const (
	defaultTickTime = 200 * time.Microsecond
	sleepTickTime   = 5 * time.Minute
)

func newQueueServe() *queueServe {
	var qs queueServe
	qs.init()
	return &qs
}

type queueServe struct {
	mu     sync.Mutex
	ticker *time.Ticker
	pq     priorityQueue
}

func (qs *queueServe) init() {
	items := make(map[string]int)
	qs.pq = make(priorityQueue, len(items))
	heap.Init(&qs.pq)
	qs.ticker = time.NewTicker(defaultTickTime)
}

func (qs *queueServe) push(task *Task) {
	qs._push(task)
	qs.ticker.Reset(defaultTickTime)
}

func (qs *queueServe) _push(task *Task) {
	qs.mu.Lock()
	heap.Push(&qs.pq, task)
	qs.mu.Unlock()
}

func (qs *queueServe) _pop() *Task {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return heap.Pop(&qs.pq).(*Task)

}

func (qs *queueServe) _remove(i int) {
	if i < 0 || i >= qs.pq.Len() {
		return
	}
	qs.mu.Lock()
	heap.Remove(&qs.pq, i)
	qs.mu.Unlock()
}

func (qs *queueServe) _len() int {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	return qs.pq.Len()

}

func (qs *queueServe) start(ctx context.Context) {
	ticker := qs.ticker
	defer ticker.Stop() // avoid leak

	for {
		select {
		case ct := <-ticker.C:
			logger.Debugf("current ticket at: %d:%d:%d", ct.Hour(), ct.Minute(), ct.Second())
			if qs._len() == 0 {
				ticker.Reset(sleepTickTime)
				continue
			}
			task := qs._pop()
			now := time.Now()
			timeAt := now.Unix()

			if task.Priority > timeAt {
				nextTick := time.Unix(task.Priority, 0).Sub(now)
				ticker.Reset(nextTick)
				qs._push(task)
			} else {
				ticker.Reset(defaultTickTime)
				go task.Run()
			}
		case <-ctx.Done():
			logger.Info("stopping task queue")
			return
		}
	}
}
