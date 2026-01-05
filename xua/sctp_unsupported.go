//go:build !linux || (linux && 386)

package xua

import "unsafe"

func sockOpen() (int, error) {
	return 0, nil
}

func sockClose(int) {}

/*
func sockListen(int) error {
	return nil
}
*/

func sctpBindx(int, []byte) error {
	return nil
}

func sctpConnectx(int, []byte) (int, error) {
	return 0, nil
}

/*
func sctpAccept(int) (int, error) {
	return 0, nil
}
*/

func sctpSend(int, []byte, uint16) (int, error) {
	return 0, nil
}

func sctpRecvmsg(int) ([]byte, error) {
	return nil, nil
}

func sctpGetladdrs(int) (unsafe.Pointer, int, error) {
	return nil, 0, nil
}

func sctpGetpaddrs(int) (unsafe.Pointer, int, error) {
	return nil, 0, nil
}
