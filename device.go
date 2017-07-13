package tun

import "io"

// Device represents a TUN device
type Device struct {
	io.ReadWriteCloser
	name string
}

// Name returns device name
func (d *Device) Name() string {
	return d.name
}
