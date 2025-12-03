package rpcclient

import (
	"testing"
)

func TestCounterNext(t *testing.T) {
	c := newCounter()
	if got := c.next(); got != 1 {
		t.Fatalf("first next = %d; want 1", got)
	}
	if got := c.next(); got != 2 {
		t.Fatalf("second next = %d; want 2", got)
	}
}

func TestCounterNextString(t *testing.T) {
	c := newCounter()
	if got := c.nextString(); got != "1" {
		t.Fatalf("nextString = %q; want \"1\"", got)
	}
	if got := c.nextString(); got != "2" {
		t.Fatalf("nextString = %q; want \"2\"", got)
	}
}

func TestCounterNextConcurrent(t *testing.T) {
	c := newCounter()
	n := 10000
	ch := make(chan int64, n)

	for i := 0; i < n; i++ {
		go func() {
			ch <- c.next()
		}()
	}

	vs := make(map[int64]bool, n)
	for i := 0; i < n; i++ {
		v := <-ch
		if _, ok := vs[v]; ok {
			t.Fatalf("duplicate value: %d", v)
		}
		vs[v] = true
	}
}
