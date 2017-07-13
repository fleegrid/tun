package tun

// Config config for a tun device
type Config struct {
	// user id for tun device, only valid for linux
	User uint
	// group id for tun device, only valid for linux
	Group uint
	// name for tun device, only valid for linux, empty for default name
	Name string
}
