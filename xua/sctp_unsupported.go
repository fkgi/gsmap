//go:build !linux || (linux && 386)

package xua

import "unsafe"

func sockOpenV4() (int, error) {
	return 0, nil
}

func sockOpenV6() (int, error) {
	return 0, nil
}

func sockClose(int) error {
	return nil
}

func sctpBindx(int, []byte) error {
	return nil
}

func sctpConnectx(int, []byte) error {
	return nil
}

func sctpSend(int, []byte, uint16, bool) (int, error) {
	return 0, nil
}

func sctpRecvmsg(int, []byte) (int, error) {
	return 0, nil
}

func sctpGetladdrs(int) (unsafe.Pointer, int, error) {
	return nil, 0, nil
}

func sctpGetpaddrs(int) (unsafe.Pointer, int, error) {
	return nil, 0, nil
}
