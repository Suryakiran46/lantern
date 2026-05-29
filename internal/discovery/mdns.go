package discovery

import (
	"context"
	"log"
	"os"
	"sync"
	"net"

	"github.com/Suryakiran46/lantern/internal/config"
	"github.com/grandcat/zeroconf"
)

const (
	serviceName = "_lantern._tcp"
	domain      = "local."
)

type MDNSDiscoverer struct {
	cfg      config.AppConfig
	devices  map[string]config.Device
	mu       sync.Mutex
	deviceCh chan config.DeviceEvent
	stopCh   chan struct{}
	server   *zeroconf.Server
	localIP	string
}

func NewDiscoverer(cfg config.AppConfig) *MDNSDiscoverer {
	return &MDNSDiscoverer{
		cfg:      cfg,
		devices:  make(map[string]config.Device),
		deviceCh: make(chan config.DeviceEvent, 32),
		stopCh:   make(chan struct{}),
	}
}

func (d *MDNSDiscoverer) Start() error {
	err := d.register()
	if err != nil {
		return err
	}
	go d.browse()
	return nil
}

func (d *MDNSDiscoverer) Stop() {
	close(d.stopCh)
	if d.server != nil {
		d.server.Shutdown()
	}
}

func (d *MDNSDiscoverer) UpdateStatus(status string) {
	d.mu.Lock()
	d.cfg.Status = status
	d.mu.Unlock()

	if d.server != nil {
		d.server.Shutdown()
	}
	err := d.register()
	if err != nil {
		log.Println("mDNS re-register failed:", err)
	}
}

func (d *MDNSDiscoverer) DeviceChan() <-chan config.DeviceEvent {
	return d.deviceCh
}

func (d *MDNSDiscoverer) AddGossipPeer(ip string) {}

// ── Internal ──────────────────────────────────────────────────

func (d *MDNSDiscoverer) register() error {
	d.mu.Lock()
	name := d.cfg.DisplayName
	port := d.cfg.Port
	status := d.cfg.Status
	d.localIP = getLocalIP()
	d.mu.Unlock()

	name = d.resolveNameCollision(name)

	txt := []string{
		"status=" + status,
	}

	// Register(instance, service, domain, port, text, ifaces)
	server, err := zeroconf.Register(
		name,
		serviceName,
		domain,
		port,
		txt,
		nil,
	)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.server = server
	d.mu.Unlock()

	log.Printf("mDNS: registered as %q on port %d", name, port)
	return nil
}

func (d *MDNSDiscoverer) browse() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-d.stopCh
		cancel()
	}()

	entries := make(chan *zeroconf.ServiceEntry)

	// NewResolver(options ...ClientOption) (*Resolver, error)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Println("mDNS resolver failed:", err)
		return
	}

	// resolver.Browse(ctx, service, domain, entries)
	err = resolver.Browse(ctx, serviceName, domain, entries)
	if err != nil {
		log.Println("mDNS browse failed:", err)
		return
	}

	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				return
			}
			d.processEntry(entry)

		case <-d.stopCh:
			return
		}
	}
}

func (d *MDNSDiscoverer) processEntry(entry *zeroconf.ServiceEntry) {
	// ignore our own announcements
	hostname, _ := os.Hostname()
	if entry.HostName == hostname+"." {
		return
	}

	// skip if no IPv4 address
	if len(entry.AddrIPv4) == 0 {
		return
	}
	ip := entry.AddrIPv4[0].String()

	// extract status from TXT records
	status := "online"
	for _, txt := range entry.Text {
		if len(txt) > 7 && txt[:7] == "status=" {
			status = txt[7:]
		}
	}

	device := config.Device{
		Name:   entry.Instance,
		IP:     ip,
		Status: status,
	}

	d.mu.Lock()
	_, exists := d.devices[ip]
	d.devices[ip] = device
	d.mu.Unlock()

	eventType := config.DeviceJoined
	if exists {
		eventType = config.DeviceUpdated
	}

	d.deviceCh <- config.DeviceEvent{
		Type:   eventType,
		Device: device,
	}

	log.Printf("mDNS: found %q at %s status=%s", device.Name, ip, status)
}

func (d *MDNSDiscoverer) resolveNameCollision(name string) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, device := range d.devices {
		if device.Name == name {
			hostname, err := os.Hostname()
			if err != nil {
				hostname = "device"
			}
			newName := name + "-" + hostname
			log.Printf("mDNS: name %q taken, using %q", name, newName)
			return newName
		}
	}
	return name
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err !=nil{
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func (d *MDNSDiscoverer) LocalIP() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.localIP
}
