package procev

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

const (
	cnIdxProc = 0x1
	cnValProc = 0x1
)

type cbID struct {
	Idx uint32
	Val uint32
}

type cnMsg struct {
	ID    cbID
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
	EventUID      chan *EventUID
	EventGID      chan *EventGID
	EventSID      chan *EventSID
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
		Groups: cnIdxProc,
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
	listener.EventUID = make(chan *EventUID)
	listener.EventGID = make(chan *EventGID)
	listener.EventSID = make(chan *EventSID)
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

	msg.ID.Idx = cnIdxProc
	msg.ID.Val = cnValProc
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
	case idEventNone:
		event := &EventAck{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventAck <- event
	case idEventFork:
		event := &EventFork{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventFork <- event
	case idEventExec:
		event := &EventExec{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventExec <- event
	case idEventUID:
		event := &EventUID{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventUID <- event
	case idEventGID:
		event := &EventGID{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventGID <- event
	case idEventSID:
		event := &EventSID{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventSID <- event
	case idEventPtrace:
		event := &EventPtrace{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventPtrace <- event
	case idEventComm:
		event := &EventComm{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventComm <- event
	case idEventCoreDump:
		event := &EventCoreDump{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventCoreDump <- event
	case idEventExit:
		event := &EventExit{}
		binary.Read(buf, binary.LittleEndian, event)
		listener.EventExit <- event
	default:
		fmt.Printf("Unknown event type: 0x%08x\n", hdr.What)
	}
}
