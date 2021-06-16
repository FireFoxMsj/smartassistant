package main

import (
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"time"
)

func main() {
	go smq.MinHeapQueue.Start()
	time.Sleep(5 * time.Second)
	item := &smq.Task{
		Value:    "orange",
		Priority: time.Now().Add(3 * time.Second).Unix(),
	}
	smq.MinHeapQueue.Push(item)

	select {

	}
}

