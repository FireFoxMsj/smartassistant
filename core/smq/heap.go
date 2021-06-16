package smq

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskFunc func(task *Task) error
type WrapperFunc func(TaskFunc) TaskFunc

// Task 将场景转化为任务
type Task struct {
	ID       string
	Value    string // The Value of the item; arbitrary.
	Priority int64  // 使用
	// The index is needed by update and is maintained by the heap.Interface methods.
	index  int // The index of the item in the heap.
	f      TaskFunc
	Parent *Task

	wrappers []WrapperFunc
}

func NewTaskAt(f TaskFunc, t time.Time) *Task {
	return &Task{
		ID:       uuid.New().String(),
		Value:    "",
		Priority: t.Unix(),
		f:        f,
	}
}

func NewTask(f TaskFunc, delay time.Duration) *Task {
	return NewTaskAt(f, time.Now().Add(delay))
}

func (item *Task) WithParent(parent *Task) *Task {
	item.Parent = parent
	return item
}

func (item *Task) WithWrapper(wrappers ...WrapperFunc) *Task {
	item.wrappers = append(item.wrappers, wrappers...)
	return item
}

// Run 执行
// TODO
func (item *Task) Run() {
	fmt.Println("Run ", item.ToString())
	if item.f != nil {
		f := item.f
		for _, wrapper := range item.wrappers {
			f = wrapper(f)
		}
		if err := f(item); err != nil {
			log.Println("task run err:", err)
		}
	}
}

func (item *Task) ToString() string {
	return fmt.Sprintf("Task Value %s, Priority %d, index %d", item.Value, item.Priority, item.index)
}

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

var queueServeOnce sync.Once
var MinHeapQueue *QueueServe

func InitHeap() {
	queueServeOnce.Do(func() {
		MinHeapQueue = NewQueueServe()
	})
	go MinHeapQueue.Start()
}

func NewQueueServe() *QueueServe {
	var qs QueueServe
	qs.init()
	return &qs
}

type QueueServe struct {
	ticker            *time.Ticker
	pq                PriorityQueue
	SceneTaskIndexMap sync.Map // 正在执行的场景的id->index
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
	heap.Push(&qs.pq, task)
}

func (qs *QueueServe) Remove(i int) {
	if i < 0 || i >= qs.pq.Len() {
		return
	}
	heap.Remove(&qs.pq, i)
}

func (qs *QueueServe) Start() {
	pq, ticker := &qs.pq, qs.ticker
	defer ticker.Stop() // avoid leak

	for {
		select {
		case ct := <-ticker.C:
			fmt.Printf("current ticket at: %d:%d \n", ct.Minute(), ct.Second())
			if pq.Len() == 0 {
				ticker.Reset(sleepTickTime)
				continue
			}

			task := heap.Pop(pq).(*Task)
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
		}
	}
}
