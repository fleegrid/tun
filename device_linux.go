// +build linux

package tun

/*

#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <net/if.h>
#include <linux/if_tun.h>
#include <stdlib.h>
#include <string.h>

unsigned int if_name_size = IFNAMSIZ;

// create and initialize a STUN device, returns fd and name
//
// @param nout used for device name output, should be pre-allocated and larger than IFNAMESIZ
// @return fd if succeeded or -1 if error occurred
int new_tun(char *nout) {
	int fd, e;

	// open tun manager device
	fd = open("/dev/net/tun", O_RDWR);
	if (fd < 0) return fd;

	// set tun IFF
	struct ifreq tun_ifreq;
	tun_ifreq.ifr_ifru.ifru_flags = IFF_TUN | IFF_NO_PI;
	strcpy(tun_ifreq.ifr_ifrn.ifrn_name, "");
	e = ioctl(fd, TUNSETIFF, &tun_ifreq);
	if (e < 0) return -1;

	// output device name
	strcpy(nout, tun_ifreq.ifr_ifrn.ifrn_name);

	// set device persist
	e = ioctl(fd, TUNSETPERSIST, 0);
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

	if cFd, err = C.new_tun(cName); err != nil {
		return nil, fmt.Errorf("error while creating TUN: %v", err)
	}

	fd := int(cFd)
	name := C.GoString(cName)

	C.free(unsafe.Pointer(cName))

	dev = &Device{
		name:            name,
		ReadWriteCloser: os.NewFile(uintptr(fd), name),
	}
	return
}
