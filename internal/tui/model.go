package tui

import (
	"github.com/Suryakiran46/lantern/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenDevices screen = iota
	screenChat
	screenRooms
	screenHistory
)

type Model struct {
	cfg           config.AppConfig
	currentScreen screen

	deviceList DeviceListModel
	chat       ChatModel

	deviceChan <-chan config.DeviceEvent
	msgChan    <-chan config.Message
	messenger config.Messenger

	width  int
	height int
}

func NewModel(cfg config.AppConfig, deviceChan <-chan config.DeviceEvent, msgChan <-chan config.Message, messenger config.Messenger,localIP string) Model {
	return Model{
		cfg:           cfg,
		currentScreen: screenDevices,
		deviceList:    NewDeviceListModel(cfg.DisplayName, localIP),
		deviceChan:    deviceChan,
		msgChan:       msgChan,
		messenger:     messenger,
	}

}

type deviceEventMsg config.DeviceEvent
type incomingMsgMsg config.Message

func listenForDevices(ch <-chan config.DeviceEvent) tea.Cmd {
	return func() tea.Msg {
		return deviceEventMsg(<-ch)
	}
}

func listenForMessages(ch <-chan config.Message) tea.Cmd {
	return func() tea.Msg {
		return incomingMsgMsg(<-ch)
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.deviceChan != nil {
		cmds = append(cmds, listenForDevices(m.deviceChan))
	}
	if m.msgChan != nil {
		cmds = append(cmds, listenForMessages(m.msgChan))
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.deviceList.SetSize(msg.Width, msg.Height)
		m.chat.SetSize(msg.Width, msg.Height)
		return m, nil

	case deviceEventMsg:
		m.deviceList.HandleDeviceEvent(config.DeviceEvent(msg))
		if m.deviceChan != nil {
			return m, listenForDevices(m.deviceChan)
		}
		return m, nil

	case incomingMsgMsg:
		if m.currentScreen == screenChat {
			m.chat.HandleIncomingMessage(config.Message(msg))
		}
		if m.msgChan != nil {
			return m, listenForMessages(m.msgChan)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.currentScreen == screenDevices {
				return m, tea.Quit
			}
		}
		
		switch m.currentScreen {

		case screenDevices:
			var cmd tea.Cmd
			m.deviceList, cmd = m.deviceList.Update(msg)
			if m.deviceList.SelectedForChat() {
				peer := m.deviceList.SelectedDevice()
				m.deviceList.ClearSelection()
				if(m.messenger!=nil){
					m.messenger.Connect(peer.IP)
				}
				m.chat = NewChatModel(m.cfg.DisplayName, peer,m.messenger)
				m.chat.SetSize(m.width, m.height)
				m.currentScreen = screenChat
			}
			return m, cmd

		case screenChat:
			var cmd tea.Cmd
			m.chat, cmd = m.chat.Update(msg)
			if m.chat.BackToDevices() {
				m.chat.ClearBack()
				m.currentScreen = screenDevices
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.currentScreen {
	case screenDevices:
		return m.deviceList.View()
	case screenChat:
		return m.chat.View()
	default:
		return "Coming soon...\n\nPress q to quit."
	}
}
