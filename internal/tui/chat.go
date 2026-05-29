package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Suryakiran46/lantern/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//Styles

var (
	styleChatHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleMsgYou      = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	styleMsgPeer     = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	styleMsgTime     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleMsgText     = lipgloss.NewStyle()
	styleTyping      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	styleInputBorder = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(lipgloss.Color("240"))
	styleChatSubtle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// chatMessage is an internal display record, separate from config.Message,, so we can store already formatted timestamp strings
type chatMessage struct {
	from string
	text string
	ts   string
	isMe bool
}

func formatTS(t time.Time) string {
	return t.Format("15:04")
}

// ChatModel is the sub-model for the chat screen, owned by the top level Model in model.go
type ChatModel struct {
	peer     config.Device //who we're chatting with rn
	myName   string
	messages []chatMessage

	viewport  viewport.Model
	input     textinput.Model
	vpReady   bool //viewport needs an initial size before it can render

	//typingPeer is set when a typing indicator message arrives
	//empty string means nobody is typing
	typingPeer string

	//backToDevices signals the parent to switch back to the device list
	backToDevices bool

	messenger config.Messenger
	peerIP    string
}

//NewChatModel creates a chat sub model for the given peer
func NewChatModel(myName string, peer config.Device, messenger config.Messenger) ChatModel {
	//input bar setup
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("Message %s...", peer.Name)
	ti.Focus()
	ti.CharLimit = 500

	m := ChatModel{
		peer:     peer,
		myName:   myName,
		messages: dummyMessages(myName, peer.Name),
		input:    ti,
		messenger: messenger,
		peerIP: peer.IP,
	}
	return m
}

//dummyMessages returns fake conversation history for Phase 1 testing
//Delete this and replace with real history.Load() in Phase 2
func dummyMessages(myName, peerName string) []chatMessage {
	now := time.Now()
	return []chatMessage{
		{from: peerName, text: "msg 1", ts: formatTS(now.Add(-5 * time.Minute)), isMe: false},
		{from: myName, text: "msg 2", ts: formatTS(now.Add(-4 * time.Minute)), isMe: true},
		{from: peerName, text: "msg 3", ts: formatTS(now.Add(-3 * time.Minute)), isMe: false},
		{from: myName, text: "msg 4", ts: formatTS(now.Add(-2 * time.Minute)), isMe: true},
		{from: peerName, text: "msg 5", ts: formatTS(now.Add(-1 * time.Minute)), isMe: false},
	}
}

//SetSize is called by the parent model on terminal resize
//The viewport must be sized before it can render this also sets vpReady
func (m *ChatModel) SetSize(w, h int) {
	headerH := 3  // header + separator
	footerH := 3  // input bar + separator
	vpH := h - headerH - footerH
	if vpH < 1 {
		vpH = 1
	}

	if !m.vpReady {
		m.viewport = viewport.New(w, vpH)
		m.viewport.SetContent(m.renderMessages())
		m.vpReady = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = vpH
	}
	m.input.Width = w - 4
}

//HandleIncomingMessage adds a real message from the network layer
//Called by the parent model when a config.Message arrives on msgChan
func (m *ChatModel) HandleIncomingMessage(msg config.Message) {
	m.typingPeer = "" //clear typing indicator on actual message
	m.messages = append(m.messages, chatMessage{
		from: msg.From,
		text: msg.Text,
		ts:   formatTS(time.Unix(msg.Ts, 0)),
		isMe: false,
	})
	m.viewport.SetContent(m.renderMessages())
	m.viewport.GotoBottom()
}

//HandleTyping sets the typing indicator for the peer.
func (m *ChatModel) HandleTyping(from string) {
	m.typingPeer = from
}

//BackToDevices returns true if the user pressed Esc to go back
func (m ChatModel) BackToDevices() bool {
	return m.backToDevices
}

//clearBack resets the flag after the parent has handled it
func (m *ChatModel) ClearBack() {
	m.backToDevices = false
}

//Update

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "esc":
			//go back to device list
			m.backToDevices = true
			return m, nil

		case "enter":
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}

			//add to local display immediately
			m.messages = append(m.messages, chatMessage{
				from: m.myName,
				text: text,
				ts:   formatTS(time.Now()),
				isMe: true,
			})
			m.input.SetValue("")
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			if m.messenger != nil {
				m.messenger.Send(m.peerIP, config.Message{
					Type: "chat",
					From: m.myName,
					Text: text,
					Ts:   time.Now().Unix(),
				})
			}

		default:
			//all other keys go to the input bar
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		//pass non-key messages (e.g., blink tick) to input and viewport
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

//View

func (m ChatModel) View() string {
	if !m.vpReady {
		return "Loading you mesages...\n"
	}

	var sb strings.Builder

	//Header
	trustStr := trustBadge(m.peer)
	sb.WriteString(styleChatHeader.Render(fmt.Sprintf("Chat with %s", m.peer.Name)))
	sb.WriteString("  ")
	sb.WriteString(trustStr)
	sb.WriteString("  ")
	sb.WriteString(styleChatSubtle.Render(m.peer.IP))
	sb.WriteString("\n")
	sb.WriteString(styleChatSubtle.Render(strings.Repeat("─", 60)))
	sb.WriteString("\n")

	//Message viewport
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	//Typing indicator sits between viewport and input
	if m.typingPeer != "" {
		sb.WriteString(styleTyping.Render(fmt.Sprintf("%s is typing...", m.typingPeer)))
	} else {
		sb.WriteString(" ") //keeps layout stable when nobody is typing
	}
	sb.WriteString("\n")

	//Input bar
	sb.WriteString(styleInputBorder.Render(m.input.View()))

	//Footer hints
	sb.WriteString("\n")
	sb.WriteString(styleChatSubtle.Render("enter =  send,   esc = back to devices,   Ctrl + C = quit"))

	return sb.String()
}

//renderMessages builds the full message log as a single string for the viewport
//Called whenever messages change or the viewport is resized
func (m ChatModel) renderMessages() string {
	if len(m.messages) == 0 {
		return styleChatSubtle.Render("  No messages yet. Say hello!")
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		var nameStyle lipgloss.Style
		if msg.isMe {
			nameStyle = styleMsgYou
		} else {
			nameStyle = styleMsgPeer
		}

		sb.WriteString(fmt.Sprintf("%s %s  %s\n",
			nameStyle.Render(msg.from),
			styleMsgTime.Render(msg.ts),
			styleMsgText.Render(msg.text),
		))
	}
	return sb.String()
}
