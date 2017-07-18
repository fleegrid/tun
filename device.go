package tun

import (
	"os"
)

// Device represents a TUN device
type Device struct {
	*os.File
}
