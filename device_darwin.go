// +build darwin

package tun

/*

#include <sys/types.h>
#include <sys/socket.h>
#include <sys/ioctl.h>
#include <sys/kern_control.h>
#include <sys/kern_event.h>
#include <sys/syscall.h>
#include <net/if.h>
#include <net/if_utun.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>

unsigned int if_name_size = IFNAMSIZ;

// create and initialize a STUN device, returns fd and name
//
// @param nout used for device name output, should be pre-allocated and larger or equal than IFNAMESIZ
// @param lout used for nout length input and actual length output
// @return fd if succeeded or -1 if error occurred
int new_tun(char *nout, unsigned int *lout) {
	int fd, e;

	// create system control socket
	fd = socket(PF_SYSTEM, SOCK_DGRAM, SYSPROTO_CONTROL);
	if (fd < 0) return fd;

	// get ctl_id
	struct ctl_info utun_ctl_info;
	strcpy(utun_ctl_info.ctl_name, UTUN_CONTROL_NAME);
	e = ioctl(fd, CTLIOCGINFO, &utun_ctl_info);
	if (e < 0) return -1;

	// connect kernel
	struct sockaddr_ctl utun_addr_ctl = (struct sockaddr_ctl){
		.sc_len = sizeof(struct sockaddr_ctl),
		.sc_family = AF_SYSTEM,
		.ss_sysaddr = AF_SYS_CONTROL,
		.sc_id = utun_ctl_info.ctl_id,
		.sc_unit = 0,
	};
	e = connect(fd, (struct sockaddr *)&utun_addr_ctl, sizeof(struct sockaddr_ctl));
	if (e < 0) return -1;

	// output device name
	e = getsockopt(fd, SYSPROTO_CONTROL, UTUN_OPT_IFNAME, nout, lout);
	if (e < 0) return -1;

	return fd;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

// NewDevice create a new TUN device
func NewDevice() (dev *Device, err error) {
	cNameLen := C.if_name_size
	cName := (*C.char)(C.malloc(C.size_t(cNameLen)))
	cFd := C.int(0)

	if cFd, err = C.new_tun(cName, &cNameLen); err != nil {
		return nil, fmt.Errorf("error while creating TUN: %v", err)
	}

	fd := int(cFd)
	name := C.GoStringN(cName, C.int(cNameLen))

	C.free(unsafe.Pointer(cName))

	return &Device{
		name: name,
		ReadWriteCloser: &tunReadCloser{
			f: os.NewFile(uintptr(fd), name),
		},
	}, nil
}

// tunReadCloser is a hack to work around the first 4 bytes "packet
// information" because there doesn't seem to be an IFF_NO_PI for darwin.
type tunReadCloser struct {
	f io.ReadWriteCloser

	rMu  sync.Mutex
	rBuf []byte

	wMu  sync.Mutex
	wBuf []byte
}

var _ io.ReadWriteCloser = (*tunReadCloser)(nil)

func (t *tunReadCloser) Read(to []byte) (int, error) {
	t.rMu.Lock()
	defer t.rMu.Unlock()

	if cap(t.rBuf) < len(to)+4 {
		t.rBuf = make([]byte, len(to)+4)
	}
	t.rBuf = t.rBuf[:len(to)+4]

	n, err := t.f.Read(t.rBuf)
	copy(to, t.rBuf[4:])
	return n - 4, err
}

func (t *tunReadCloser) Write(from []byte) (int, error) {

	if len(from) == 0 {
		return 0, syscall.EIO
	}

	t.wMu.Lock()
	defer t.wMu.Unlock()

	if cap(t.wBuf) < len(from)+4 {
		t.wBuf = make([]byte, len(from)+4)
	}
	t.wBuf = t.wBuf[:len(from)+4]

	// Determine the IP Family for the NULL L2 Header
	ipVer := from[0] >> 4
	if ipVer == 4 {
		t.wBuf[3] = syscall.AF_INET
	} else if ipVer == 6 {
		t.wBuf[3] = syscall.AF_INET6
	} else {
		return 0, errors.New("Unable to determine IP version from packet")
	}

	copy(t.wBuf[4:], from)

	n, err := t.f.Write(t.wBuf)
	return n - 4, err
}

func (t *tunReadCloser) Close() error {
	// lock to make sure no read/write is in process.
	t.rMu.Lock()
	defer t.rMu.Unlock()
	t.wMu.Lock()
	defer t.wMu.Unlock()

	return t.f.Close()
}
