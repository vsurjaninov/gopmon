package main

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

type CnConnection struct {
    fd int
    sa syscall.SockaddrNetlink
}

func (cn *CnConnection) Connect() (err error) {
    cn.sa = syscall.SockaddrNetlink{
        Family: syscall.AF_NETLINK,
        Groups: CN_IDX_PROC,
        Pid: uint32(os.Getegid()),
    }

    cn.fd, err = syscall.Socket(
        syscall.AF_NETLINK,
        syscall.SOCK_DGRAM,
        syscall.NETLINK_CONNECTOR,
    )

    if err != nil {
        err = os.NewSyscallError("socket", err)
        return fmt.Errorf("cannot create netlink socket: %s", err)
    }

    err = syscall.Bind(cn.fd, &cn.sa)
    if err != nil {
        err = os.NewSyscallError("bind", err)
        syscall.Close(cn.fd)
        return fmt.Errorf("cannot bind netlink socket: %s", err)
    }

    return cn.setListeningOp(1)
}

func (cn *CnConnection) Close() {
    if cn.fd != -1 {
        fmt.Println("stop listening")
        cn.setListeningOp(0)
        fmt.Println("close connection")
        syscall.Close(cn.fd)
    }
}

func (cn *CnConnection) setListeningOp(op uint32) error {
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

    return syscall.Sendto(cn.fd, buf.Bytes(), 0, &cn.sa)
}

type RawMsg struct {
    Data  []byte
    Error error
}

func (cn *CnConnection) RecvRawProcEvents(events chan RawMsg) {
    rb := make([]byte, syscall.Getpagesize())
    fmt.Println("call recvfrom")

    for {
        msglen, _, err := syscall.Recvfrom(cn.fd, rb, 0)
        if err != nil {
            events <- RawMsg{Data: nil, Error: err}
            continue
        }

        if msglen < syscall.NLMSG_HDRLEN {
            events <- RawMsg{
                Data:  nil,
                Error: fmt.Errorf("got short response from netlink"),
            }
            continue
        }

        nlMessages, err := syscall.ParseNetlinkMessage(rb[:msglen])
        if err != nil {
            events <- RawMsg{Data: nil, Error: err}
            continue
        }

        for _, msg := range nlMessages {
            events <- RawMsg{Data: msg.Data}
        }
    }
}
