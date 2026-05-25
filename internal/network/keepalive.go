package network

import (
	"time"

	"github.com/Suryakiran46/lantern/internal/config"
)

func StartKeepalive(send func(config.Message) error, pongChan <-chan struct{}, onDead func()) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		missed := 0
		for {
			select {
			case <-ticker.C:
				err := send(config.Message{Type: "ping"})
				if err != nil {
					onDead()
					return
				}
				missed++
				if missed >= 3 {
					onDead()
					return
				}
			case <-pongChan:
				missed = 0
			}
		}
	}()
}
