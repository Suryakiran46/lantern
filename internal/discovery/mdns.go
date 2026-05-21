package discovery

type Discoverer interface {
	Start() error
	Stop()
	UpdateStatus(status string)
	DeviceChan() <-chan DeviceEvent // join / leave / update events
	AddGossipPeer(ip string)        // called on manual connect
}
