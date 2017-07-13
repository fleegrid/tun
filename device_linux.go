// +build linux

package tun

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	cIFFTUN  = 0x0001
	cIFFTAP  = 0x0002
	cIFFNOPI = 0x1000
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func ioctl(fd uintptr, request int, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

// NewDevice create a new TUN device
func NewDevice(config Config) (dev *Device, err error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	name, err := createDevice(file.Fd(), config.Name, cIFFTUN|cIFFNOPI)
	if err != nil {
		return nil, err
	}

	if err = setDeviceOptions(file.Fd(), config); err != nil {
		return nil, err
	}

	dev = &Device{ReadWriteCloser: file, name: name}
	return
}

func createDevice(fd uintptr, ifName string, flags uint16) (createdIFName string, err error) {
	var req ifReq
	req.Flags = flags
	copy(req.Name[:], ifName)

	err = ioctl(fd, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		return
	}

	createdIFName = strings.Trim(string(req.Name[:]), "\x00")
	return
}

func setDeviceOptions(fd uintptr, config Config) (err error) {

	// Device Permissions
	if config.User != 0 || config.Group != 0 {

		// Set Owner
		if err = ioctl(fd, syscall.TUNSETOWNER, uintptr(config.User)); err != nil {
			return
		}

		// Set Group
		if err = ioctl(fd, syscall.TUNSETGROUP, uintptr(config.Group)); err != nil {
			return
		}
	}

	// no persistent
	value := 0
	return ioctl(fd, syscall.TUNSETPERSIST, uintptr(value))
}
