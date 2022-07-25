package common

import "sync/atomic"

type AtomicBool struct {
	flag int32
}

func (b *AtomicBool) Set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), i)
}

func (b *AtomicBool) Get() bool {
	return atomic.LoadInt32(&(b.flag)) == 1
}

func (b *AtomicBool) GetAndSet(value bool) bool {
	var i int32 = 0
	if value {
		i = 1
	}
	if atomic.SwapInt32(&(b.flag), i) == 1 {
		return true
	}
	return false
}
