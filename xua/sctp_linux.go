//go:build linux && !386

package xua

import (
	"bytes"
	"encoding/binary"
	"io"
	"syscall"
	"unsafe"
)

func sockOpenV4() (int, error) {
	return syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_STREAM, //syscall.SOCK_SEQPACKET,
		syscall.IPPROTO_SCTP)
}

func sockOpenV6() (int, error) {
	return syscall.Socket(
		syscall.AF_INET6,
		syscall.SOCK_STREAM, //syscall.SOCK_SEQPACKET,
		syscall.IPPROTO_SCTP)
}

func sockClose(fd int) error {
	syscall.Shutdown(fd, syscall.SHUT_RDWR)
	return syscall.Close(fd)
}

func sctpBindx(fd int, addr []byte) error {
	_, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		132, // SOL_SCTP
		100, // SCTP_SOCKOPT_BINDX_ADD
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		0)
	if e != 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, addr []byte) error {
	opt := struct {
		numOstreams  uint16
		maxIstreams  uint16
		maxAttemts   uint16
		maxInitTimeo uint16
	}{
		numOstreams: 17,
	}

	_, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		132, // SOL_SCTP
		2,   // SCTP_INITMSG
		uintptr(unsafe.Pointer(&opt)),
		uintptr(unsafe.Sizeof(opt)),
		0)
	if e != 0 {
		return e
	}
	_, _, e = syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		132, // SOL_SCTP
		110, // SCTP_SOCKOPT_CONNECTX
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		0)
	if e != 0 {
		return e
	}
	return nil
}

func sctpSend(fd int, b []byte, sid uint16, m3ua bool) (int, error) {
	hdr := &syscall.Cmsghdr{
		Level: syscall.IPPROTO_SCTP,
		Type:  2, //SCTP_SNDINFO
	}
	hdr.SetLen(syscall.CmsgSpace(16))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, hdr)
	binary.Write(buf, binary.LittleEndian, sid)            // stream ID = 0
	binary.Write(buf, binary.LittleEndian, uint16(0x0001)) // flag = SCTP_UNORDERED
	if m3ua {
		binary.Write(buf, binary.BigEndian, uint32(3)) // PPID = SUA(67108864), M3UA(50331648)
	} else {
		binary.Write(buf, binary.BigEndian, uint32(4)) // PPID = SUA(67108864), M3UA(50331648)
	}
	buf.Write(make([]byte, 8)) // context(4 bytes) = empty, assoc ID(4 bytes) = 0

	return syscall.SendmsgN(fd, b, buf.Bytes(), nil, 0)
}

func sctpRecvmsg(fd int, b []byte) (int, error) {
	n, on, _, _, e := syscall.Recvmsg(fd, b, make([]byte, 256), 0)
	if e == nil && n == 0 && on == 0 {
		e = io.EOF
	}
	return n, e
}

func sctpGetladdrs(fd int) (unsafe.Pointer, int, error) {
	addr := struct {
		_     int32
		num   uint32
		addrs [4096]byte
	}{}
	l := unsafe.Sizeof(addr)
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		132, // SOL_SCTP
		109, // SCTP_GET_LOCAL_ADDRS
		uintptr(unsafe.Pointer(&addr)),
		uintptr(unsafe.Pointer(&l)),
		0)
	if e != 0 {
		return nil, 0, e
	}
	return unsafe.Pointer(&addr.addrs), int(addr.num), nil
}

func sctpGetpaddrs(fd int) (unsafe.Pointer, int, error) {
	addr := struct {
		_     int32
		num   uint32
		addrs [4096]byte
	}{}
	l := unsafe.Sizeof(addr)
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		132, // SOL_SCTP
		108, // SCTP_GET_PEER_ADDRS
		uintptr(unsafe.Pointer(&addr)),
		uintptr(unsafe.Pointer(&l)),
		0)
	if e != 0 {
		return nil, 0, e
	}
	return unsafe.Pointer(&addr.addrs), int(addr.num), nil
}
