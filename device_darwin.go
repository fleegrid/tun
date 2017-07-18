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
		.sc_reserved = {0, 0, 0, 0, 0}
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
	"fmt"
	"os"
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
	name := C.GoStringN(cName, C.int(cNameLen-1))

	C.free(unsafe.Pointer(cName))

	dev = &Device{
		File: os.NewFile(uintptr(fd), name),
	}
	return
}
