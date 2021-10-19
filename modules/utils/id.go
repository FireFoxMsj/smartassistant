package utils

import (
	"github.com/zhiting-tech/smartassistant/modules/config"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Tue Aug 31 2021 03:27:50 GMT+0800
const startTime = 1630351670

var (
	lastMacAddrByte byte

	once sync.Once
)

// getLastMacAddrByte 获取mac地址最后一个字节
func getLastMacAddrByte() byte {
	once.Do(func() {

		ifas, err := net.Interfaces()
		if err != nil {
			panic(err)
		}

		for _, ifa := range ifas {
			if len(ifa.HardwareAddr) >= 6 {
				lastMacAddrByte = ifa.HardwareAddr[5]
				break
			}
		}
	})
	return lastMacAddrByte
}

func SAAreaID() (id uint64) {

	// milliseconds from startTime
	timeShift := time.Since(time.Unix(startTime, 0)).Milliseconds()
	// last byte of mac addr
	lastByte := getLastMacAddrByte()

	// fmt.Printf("%064b\n", uint64(timeShift)<<22|uint64(lastByte)<<12|uint64(rand.Intn(1<<12)))
	return uint64(timeShift)<<22 | uint64(lastByte)<<12 | uint64(rand.Intn(1<<12))
}

func CloudAreaID(dataCenterID, workerID int) (id uint64) {
	// milliseconds from startTime
	timeShift := time.Since(time.Unix(startTime, 0)).Milliseconds()

	// fmt.Printf("%064b\n", uint64(timeShift)<<22 | 1<<21 | uint64(lastByte)<<12 | uint64(counterNext()))
	return uint64(timeShift)<<22 | 1<<21 | uint64(dataCenterID<<17) | uint64(workerID<<12) | uint64(counterNext())
}

// GetDataCenterAndWorkerID() 获取配置中的dataCenterID和WorkerID
func GetDataCenterAndWorkerID() (dataCenterID, workID int) {
	return config.GetConf().SmartCloud.DataCenterID, config.GetConf().SmartCloud.WorkerID
}

func IsSA(id uint64) bool {
	return id&3<<20 == 0
}

var (
	counter       int64
	lastTimeShift int64
)

func counterNext() int64 {
	// milliseconds from startTime
	timeShift := time.Since(time.Unix(startTime, 0)).Milliseconds()
	if timeShift != lastTimeShift { // FIXME data race
		defer atomic.CompareAndSwapInt64(&lastTimeShift, lastTimeShift, timeShift)
		if atomic.CompareAndSwapInt64(&counter, counter, 0) {
			return counter | 1<<12 - 1
		}
	}
	return atomic.AddInt64(&counter, 1) | 1<<12 - 1
}
