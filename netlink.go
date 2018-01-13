package procev

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

const (
	CN_IDX_PROC = 0x1
	CN_VAL_PROC = 0x1
)

type cbId struct {
	Idx uint32
	Val uint32
}

type cnMsg struct {
	Id    cbId
	Seq   uint32
	Ack   uint32
	Len   uint16
	Flags uint16
}

type eventMsgHeader struct {
	What      uint32
	Cpu       uint32
	Timestamp uint64
}

type ProcListener struct {
	fd int
	sa syscall.SockaddrNetlink

	EventAck      chan *EventAck
	EventFork     chan *EventFork
	EventExec     chan *EventExec
	EventUid      chan *EventUid
	EventGid      chan *EventGid
	EventSid      chan *EventSid
	EventPtrace   chan *EventPtrace
	EventComm     chan *EventComm
	EventCoreDump chan *EventCoreDump
	EventExit     chan *EventExit
	Error         chan error
	Stop          bool
}

func (listener *ProcListener) Connect() (err error) {
	listener.sa = syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: CN_IDX_PROC,
		Pid:    uint32(os.Getegid()),
	}

	listener.fd, err = syscall.Socket(
		syscall.AF_NETLINK,
		syscall.SOCK_DGRAM,
		syscall.NETLINK_CONNECTOR,
	)

	if err != nil {
		err = os.NewSyscallError("socket", err)
		return fmt.Errorf("cannot create netlink socket: %s", err)
	}

	err = syscall.Bind(listener.fd, &listener.sa)
	if err != nil {
		err = os.NewSyscallError("bind", err)
		syscall.Close(listener.fd)
		return fmt.Errorf("cannot bind netlink socket: %s", err)
	}

	err = listener.setListeningOp(1)
	if err != nil {
		syscall.Close(listener.fd)
		return
	}

	listener.EventAck = make(chan *EventAck)
	listener.EventFork = make(chan *EventFork)
	listener.EventExec = make(chan *EventExec)
	listener.EventUid = make(chan *EventUid)
	listener.EventGid = make(chan *EventGid)
	listener.EventSid = make(chan *EventSid)
	listener.EventPtrace = make(chan *EventPtrace)
	listener.EventComm = make(chan *EventComm)
	listener.EventCoreDump = make(chan *EventCoreDump)
	listener.EventExit = make(chan *EventExit)
	listener.Error = make(chan error)
	return
}

func (listener *ProcListener) Close() {
	listener.Stop = true
	if listener.fd == -1 {
		return
	}

	fmt.Println("stop listening")
	listener.setListeningOp(0)

	fmt.Println("close connection")
	syscall.Close(listener.fd)
}

func (listener *ProcListener) setListeningOp(op uint32) error {
	hdr := &syscall.NlMsghdr{}
	msg := &cnMsg{}

	size := binary.Size(msg) + binary.Size(op)

	hdr.Len = syscall.NLMSG_HDRLEN + uint32(size)
	hdr.Type = uint16(syscall.NLMSG_DONE)
	hdr.Flags = 0
	hdr.Seq = uint32(0)
	hdr.Pid = uint32(os.Getpid())

	msg.Id.Idx = CN_IDX_PROC
	msg.Id.Val = CN_VAL_PROC
	msg.Len = uint16(binary.Size(op))

	buf := bytes.NewBuffer(make([]byte, 0, hdr.Len))
	binary.Write(buf, binary.LittleEndian, hdr)
	binary.Write(buf, binary.LittleEndian, msg)
	binary.Write(buf, binary.LittleEndian, op)

	return syscall.Sendto(listener.fd, buf.Bytes(), 0, &listener.sa)
}

func (listener *ProcListener) ListenEvents() {
	rb := make([]byte, syscall.Getpagesize())
	fmt.Println("call recvfrom")

	for !listener.Stop {
		msglen, _, err := syscall.Recvfrom(listener.fd, rb, 0)
		if err != nil {
			listener.Error <- err
			continue
		}

		if msglen < syscall.NLMSG_HDRLEN {
			listener.Error <- fmt.Errorf("got short response from netlink")
			continue
		}

		fmt.Println("parse netlink message")
		nlMessages, err := syscall.ParseNetlinkMessage(rb[:msglen])
		if err != nil {
			listener.Error <- err
			continue
		}

		fmt.Println("receive events")
		for _, msg := range nlMessages {
			listener.handleRawEvent(msg.Data)
		}
	}
}

func (listener *ProcListener) handleRawEvent(data []byte) {
	buf := bytes.NewBuffer(data)
	msg := &cnMsg{}
	hdr := &eventMsgHeader{}

	binary.Read(buf, binary.LittleEndian, msg)
	binary.Read(buf, binary.LittleEndian, hdr)

	switch hdr.What {
	case PROC_EVENT_NONE:
		event := &EventAck{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventAck <- event
	case PROC_EVENT_FORK:
		event := &EventFork{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventFork <- event
	case PROC_EVENT_EXEC:
		event := &EventExec{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventExec <- event
	case PROC_EVENT_UID:
		event := &EventUid{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventUid <- event
	case PROC_EVENT_GID:
		event := &EventGid{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventGid <- event
	case PROC_EVENT_SID:
		event := &EventSid{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventSid <- event
	case PROC_EVENT_PTRACE:
		event := &EventPtrace{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventPtrace <- event
	case PROC_EVENT_COMM:
		event := &EventComm{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventComm <- event
	case PROC_EVENT_COREDUMP:
		event := &EventCoreDump{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventCoreDump <- event
	case PROC_EVENT_EXIT:
		event := &EventExit{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventExit <- event
	default:
		fmt.Printf("Unknown event type: 0x%08x\n", hdr.What)
	}
}
