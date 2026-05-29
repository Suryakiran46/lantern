package main

import (
	"log"
	"fmt"
	"github.com/Suryakiran46/lantern/internal/config"
	"github.com/Suryakiran46/lantern/internal/discovery"
	"github.com/Suryakiran46/lantern/internal/network"
	"github.com/Suryakiran46/lantern/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err:= run(); err !=nil{
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w",err)
	}

	disc := discovery.NewDiscoverer(cfg)
	if err := disc.Start(); err != nil {
		return fmt.Errorf("starting discovery: %w", err)
	}
	defer disc.Stop()

	localIP := disc.LocalIP()

	serv := network.NewServer(cfg.Port)
	if err := serv.Listen(); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}
	defer serv.Close()

	model := tui.NewModel(cfg, disc.DeviceChan(), serv.MessageChan(),serv, localIP)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w\n", err)
	}
	return nil
}

