package pmon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

type testListener struct {
	t        *testing.T
	listener *ProcListener
	done     chan bool
	acks     []EventAck
	forks    []EventFork
	execs    []EventExec
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
			case event := <-tl.listener.EventAck:
				fmt.Printf("%T no=%d\n", event, event.No)
				tl.acks = append(tl.acks, *event)
			case event := <-tl.listener.EventFork:
				fmt.Printf("%T ppid=%d ptid=%d cpid=%d ctid=%d\n",
					event, event.ParentPid, event.ParentTid, event.ChildPid, event.ChildTid)
				tl.forks = append(tl.forks, *event)
			case event := <-tl.listener.EventExec:
				fmt.Printf("%T pid=%d tid=%d\n", event, event.Pid, event.Tid)
				tl.execs = append(tl.execs, *event)

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

func TestFork(t *testing.T) {
	parentPid := os.Getpid()
	tl := newTestListener(t)

	childPid, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		t.Fatal("Error on fork syscall")
	}

	if childPid == 0 {
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}

	tl.close()
	if len(tl.forks) < 1 {
		t.Errorf("Expected at least 1 fork event")
	}

	for _, event := range tl.forks {
		if event.ParentPid == uint32(parentPid) && event.ChildPid == uint32(childPid) {
			return
		}
	}

	t.Errorf("Not found expected fork event")
}

func TestExec(t *testing.T) {
	tl := newTestListener(t)
	cmd := exec.Command("sleep", "0.1")
	err := cmd.Run()
	if err != nil {
		t.Fatal("Error on exec command:", err)
	}

	pid := cmd.Process.Pid
	tl.close()

	if len(tl.execs) < 1 {
		t.Errorf("Expected at least 1 exec event")
	}

	for _, event := range tl.execs {
		if event.Pid == uint32(pid) {
			return
		}
	}

	t.Errorf("Not found expected fork event")
}