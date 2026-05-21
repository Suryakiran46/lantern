package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type AppConfig struct {
	DisplayName string `json:"display_name"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
}

type Device struct {
	Name      string
	IP        string
	Status    string
	PublicKey string
	Known     bool
}

type DeviceEventType int

const (
	DeviceJoined DeviceEventType = iota
	DeviceLeft
	DeviceUpdated
)

type DeviceEvent struct {
	Type   DeviceEventType
	Device Device
}

type Discoverer interface {
	Start() error
	Stop()
	UpdateStatus(status string)
	DeviceChan() <-chan DeviceEvent
	AddGossipPeer(ip string)
}

type Message struct {
	Type string `json:"type"`
	From string `json:"from,omitempty"`
	Text string `json:"text,omitempty"`
	Room string `json:"room,omitempty"`
	Ts   int64  `json:"ts,omitempty"`
}

func ConfigDir() string {
	homedir, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			appdata = homedir
		}
		return filepath.Join(appdata, "lantern")
	default:
		return filepath.Join(homedir, ".config", "lantern")
	}
}

func DataDir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			local = home
		}
		return filepath.Join(local, "lantern", "history")
	default:
		return filepath.Join(home, ".local", "share", "lantern", "history")
	}
}

func Save(cfg AppConfig) error {
	appconfigbytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := ConfigDir()
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "config.json"), appconfigbytes, 0644)
}

func Load() (AppConfig, error) {
	var cfg AppConfig
	path := filepath.Join(ConfigDir(), "config.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		fmt.Println("Enter your display name: ")
		var name string
		_, err = fmt.Scan(&name)
		cfg = AppConfig{
			DisplayName: name,
			Port:        5000,
			Status:      "Active",
		}
		err = Save(cfg)
		if err != nil {
			return AppConfig{}, err
		}
		return cfg, nil
	} else if err != nil {
		return AppConfig{}, err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return AppConfig{}, err
	}

	return cfg, nil

}
