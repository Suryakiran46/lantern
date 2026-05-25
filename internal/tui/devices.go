package tui

import (
	"fmt"
	"strings"

	"github.com/Suryakiran46/lantern/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//Styles 
//Keep all colour/style decisions here for now. styles.go will centralise everything later,, for Phase 1 this is fine.

var (
	stylePrimary   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // bright blue
	styleSubtle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // grey
	styleSelected  = lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true)
	styleBadgeOK   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))  // green  — Known
	styleBadgeWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // yellow — New / KEY CHANGED
	styleBadgeAway = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // grey   — Away
	styleHeader    = lipgloss.NewStyle().Bold(true).Underline(true)
)

//DeviceListModel is the sub-model for the device list screen.
// It is owned by the top-level Model in model.go.
type DeviceListModel struct {
	myName  string
	devices []config.Device
	cursor  int
	width   int
	height  int

	// selectedForChat is set to true when the user presses Enter.
	// The parent model reads this flag and switches to the chat screen.
	selectedForChat bool
	selectedDevice  config.Device
}

// NewDeviceListModel creates the device list sub-model.
// It pre-populates with dummy devices so the TUI works before
// mDNS layer is wired in. On Day 7 swap to real DeviceEvents.
func NewDeviceListModel(myName string) DeviceListModel {
	return DeviceListModel{
		myName:  myName,
		devices: dummyDevices(), // <== two lines change later (i hope)
	}
}

// dummyDevices returns hardcoded test date
// Replace with real channel data later (again i hope T-T)
func dummyDevices() []config.Device {
	return []config.Device{
		{Name: "Suryaaaa", IP: "192.168.41.122", Status: "Active", Known: true},
		{Name: "nsha256", IP: "192.168.41.123", Status: "Away", Known: false},
		{Name: "Soumya Chechi", IP: "192.168.41.124", Status: "Active", Known: true},
	}
}

// SetSize is called by the parent model when the terminal is resized.
func (m *DeviceListModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// HandleDeviceEvent applies a real DeviceEvent from the discovery layer.
// Called by the parent model when a deviceEventMsg arrives.
func (m *DeviceListModel) HandleDeviceEvent(ev config.DeviceEvent) {
	switch ev.Type {
	case config.DeviceJoined, config.DeviceUpdated:
		// Update in place if already present, otherwise append
		for i, d := range m.devices {
			if d.IP == ev.Device.IP {
				m.devices[i] = ev.Device
				return
			}
		}
		m.devices = append(m.devices, ev.Device)

	case config.DeviceLeft:
		for i, d := range m.devices {
			if d.IP == ev.Device.IP {
				// Mark inactive rather than delete,, keeps the list stable
				m.devices[i].Status = "Inactive"
				return
			}
		}
	}
}

// SelectedForChat returns true if the user pressed "Enter" on a device.
func (m DeviceListModel) SelectedForChat() bool {
	return m.selectedForChat
}

// SelectedDevice returns the device the user chose to chat with.
func (m DeviceListModel) SelectedDevice() config.Device {
	return m.selectedDevice
}

// ClearSelection resets the flag after the parent has handled it.
func (m *DeviceListModel) ClearSelection() {
	m.selectedForChat = false
}

// Update handles keyboard input for the device list screen.
// Returns (updated model, optional command).
func (m DeviceListModel) Update(msg tea.Msg) (DeviceListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.devices) > 0 {
				m.selectedDevice = m.devices[m.cursor]
				m.selectedForChat = true
			}
		}
	}
	return m, nil
}

// View renders the device list screen as a string.
// This is a pure function — same model in, same string out.
func (m DeviceListModel) View() string {
	var sb strings.Builder

	//Header
	sb.WriteString(stylePrimary.Render("Lantern"))
	sb.WriteString("  ")
	sb.WriteString(styleSubtle.Render(fmt.Sprintf("Hello, your username is: %s", m.myName)))
	sb.WriteString("\n\n")

	//Column headers
	sb.WriteString(styleHeader.Render(
		fmt.Sprintf("  %-20s %-16s %-10s %s", "Name", "IP", "Status", "Trust"),
	))
	sb.WriteString("\n")
	sb.WriteString(styleSubtle.Render(strings.Repeat("─", 60)))
	sb.WriteString("\n")

	//Device rows
	if len(m.devices) == 0 {
		sb.WriteString(styleSubtle.Render("  No devices found on LAN yet..."))
		sb.WriteString("\n")
	}

	for i, d := range m.devices {
		//Trust badge
		trust := trustBadge(d)

		//Status indicator
		statusStr := statusLabel(d.Status)

		row := fmt.Sprintf("  %-20s %-16s %-10s %s", d.Name, d.IP, statusStr, trust)

		if i == m.cursor {
			sb.WriteString(styleSelected.Render(row))
		} else {
			sb.WriteString(row)
		}
		sb.WriteString("\n")
	}

	//Footer/keybinds
	sb.WriteString("\n")
	sb.WriteString(styleSubtle.Render("up/down to navigate,   Enter open chat,   q to quit"))

	return sb.String()
}

//Helpers

func trustBadge(d config.Device) string {
	if !d.Known {
		return styleBadgeWarn.Render("New")
	}
	// PublicKey mismatch check stuff
	return styleBadgeOK.Render("Known")
}

func statusLabel(status string) string {
	switch status {
	case "Away", "Inactive":
		return styleBadgeAway.Render(status)
	default:
		return status
	}
}
