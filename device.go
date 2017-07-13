package tun

import "io"

// Device represents a TUN device
type Device struct {
	io.ReadWriteCloser
	name string
}
