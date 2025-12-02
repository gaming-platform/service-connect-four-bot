package rpcclient

import (
	"strconv"
	"sync/atomic"
)

type counter struct {
	value int64
}

func newCounter() *counter {
	return &counter{}
}

func (c *counter) next() int64 {
	return atomic.AddInt64(&c.value, 1)
}

func (c *counter) nextString() string {
	return strconv.FormatInt(c.next(), 10)
}
