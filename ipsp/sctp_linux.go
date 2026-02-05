//go:build linux && !386

package ipsp

import (
	"bytes"
	"encoding/binary"
	"io"
	"syscall"
	"unsafe"
)

func sockOpen() (int, error) {
	return syscall.Socket(syscall.AF_INET, syscall.SOCK_SEQPACKET, syscall.IPPROTO_SCTP)
}

func sockClose(fd int) {
	syscall.Shutdown(fd, syscall.SHUT_RDWR)
	syscall.Close(fd)
}

/*
func sockListen(fd int) error {
	return syscall.Listen(fd, syscall.SOMAXCONN)
}
*/

func sctpBindx(fd int, addr []byte) error {
	if _, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
		100, // SCTP_SOCKOPT_BINDX_ADD
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		0); e != 0 {
		return e
	}
	return nil
}

func sctpConnectx(fd int, addr []byte) (int, error) {
	opt := struct {
		numOstreams  uint16
		maxIstreams  uint16
		maxAttemts   uint16
		maxInitTimeo uint16
	}{numOstreams: 17}
	if _, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
		2, // SCTP_INITMSG
		uintptr(unsafe.Pointer(&opt)),
		unsafe.Sizeof(opt),
		0); e != 0 {
		return 0, e
	}

	/*
		on := 1
		if _, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
			uintptr(fd),
			syscall.IPPROTO_SCTP,
			0x20, // SCTP_RECVRCVINFO
			uintptr(unsafe.Pointer(&on)),
			unsafe.Sizeof(on),
			0); e != 0 {
			return 0, e
		}
	*/

	t, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
		110, // SCTP_SOCKOPT_CONNECTX
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(len(addr)),
		0)
	if e != 0 {
		return 0, e
	}

	peel := struct {
		aid int32
		sd  int32
	}{aid: int32(t)}
	l := unsafe.Sizeof(peel)
	if _, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
		102, // SCTP_SOCKOPT_PEELOFF
		uintptr(unsafe.Pointer(&peel)),
		uintptr(unsafe.Pointer(&l)),
		0); e != 0 {
		return 0, e
	}
	return int(peel.sd), nil
}

/*
	func sctpAccept(fd int) (nfd int, e error) {
		nfd, _, e = syscall.Accept(fd)
		return
	}
*/

func sctpSend(fd int, b []byte, sid uint16) (int, error) {
	buf := new(bytes.Buffer)
	hdr := syscall.Cmsghdr{
		Level: syscall.IPPROTO_SCTP,
		Type:  2, //SCTP_SNDINFO
	}
	hdr.SetLen(syscall.CmsgSpace(16))

	binary.Write(buf, binary.LittleEndian, hdr)
	binary.Write(buf, binary.LittleEndian, sid)            // stream ID(2 byte)
	binary.Write(buf, binary.LittleEndian, uint16(0x0001)) // flag(2 byte) = SCTP_UNORDERED
	binary.Write(buf, binary.BigEndian, uint32(3))         // PPID(4 byte) = M3UA(50331648)
	binary.Write(buf, binary.LittleEndian, uint32(0))      // context(4 byte) = empty
	binary.Write(buf, binary.LittleEndian, uint32(0))      // assoc ID(4 byte)

	return syscall.SendmsgN(fd, b, buf.Bytes(), nil, 0)
}

func sctpRecvmsg(fd int) (data []byte, e error) {
	info := make([]byte, syscall.CmsgSpace(32))
	data = make([]byte, 1500)
	n, on, _, _, e := syscall.Recvmsg(fd, data, info, 0)
	if e != nil {
		return
	}
	if n <= 0 && on <= 0 {
		e = io.EOF
		return
	}
	data = data[:n]

	/*
		if on > 0 {
			var msgs []syscall.SocketControlMessage
			if msgs, e = syscall.ParseSocketControlMessage(info); e != nil {
				return
			}
			for _, m := range msgs {
				if m.Header.Level == syscall.IPPROTO_SCTP &&
					m.Header.Type == 0x03 { // SCTP_RECVINFO
					buf := bytes.NewBuffer(m.Data)
					buf.Read(make([]byte, 2))                   // ???
					binary.Read(buf, binary.LittleEndian, &sid) // stream ID(2 byte)
					buf.Read(make([]byte, 2+2+4+4+4+4)) // ssn, flags, ppid, tsn, cumtsn, context
					binary.Read(buf, binary.LittleEndian, &aid) // assoc ID(4 byte)
				}
			}
		}
	*/
	return
}

func sctpGetladdrs(fd int) (unsafe.Pointer, int, error) {
	addr := struct {
		assoc int32
		num   uint32
		addrs [4096]byte
	}{}
	l := unsafe.Sizeof(addr)
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
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
		assoc int32
		num   uint32
		addrs [4096]byte
	}{}
	l := unsafe.Sizeof(addr)
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_SCTP,
		108, // SCTP_GET_PEER_ADDRS
		uintptr(unsafe.Pointer(&addr)),
		uintptr(unsafe.Pointer(&l)),
		0)
	if e != 0 {
		return nil, 0, e
	}
	return unsafe.Pointer(&addr.addrs), int(addr.num), nil
}
