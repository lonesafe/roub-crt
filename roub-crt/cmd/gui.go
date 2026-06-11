// +build gui

package cmd

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"

	"roub-crt/internal/config"
	"roub-crt/internal/connection"
	"roub-crt/internal/session"
	"roub-crt/internal/terminal"
)

type GUIApp struct {
	window fyne.Window
	app    fyne.App
	config *config.Config
	tabs   *widget.AppTabs
	terms  map[string]*terminalInfo
}

type terminalInfo struct {
	term     *terminal.Terminal
	conn     interface{}
	protocol string
}

func RunGUI() {
	a := app.New()
	gui := &GUIApp{
		app:   a,
		terms: make(map[string]*terminalInfo),
	}
	gui.Run()
}

func (g *GUIApp) Run() {
	g.window = g.app.NewWindow("roub-crt - Professional Terminal Emulator")
	g.window.Resize(fyne.NewSize(1200, 800))
	g.window.SetMaster()

	cfg, _ := config.LoadConfig("")
	if cfg == nil {
		cfg = &config.Config{}
	}
	g.config = cfg

	g.buildUI()

	g.window.Show()
	g.app.Run()
}

func (g *GUIApp) buildUI() {
	g.tabs = widget.NewAppTabs()

	g.tabs.Append(widget.NewTabItem("+", widget.NewLabel("")))

	header := g.createHeader()

	content := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(header, nil, nil, nil),
		header,
		g.tabs,
	)

	menu := g.createMainMenu()
	g.window.SetMainMenu(menu)
	g.window.SetContent(content)
}

func (g *GUIApp) createHeader() fyne.CanvasObject {
	logo := canvas.NewText("roub-crt", theme.PrimaryColor())
	logo.TextSize = 24
	logo.TextStyle = fyne.TextStyle{Bold: true}

	newConnBtn := widget.NewButton("New Connection", func() {
		g.showQuickConnect()
	})

	sessionBtn := widget.NewButton("Sessions", func() {
		g.showSessionManager()
	})

	transferBtn := widget.NewButton("File Transfer", func() {
		g.showFileTransfer(nil)
	})

	tunnelBtn := widget.NewButton("Tunnels", func() {
		g.showTunnelManager()
	})

	header := widget.NewHBox(
		layout.NewSpacer(),
		logo,
		layout.NewSpacer(),
		newConnBtn,
		sessionBtn,
		transferBtn,
		tunnelBtn,
	)

	return header
}

func (g *GUIApp) createMainMenu() *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New Connection", func() { g.showQuickConnect() }),
		fyne.NewMenuItem("Session Manager", func() { g.showSessionManager() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings", func() { g.showSettings() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() { g.app.Quit() }),
	)

	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Copy", func() {}),
		fyne.NewMenuItem("Paste", func() {}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Select All", func() {}),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Zoom In", func() {}),
		fyne.NewMenuItem("Zoom Out", func() {}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Full Screen", func() {
			g.window.ToggleFullScreen()
		}),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() { g.showAbout() }),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, viewMenu, helpMenu)
}

func (g *GUIApp) showQuickConnect() {
	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("Host or /dev/ttyUSB0")

	portEntry := widget.NewEntry()
	portEntry.SetText("22")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username (optional)")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password (optional)")

	protocolSelect := widget.NewSelect([]string{"SSH", "Telnet", "Serial", "Rlogin"}, func(value string) {})
	protocolSelect.SetSelected("SSH")

	baudSelect := widget.NewSelect([]string{"9600", "19200", "38400", "57600", "115200"}, func(value string) {})
	baudSelect.SetSelected("115200")
	baudSelect.Hide()

	protocolSelect.OnChanged = func(value string) {
		if value == "Serial" {
			baudSelect.Show()
			portEntry.Hide()
		} else {
			baudSelect.Hide()
			portEntry.Show()
		}
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Protocol", Widget: protocolSelect},
			{Text: "Host/Port", Widget: portEntry},
			{Text: "Username", Widget: usernameEntry},
			{Text: "Password", Widget: passwordEntry},
			{Text: "Baud Rate", Widget: baudSelect},
		},
	}

	dialog.NewForm("Quick Connect", "Connect", "Cancel", form, func(confirmed bool) {
		if !confirmed {
			return
		}

		host := hostEntry.Text
		if host == "" {
			dialog.ShowError(fmt.Errorf("host is required"), g.window)
			return
		}

		port, _ := strconv.Atoi(portEntry.Text)
		username := usernameEntry.Text
		password := passwordEntry.Text
		protocol := protocolSelect.Selected
		baudRate, _ := strconv.Atoi(baudSelect.Selected)

		if port == 0 {
			switch protocol {
			case "SSH":
				port = 22
			case "Telnet":
				port = 23
			case "Rlogin":
				port = 513
			}
		}

		g.connect(host, port, username, password, protocol, baudRate)
	}, g.window).Show()
}

func (g *GUIApp) connect(host string, port int, username, password, protocol string, baudRate int) {
	tabLabel := host
	if username != "" {
		tabLabel = fmt.Sprintf("%s@%s", username, host)
	}

	var conn interface{}
	var err error

	switch protocol {
	case "SSH":
		cfg := connection.NewSSHConfig(host, port, username, password)
		conn, err = connection.ConnectSSH(cfg)
	case "Telnet":
		cfg := connection.NewTelnetConfig(host, port)
		conn, err = connection.ConnectTelnet(cfg)
		if err == nil && username != "" {
			conn.(*connection.TelnetConn).Login(username, password)
		}
	case "Serial":
		cfg := connection.NewSerialConfig(host, baudRate)
		conn, err = connection.ConnectSerial(cfg)
	case "Rlogin":
		cfg := connection.NewRloginConfig(host, port, username)
		conn, err = connection.ConnectRlogin(cfg)
	default:
		dialog.ShowError(fmt.Errorf("unsupported protocol: %s", protocol), g.window)
		return
	}

	if err != nil {
		dialog.ShowError(fmt.Errorf("connection failed: %v", err), g.window)
		return
	}

	term := terminal.NewTerminal(80, 24)
	termID := fmt.Sprintf("%d", time.Now().UnixNano())
	g.terms[termID] = &terminalInfo{
		term:     term,
		conn:     conn,
		protocol: protocol,
	}

	content := g.createTerminalView(termID, term)
	tab := widget.NewTabItem(tabLabel, content)

	isFirstTab := len(g.tabs.Items) == 1 && g.tabs.Items[0].Text == "+"
	if isFirstTab {
		g.tabs.Items[0].Text = tabLabel
		g.tabs.Items[0].Content = content
		g.tabs.Items[0].Close = func() {
			g.closeTab(termID)
		}
		g.tabs.Refresh()
	} else {
		g.tabs.Append(tab)
	}
}

func (g *GUIApp) createTerminalView(termID string, term *terminal.Terminal) fyne.CanvasObject {
	label := widget.NewLabel("Terminal session - " + termID)
	label.Alignment = fyne.TextAlignCenter

	closeBtn := widget.NewButton("Close", func() {
		g.closeTab(termID)
	})

	toolbar := widget.NewHBox(
		layout.NewSpacer(),
		closeBtn,
	)

	content := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(toolbar, nil, nil, nil),
		toolbar,
		widget.NewScrollContainer(label),
	)

	return content
}

func (g *GUIApp) closeTab(termID string) {
	info, ok := g.terms[termID]
	if !ok {
		return
	}

	if sshConn, ok := info.conn.(*connection.SSHConn); ok {
		sshConn.Close()
	} else if telnetConn, ok := info.conn.(*connection.TelnetConn); ok {
		telnetConn.Close()
	} else if serialConn, ok := info.conn.(*connection.SerialConn); ok {
		serialConn.Close()
	} else if rloginConn, ok := info.conn.(*connection.RloginConn); ok {
		rloginConn.Close()
	}

	delete(g.terms, termID)

	for i, tab := range g.tabs.Items {
		if tab.Text != "+" {
			continue
		}
		if i == 0 && len(g.tabs.Items) == 1 {
			tab.Text = "+"
			tab.Content = widget.NewLabel("")
			g.tabs.Refresh()
		} else {
			g.tabs.RemoveIndex(i)
		}
		break
	}
}

func (g *GUIApp) showSessionManager() {
	sm, err := session.NewSessionManager("")
	if err != nil {
		dialog.ShowError(err, g.window)
		return
	}

	sessions := sm.ListSessions()

	var sessionItems []string
	for _, s := range sessions {
		sessionItems = append(sessionItems, fmt.Sprintf("%s - %s@%s:%d (%s)",
			s.Name, s.Username, s.Host, s.Port, s.Protocol))
	}

	if len(sessionItems) == 0 {
		dialog.ShowInformation("Sessions", "No saved sessions", g.window)
		return
	}

	list := widget.NewList(
		func() int { return len(sessionItems) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Session")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(sessions) {
				obj.(*widget.Label).SetText(sessionItems[id])
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id < len(sessions) {
			s := sessions[id]
			g.connect(s.Host, s.Port, s.Username, s.Password, string(s.Protocol), 115200)
		}
	}

	addBtn := widget.NewButton("Add", func() {
		g.showAddSession()
	})

	deleteBtn := widget.NewButton("Delete", func() {
		selected := list.SelectedIndex()
		if selected >= 0 && selected < len(sessions) {
			sm.DeleteSession(sessions[selected].ID)
			g.showSessionManager()
		}
	})

	importBtn := widget.NewButton("Import", func() {
		dialog.ShowInformation("Import", "Use 'roub-crt session import <file>' command", g.window)
	})

	exportBtn := widget.NewButton("Export", func() {
		dialog.ShowInformation("Export", "Use 'roub-crt session export <file>' command", g.window)
	})

	btnBar := widget.NewHBox(addBtn, deleteBtn, widget.NewLabel(""), importBtn, exportBtn)

	content := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(btnBar, nil, nil, nil),
		btnBar,
		list,
	)

	dialog.ShowCustom("Session Manager", "Close", content, g.window)
}

func (g *GUIApp) showAddSession() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("My Server")

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("192.168.1.100")

	portEntry := widget.NewEntry()
	portEntry.SetText("22")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("root")

	protocolSelect := widget.NewSelect([]string{"ssh", "telnet", "serial", "rlogin"}, func(value string) {})
	protocolSelect.SetSelected("ssh")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Protocol", Widget: protocolSelect},
			{Text: "Host", Widget: hostEntry},
			{Text: "Port", Widget: portEntry},
			{Text: "Username", Widget: usernameEntry},
		},
	}

	dialog.NewForm("Add Session", "Save", "Cancel", form, func(confirmed bool) {
		if !confirmed {
			return
		}

		sm, _ := session.NewSessionManager("")
		port, _ := strconv.Atoi(portEntry.Text)
		if port == 0 {
			port = 22
		}

		s := &session.Session{
			Name:     nameEntry.Text,
			Protocol: session.ProtocolType(protocolSelect.Selected),
			Host:     hostEntry.Text,
			Port:     port,
			Username: usernameEntry.Text,
		}

		sm.SaveSession(s)
		dialog.ShowInformation("Success", fmt.Sprintf("Session '%s' saved", s.Name), g.window)
	}, g.window).Show()
}

func (g *GUIApp) showFileTransfer(s *session.Session) {
	info := `File Transfer

Use SCP/SFTP to transfer files between local and remote systems.

Commands:
- roub-crt transfer --session <id>  Start file transfer
- F5: Upload selected file
- F6: Download selected file
- Tab: Switch between panels
- F10: Quit

Or use the CLI command:
  roub-crt transfer -s <session-id>`

	dialog.ShowInformation("File Transfer", info, g.window)
}

func (g *GUIApp) showTunnelManager() {
	info := `Port Tunnel Manager

SSH Tunnel Types:
- Local Forward (-L):   Forward local port to remote
- Remote Forward (-R):  Forward remote port to local
- Dynamic (-D):        SOCKS5 proxy

Example:
  roub-crt tunnel local 8080:localhost:80 user@host

Current active tunnels:
  (Use 'roub-crt tunnel list' to view)`

	dialog.ShowInformation("Port Tunnels", info, g.window)
}

func (g *GUIApp) showSettings() {
	settingsText := `Settings

Terminal:
  Font: Monospace
  Font Size: 14
  Scrollback: 10000 lines
  Cursor Shape: Block

Colorschemes:
  - Default (Dark)
  - Monokai
  - Solarized Dark
  - Dracula
  - Nord

Connection:
  Timeout: 30 seconds
  Keepalive: 10 seconds
  Default Encoding: UTF-8

Security:
  Encrypt Transfers: Yes
  Strict Host Key: Yes

Configure via config file:
  ~/.roub-crt/config.yaml`

	label := widget.NewLabel(settingsText)
	scroll := widget.NewScrollContainer(label)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	dialog.ShowCustom("Settings", "Close", scroll, g.window)
}

func (g *GUIApp) showAbout() {
	about := `roub-crt v1.0.0
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Professional Terminal Emulation & File Transfer

Supported Protocols:
  • SSH1 / SSH2 (Secure Shell)
  • Telnet
  • Serial (RS-232)
  • Rlogin
  • TAPI

Features:
  ✓ Multi-tab terminal sessions
  ✓ Session management & folders
  ✓ SFTP/SCP file transfer
  ✓ Local/Remote/Dynamic tunnels
  ✓ AES/Twofish encryption
  ✓ RSA/ECDSA key authentication
  ✓ Custom color schemes
  ✓ VT100/xterm emulation

Keyboard Shortcuts:
  Ctrl+C: Copy
  Ctrl+V: Paste
  Ctrl+L: Clear screen
  Ctrl+W: Close tab
  F11: Full screen

© 2026 roub-crt`

	dialog.ShowInformation("About roub-crt", about, g.window)
}

