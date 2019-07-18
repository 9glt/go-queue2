package queue2

import (
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	q := New()

	q.WriteSecondary("s3")
	q.WriteSecondary("s4")

	q.WritePrimary("s1")
	q.WritePrimary("s2")
	q.Switch()

	var switched bool
	q.OnSwitch = func() {
		switched = true
	}

	var s []string

	unsub := q.Subscribe(func(i Item) {
		s = append(s, i.(string))
	})

	time.Sleep(100 * time.Millisecond)
	q.WriteSecondary("s5")
	unsub()
	time.Sleep(100 * time.Millisecond)
	if len(s) != 5 {
		t.Fatal()
	}
	if s[0] != "s1" && s[1] != "s2" && s[2] != "s3" && s[4] != "s4" {
		t.Fatal()
	}
	if !switched {
		t.Error()
	}
}

func TestPrimary(t *testing.T) {
	q := New()

	q.WritePrimary("s1")
	q.WritePrimary("s2")

	q.WriteSecondary("s3")
	q.WriteSecondary("s4")

	var s []string

	q.ProcessPrimarySync(func(i Item) {
		s = append(s, i.(string))
	})

	if len(s) != 2 {
		t.Fatal()
	}

	q.ProcessSecondarySync(func(i Item) {
		s = append(s, i.(string))
	})
	if len(s) != 4 {
		t.Fatal()
	}
}

func TestNotifyOnLimit(t *testing.T) {
	q := New()
	q.PCap = 1
	q.SCap = 1

	var a, b int

	q.PC = func(num int) {
		a = num
	}
	q.SC = func(num int) {
		b = num
	}
	q.WritePrimary("p1")
	q.WritePrimary("p2")

	if a != 2 && b != 0 {
		t.Fatal()
	}
	q.WriteSecondary("p3")
	q.WriteSecondary("p4")

	if a != 2 && b != 2 {
		t.Fatal()
	}
}

func TestSubscribe(t *testing.T) {
	q := New()
	var s string
	unsub := q.Subscribe(func(i Item) {
		s = i.(string)
	})
	unsub()
	q.WritePrimary("p1")
	if s != "" {
		t.Fatal()
	}
}
