package main

import (
	"fmt"
	"os"

	"github.com/Suryakiran46/lantern/internal/config"
	"github.com/Suryakiran46/lantern/internal/discovery"
	"github.com/Suryakiran46/lantern/internal/network"
	"github.com/Suryakiran46/lantern/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	disc := discovery.NewDiscoverer(cfg)
	if err := disc.Start(); err != nil {
		fmt.Printf("Error starting discovery: %v\n", err)
		os.Exit(1)
	}
	defer disc.Stop()

	serv := network.NewServer(cfg.Port)
	if err := serv.Listen(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	defer serv.Close()

	model := tui.NewModel(cfg, disc.DeviceChan(), serv.MessageChan())

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting TUI: %v\n", err)
		os.Exit(1)
	}
}
