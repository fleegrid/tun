// +build !darwin,!linux

package tun

import "errors"

// NewDevice create a new TUN device
func NewDevice(config Config) (*Device, error) {
	return nil, errors.New("platform not supported, only darwin, linux are supported")
}
