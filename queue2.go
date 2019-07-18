package queue2

import (
	"fmt"
	"sync"
)

type Item interface{}

func New() *Q {
	t := new(Q)
	t.l = &t.l1
	t.mu = &sync.Mutex{}
	t.cond = sync.NewCond(t.mu)

	t.PCap = 10000
	t.SCap = 20000
	return t
}

type Q struct {
	l  *localL
	l1 localL
	l2 localL

	sw   bool
	swd  bool
	scnt int

	mu   *sync.Mutex
	cond *sync.Cond
	PC   func(int)
	SC   func(int)
	PCap int
	SCap int

	OnSwitch func()
}

func (t *Q) WritePrimary(m Item) {
	t.mu.Lock()
	t.l1.Write1(m)
	if t.l1.Len() > t.PCap {
		if t.PC != nil {
			t.PC(t.l1.Len())
		}
	}
	t.cond.Broadcast()
	t.mu.Unlock()
}

func (t *Q) WriteSecondary(m Item) {
	t.mu.Lock()
	t.l2.Write1(m)
	if t.l2.Len() > t.SCap {
		if t.SC != nil {
			t.SC(t.l2.Len())
		}
	}
	if t.sw == true {
		t.cond.Broadcast()
	}
	t.mu.Unlock()
}

func (t *Q) Switch() {
	t.mu.Lock()
	t.sw = true
	t.mu.Unlock()
}

func (t *Q) process() (Item, error) {
	m, e := t.l.process()
	if t.sw && e != nil && !t.swd {
		t.swd = true
		if t.OnSwitch != nil {
			t.OnSwitch()
		}
		t.l = &t.l2
		m, e = t.l.process()
	}
	return m, e
}

func (t *Q) ProcessPrimary() (Item, error) {
	t.mu.Lock()
	m, e := t.l1.process()
	t.mu.Unlock()
	return m, e
}

func (t *Q) ProcessSecondary() (Item, error) {
	t.mu.Lock()
	m, e := t.l2.process()
	t.mu.Unlock()
	return m, e
}

func (t *Q) ProcessPrimarySync(fn func(i Item)) {
	for i, e := t.ProcessPrimary(); e == nil; i, e = t.ProcessPrimary() {
		fn(i)
	}
}

func (t *Q) ProcessSecondarySync(fn func(i Item)) {
	for i, e := t.ProcessSecondary(); e == nil; i, e = t.ProcessSecondary() {
		fn(i)
	}
}

func (t *Q) process2() Item {
	t.mu.Lock()
	m, e := t.process()
	if e != nil {
		t.cond.Wait()
		m, e = t.process()
	}
	t.mu.Unlock()
	return m
}

func (t *Q) Subscribe(fn func(Item)) (ubsub func()) {
	t.scnt++
	ch := make(chan struct{})
	go func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return
				}
			default:
			}
			fn(t.process2())
		}
	}()
	return func() { close(ch) }
}

// linked list node
type localLI struct {
	Value Item
	next  *localLI
}

// linked list
type localL struct {
	head *localLI
	tail *localLI
	cnt  int
	mu   *sync.Mutex
}

func (l *localL) process() (Item, error) {
	msg := l.head
	if msg != nil {
		l.cnt--
		l.head = msg.next
		if l.head == nil {
			l.tail = nil
		}
	} else {
		return "", fmt.Errorf("error")
	}
	return msg.Value, nil
}

func (l *localL) Write1(v Item) {
	li := &localLI{v, nil}
	if l.head == nil {
		l.head = li
		l.tail = li
	} else {
		l.tail.next = li
		l.tail = li
	}
	l.cnt++
}

func (l *localL) Len() int {
	return l.cnt
}
