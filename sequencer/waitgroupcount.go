package sequencer

import (
	"sync"
	"sync/atomic"
)

type WaitGroupCount struct {
	sync.WaitGroup
	count atomic.Int32
}

func (wg *WaitGroupCount) Add(delta int) {
	wg.count.Add(int32(delta))
	wg.WaitGroup.Add(delta)
}

func (wg *WaitGroupCount) Done() {
	wg.count.Add(-1)
	wg.WaitGroup.Done()
}

func (wg *WaitGroupCount) Count() int {
	return int(wg.count.Load())
}
