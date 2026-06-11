package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/charmbracelet/lipgloss"
	"roub-crt/internal/config"
	"roub-crt/internal/session"
	"roub-crt/pkg/ui"
)

var (
	interactiveMode   bool
	selectedSession   int
	showHiddenFiles   bool
)

func init() {
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(configCmd)
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive mode",
	Long:  `Start the interactive TUI with menu navigation, session management, and file transfer.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractive()
	},
}

func runInteractive() {
	oldState, err := terminalMakeRaw()
	if err != nil {
		fmt.Println("Failed to initialize terminal:", err)
		return
	}
	defer restoreTerminal(oldState)

	ui.ClearScreen()
	ui.HideCursor()
	defer ui.ShowCursor()

	cfg, err := config.LoadConfig("")
	if err != nil {
		cfg = &config.Config{}
	}

	uiInst := ui.NewTerminalUI(80, 24)
	uiInst.ColorScheme = getColorSchemeFromConfig(cfg)

	sessionMgr, _ := session.NewSessionManager("")

	mainMenu := ui.NewMenu("roub-crt v1.0.0 - Main Menu", []ui.MenuItem{
		{Label: "Quick Connect", Description: "Connect to a host without saving", Key: "1"},
		{Label: "Session Manager", Description: "Manage saved sessions", Key: "2"},
		{Label: "File Transfer", Description: "Transfer files between local and remote", Key: "3"},
		{Label: "Port Tunnel", Description: "Manage SSH tunnels", Key: "4"},
		{Label: "Key Manager", Description: "Generate and manage SSH keys", Key: "5"},
		{Label: "Settings", Description: "Configure application settings", Key: "6"},
		{Label: "Exit", Description: "Exit roub-crt", Key: "0"},
	})

	selected := 0
	menuHeight := 12

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	menuOffset := 2

	for {
		ui.MoveCursor(0, 0)

		header := lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Blue).
			Bold(true).
			Render("╔════════════════════════════════════════════════════════════╗\n")
		header += lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Blue).
			Bold(true).
			Render("║") + lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Foreground).
			Bold(true).
			Render("                roub-crt v1.0.0                             ") +
			lipgloss.NewStyle().
				Foreground(uiInst.ColorScheme.Blue).
				Bold(true).
				Render("║\n")
		header += lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Blue).
			Bold(true).
			Render("║") + lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Foreground).
			Render("        Professional Terminal Emulation & File Transfer     ") +
			lipgloss.NewStyle().
				Foreground(uiInst.ColorScheme.Blue).
				Bold(true).
				Render("║\n")
		header += lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Blue).
			Bold(true).
			Render("╠════════════════════════════════════════════════════════════╣\n")

		fmt.Print(header)

		menuContent := mainMenu.Render(selected)

		menuLines := strings.Split(menuContent, "\n")
		for i, line := range menuLines {
			if i >= menuHeight {
				break
			}
			prefix := "║  "
			if i == 0 {
				prefix = "╠══ "
			}
			if i == menuHeight-1 {
				prefix = "╠══ "
			}

			padding := 58 - len(line)
			if padding < 0 {
				padding = 0
			}

			fmt.Printf("%s%s%s║\n", prefix, line, strings.Repeat(" ", padding))
		}

		footer := lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.Blue).
			Bold(true).
			Render("╚════════════════════════════════════════════════════════════╝\n")
		footer += lipgloss.NewStyle().
			Foreground(uiInst.ColorScheme.BrightBlack).
			Render("  [↑↓] Navigate  [Enter] Select  [q] Quit\n")
		fmt.Print(footer)

		ui.MoveCursor(0, menuOffset+selected+4)

		key := readKey()

		switch key {
		case "q", "Q", "0":
			ui.ClearScreen()
			fmt.Println("Goodbye!")
			return
		case "up":
			if selected > 0 {
				selected--
			}
		case "down":
			if selected < len(mainMenu.Items)-1 {
				selected++
			}
		case "enter", "":
			switch selected {
			case 0:
				showQuickConnect(sessionMgr, uiInst)
			case 1:
				showSessionManager(sessionMgr, uiInst)
			case 6:
				ui.ClearScreen()
				fmt.Println("Goodbye!")
				return
			}
		}
	}
}

func showQuickConnect(sm *session.SessionManager, uiInst *ui.TerminalUI) {
	ui.ClearScreen()
	ui.MoveCursor(0, 0)

	header := lipgloss.NewStyle().
		Foreground(uiInst.ColorScheme.Blue).
		Bold(true).
		Render("Quick Connect\n\n")
	fmt.Print(header)

	fmt.Println("Enter host: ")
	fmt.Scanln()

	fmt.Println("\nProtocol: [1] SSH  [2] Telnet  [3] Serial  [4] Rlogin")

	fmt.Println("\nPress any key to return...")
	readKey()
}

func showSessionManager(sm *session.SessionManager, uiInst *ui.TerminalUI) {
	ui.ClearScreen()
	ui.MoveCursor(0, 0)

	header := lipgloss.NewStyle().
		Foreground(uiInst.ColorScheme.Blue).
		Bold(true).
		Render("Session Manager\n\n")
	fmt.Print(header)

	if sm == nil {
		fmt.Println("Session manager not available")
		readKey()
		return
	}

	sessions := sm.ListSessions()

	if len(sessions) == 0 {
		fmt.Println("No saved sessions")
		fmt.Println("\nPress any key to return...")
		readKey()
		return
	}

	fmt.Printf("%-6s %-20s %-15s %-20s\n", "ID", "Name", "Protocol", "Host")
	fmt.Println("------------------------------------------------------------")

	for i, s := range sessions {
		host := s.Host
		if host == "" {
			host = "(not set)"
		}

		prefix := "  "
		if i == selectedSession {
			prefix = "> "
		}

		fmt.Printf("%s%-6s %-20s %-15s %-20s\n",
			prefix,
			truncate(s.ID, 6),
			truncate(s.Name, 20),
			s.Protocol,
			truncate(host, 20))
	}

	fmt.Println("\n[Enter] Connect  [d] Delete  [q] Back")

	key := readKey()
	switch key {
	case "q", "Q":
		return
	case "d", "D":
		if len(sessions) > 0 && selectedSession < len(sessions) {
			sm.DeleteSession(sessions[selectedSession].ID)
			if selectedSession > 0 {
				selectedSession--
			}
		}
	}
}

func getColorSchemeFromConfig(cfg *config.Config) ui.ColorScheme {
	if cfg == nil {
		return ui.GetDefaultColorScheme()
	}

	return ui.ColorScheme{
		Background:      lipgloss.Color("#1E1E1E"),
		Foreground:      lipgloss.Color("#D4D4D4"),
		Cursor:          lipgloss.Color("#FFFFFF"),
		Selection:       lipgloss.Color("#264F78"),
		Black:           lipgloss.Color("#000000"),
		Red:             lipgloss.Color("#CD3131"),
		Green:           lipgloss.Color("#0DBC79"),
		Yellow:          lipgloss.Color("#E5E510"),
		Blue:            lipgloss.Color("#2472C8"),
		Magenta:         lipgloss.Color("#BC3FBC"),
		Cyan:            lipgloss.Color("#11A8CD"),
		White:           lipgloss.Color("#E5E5E5"),
		BrightBlack:     lipgloss.Color("#666666"),
		BrightRed:       lipgloss.Color("#F14C4C"),
		BrightGreen:     lipgloss.Color("#23D18B"),
		BrightYellow:    lipgloss.Color("#F5F543"),
		BrightBlue:      lipgloss.Color("#3B8EEA"),
		BrightMagenta:   lipgloss.Color("#D670D6"),
		BrightCyan:      lipgloss.Color("#29B8DB"),
		BrightWhite:     lipgloss.Color("#FFFFFF"),
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and modify roub-crt configuration settings.`,
}

func terminalMakeRaw() (interface{}, error) {
	return nil, nil
}

func restoreTerminal(state interface{}) {}

func readKey() string {
	return ""
}
