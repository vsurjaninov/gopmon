package pmon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

type testListener struct {
	listener  *ProcListener
	done      chan bool
	acks      []EventAck
	forks     []EventFork
	execs     []EventExec
	uids      []EventUid
	gids      []EventGid
	sids      []EventSid
	comms     []EventComm
	coredumps []EventCoreDump
	exits     []EventExit
}

func newTestListener(t *testing.T) *testListener {
	tl := &testListener{
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
				fmt.Println(event)
				tl.acks = append(tl.acks, *event)
			case event := <-tl.listener.EventFork:
				fmt.Println(event)
				tl.forks = append(tl.forks, *event)
			case event := <-tl.listener.EventExec:
				fmt.Println(event)
				tl.execs = append(tl.execs, *event)
			case event := <-tl.listener.EventUid:
				fmt.Println(event)
				tl.uids = append(tl.uids, *event)
			case event := <-tl.listener.EventGid:
				fmt.Println(event)
				tl.gids = append(tl.gids, *event)
			case event := <-tl.listener.EventSid:
				fmt.Println(event)
				tl.sids = append(tl.sids, *event)
			case event := <-tl.listener.EventComm:
				fmt.Println(event)
				tl.comms = append(tl.comms, *event)
			case event := <-tl.listener.EventCoreDump:
				fmt.Println(event)
				tl.coredumps = append(tl.coredumps, *event)
			case event := <-tl.listener.EventExit:
				fmt.Println(event)
				tl.exits = append(tl.exits, *event)
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

func TestForkAndUidAndGidAndSidAndComm(t *testing.T) {
	parentPid := os.Getpid()
	tl := newTestListener(t)

	childPid, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		t.Fatal("Error on fork syscall")
	}

	childGid := 65534
	childUid := 1000
	childName := "0123456789ABCDEFG"

	if childPid == 0 {
		bytes := append([]byte(childName[:15]), 0)
		ptr := unsafe.Pointer(&bytes[0])
		_, _, err := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0)

		if err != 0 {
			fmt.Println("SYS_PRCTL PR_SET_NAME error:", err)
			os.Exit(1)
		}

		_, _, err = syscall.Syscall(syscall.SYS_SETSID, 0, 0, 0)
		if err != 0 {
			fmt.Println("SYS_SETSID error:", err)
			os.Exit(1)
		}

		_, _, err = syscall.Syscall(syscall.SYS_SETREGID, uintptr(childGid), uintptr(childGid), 0)
		if err != 0 {
			fmt.Println("SYS_SETREGID error:", err)
			os.Exit(1)
		}

		_, _, err = syscall.Syscall(syscall.SYS_SETREUID, uintptr(childUid), uintptr(childUid), 0)
		if err != 0 {
			fmt.Println("SYS_SETREUID error:", err)
			os.Exit(1)
		}

		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}

	tl.close()

	forkFound := false
	for _, event := range tl.forks {
		if event.ParentPid == uint32(parentPid) && event.ChildPid == uint32(childPid) {
			forkFound = true
		}
	}

	if !forkFound {
		t.Errorf("Not found expected fork event")
	}

	gidFound := false
	for _, event := range tl.gids {
		if event.Rgid == uint32(childGid) && event.Egid == uint32(childGid) {
			gidFound = true
		}
	}

	if !gidFound {
		t.Errorf("Not found expected gid event")
	}

	uidFound := false
	for _, event := range tl.uids {
		if event.Ruid == uint32(childUid) && event.Euid == uint32(childUid) {
			uidFound = true
		}
	}

	if !uidFound {
		t.Errorf("Not found expected uid event")
	}

	sidFound := false
	for _, event := range tl.sids {
		if event.Pid == uint32(childPid) {
			sidFound = true
		}
	}

	if !sidFound {
		t.Errorf("Not found expected sid event")
	}

	commFound := false
	for _, event := range tl.comms {
		if event.Pid == uint32(childPid) && event.getName() == childName[:15] {
			commFound = true
		}
	}

	if !commFound {
		t.Errorf("Not found expected comm event")
	}
}

func TestExecAndExitSuccess(t *testing.T) {
	tl := newTestListener(t)
	cmd := exec.Command("sleep", "0.1")
	if err := cmd.Run(); err != nil {
		t.Fatal("Error on exec command:", err)
	}

	pid := uint32(cmd.Process.Pid)
	tl.close()

	execFound := false
	for _, event := range tl.execs {
		if event.Pid == pid {
			execFound = true
		}
	}

	if !execFound {
		t.Errorf("Not found expected fork event")
	}

	exitFound := false
	for _, event := range tl.exits {
		if event.Pid == pid && event.Code == 0 {
			exitFound = true
		}
	}

	if !exitFound {
		t.Errorf("Not found expected exit event")
	}
}

func TestExecAndExitBySignalAndCoreDump(t *testing.T) {
	tl := newTestListener(t)
	cmd := exec.Command("sleep", "100")
	if err := cmd.Start(); err != nil {
		t.Fatal("Error on exec command:", err)
	}

	pid := uint32(cmd.Process.Pid)
	sig := syscall.SIGILL

	syscall.Kill(cmd.Process.Pid, sig)
	cmd.Wait()

	tl.close()

	execFound := false
	for _, event := range tl.execs {
		if event.Pid == pid {
			execFound = true
		}
	}

	if !execFound {
		t.Errorf("Not found expected fork event")
	}

	exitFound := false
	for _, event := range tl.exits {
		if event.Pid == pid && signalFromCode(event.Code) == sig {
			exitFound = true
		}
	}

	if !exitFound {
		t.Errorf("Not found expected exit event")
	}

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_CORE, &rLimit)
	if err != nil {
		t.Fatal("Error Getting Rlimit ", err)
	}

	if rLimit.Cur != 0 {
		t.Fatal("Core dumps not enabled!")
	}

	coreDumpFound := false
	for _, event := range tl.coredumps {
		if event.Pid == pid {
			coreDumpFound = true
		}
	}

	if !coreDumpFound {
		t.Errorf("Not found expected core dump event")
	}
}

func signalFromCode(code uint32) syscall.Signal {
	if code > 128 {
		return syscall.Signal(code - 128)
	}
	return syscall.Signal(code)
}
