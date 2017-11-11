package pmon

import (
	"fmt"
	"testing"
	"time"
)

type testListener struct {
	t        *testing.T
	listener *ProcListener
	done     chan bool
	acks     []EventAck
}

func newTestListener(t *testing.T) *testListener {
	tl := &testListener{
		t:        t,
		listener: &ProcListener{},
		done:     make(chan bool, 1),
	}

	err := tl.listener.Connect()
	if err != nil {
		t.Fatal("Failed connect")
	}

	go tl.listener.ListenEvents()
	go func() {
		for {
			select {
			case <-tl.done:
				return
			case <-tl.listener.Error:
				t.Fatal("Error on recv")
			case ev := <-tl.listener.EventAck:
				fmt.Printf("%T no=%d\n", ev, ev.No)
				tl.acks = append(tl.acks, *ev)
			}
		}
	}()

	return tl
}

func (tl *testListener) close() {
	pause := 100 * time.Millisecond
	time.Sleep(pause)
	tl.done <- true
	tl.listener.Close()
	time.Sleep(pause)
}

func TestAck(t *testing.T) {
	tl := newTestListener(t)
	tl.close()

	if len(tl.acks) != 1 && tl.acks[0].No != 0 {
		t.Errorf("Expected 1 ack event")
	}
}
