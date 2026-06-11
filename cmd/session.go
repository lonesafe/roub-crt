package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/AlecAivazis/survey/v2"
	"roub-crt/internal/session"
)

var sessionManager *session.SessionManager

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionAddCmd)
	sessionCmd.AddCommand(sessionEditCmd)
	sessionCmd.AddCommand(sessionDelCmd)
	sessionCmd.AddCommand(sessionImportCmd)
	sessionCmd.AddCommand(sessionExportCmd)

	var err error
	sessionManager, err = session.NewSessionManager("")
	if err != nil {
		fmt.Printf("Warning: failed to initialize session manager: %v\n", err)
	}
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage saved sessions",
	Long:  `Manage saved connection sessions including listing, adding, editing, and deleting sessions.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved sessions",
	Run: func(cmd *cobra.Command, args []string) {
		if sessionManager == nil {
			fmt.Println("Session manager not initialized")
			return
		}

		sessions := sessionManager.ListSessions()

		if len(sessions) == 0 {
			fmt.Println("No saved sessions")
			return
		}

		fmt.Printf("\n%-6s %-20s %-15s %-30s %-10s\n",
			"ID", "Name", "Protocol", "Host", "Folder")
		fmt.Println("--------------------------------------------------------------------")

		for _, s := range sessions {
			host := s.Host
			if s.Host == "" {
				host = "(not set)"
			}
			port := fmt.Sprintf("%d", s.Port)
			if s.Port == 0 {
				port = "-"
			}
			folder := s.Folder
			if folder == "" {
				folder = "/"
			}

			fmt.Printf("%-6s %-20s %-15s %s:%-10s %-10s\n",
				truncate(s.ID, 6),
				truncate(s.Name, 20),
				s.Protocol,
				truncate(host, 20),
				port,
				truncate(folder, 10))
		}
		fmt.Println()
	},
}

var sessionAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new session",
	Run: func(cmd *cobra.Command, args []string) {
		newSession := session.GetDefaultSessionConfig()

		questions := []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Session name:",
				},
				Validate: func(val interface{}) error {
					if str, ok := val.(string); !ok || str == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				},
			},
			{
				Name: "protocol",
				Prompt: &survey.Select{
					Message: "Connection protocol:",
					Options: []string{"ssh", "telnet", "serial", "rlogin"},
					Default: "ssh",
				},
			},
			{
				Name: "host",
				Prompt: &survey.Input{
					Message: "Host address:",
				},
			},
			{
				Name: "port",
				Prompt: &survey.Input{
					Message: "Port (0 for default):",
					Default: "0",
				},
			},
			{
				Name: "username",
				Prompt: &survey.Input{
					Message: "Username:",
				},
			},
		}

		answers := struct {
			Name     string
			Protocol string
			Host     string
			Port     string
			Username string
		}{}

		if err := survey.Ask(questions, &answers); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		newSession.Name = answers.Name
		newSession.Protocol = session.ProtocolType(answers.Protocol)
		newSession.Host = answers.Host
		newSession.Username = answers.Username

		fmt.Sscanf(answers.Port, "%d", &newSession.Port)

		if newSession.Port == 0 {
			switch newSession.Protocol {
			case session.ProtocolSSH:
				newSession.Port = 22
			case session.ProtocolTelnet:
				newSession.Port = 23
			case session.ProtocolRlogin:
				newSession.Port = 513
			}
		}

		if err := sessionManager.SaveSession(newSession); err != nil {
			fmt.Printf("Error saving session: %v\n", err)
			return
		}

		fmt.Printf("Session '%s' saved successfully!\n", newSession.Name)
	},
}

var sessionEditCmd = &cobra.Command{
	Use:   "edit [session-id]",
	Short: "Edit a saved session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		s, ok := sessionManager.GetSession(id)
		if !ok {
			fmt.Printf("Session not found: %s\n", id)
			return
		}

		fmt.Printf("Editing session: %s\n", s.Name)

		questions := []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Session name:",
					Default: s.Name,
				},
			},
			{
				Name: "host",
				Prompt: &survey.Input{
					Message: "Host address:",
					Default: s.Host,
				},
			},
			{
				Name: "port",
				Prompt: &survey.Input{
					Message: "Port:",
					Default: fmt.Sprintf("%d", s.Port),
				},
			},
			{
				Name: "username",
				Prompt: &survey.Input{
					Message: "Username:",
					Default: s.Username,
				},
			},
		}

		answers := struct {
			Name     string
			Host     string
			Port     string
			Username string
		}{}

		if err := survey.Ask(questions, &answers); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		s.Name = answers.Name
		s.Host = answers.Host
		s.Username = answers.Username
		fmt.Sscanf(answers.Port, "%d", &s.Port)

		if err := sessionManager.SaveSession(s); err != nil {
			fmt.Printf("Error saving session: %v\n", err)
			return
		}

		fmt.Printf("Session '%s' updated successfully!\n", s.Name)
	},
}

var sessionDelCmd = &cobra.Command{
	Use:   "del [session-id]",
	Short: "Delete a saved session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		s, ok := sessionManager.GetSession(id)
		if !ok {
			fmt.Printf("Session not found: %s\n", id)
			return
		}

		var confirm bool
		survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Delete session '%s'?", s.Name),
			Default: false,
		}, &confirm)

		if !confirm {
			fmt.Println("Deletion cancelled")
			return
		}

		if err := sessionManager.DeleteSession(id); err != nil {
			fmt.Printf("Error deleting session: %v\n", err)
			return
		}

		fmt.Printf("Session '%s' deleted successfully!\n", s.Name)
	},
}

var sessionImportCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "Import sessions from a JSON file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		if err := sessionManager.ImportSessions(path); err != nil {
			fmt.Printf("Error importing sessions: %v\n", err)
			return
		}

		fmt.Println("Sessions imported successfully!")
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export sessions to a JSON file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		sessions := sessionManager.ListSessions()
		if len(sessions) == 0 {
			fmt.Println("No sessions to export")
			return
		}

		ids := make([]string, len(sessions))
		for i, s := range sessions {
			ids[i] = s.ID
		}

		if err := sessionManager.ExportSessions(ids, path); err != nil {
			fmt.Printf("Error exporting sessions: %v\n", err)
			return
		}

		fmt.Printf("Exported %d sessions to %s\n", len(sessions), path)
	},
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 0 {
		return ""
	}
	return s[:maxLen-1] + "…"
}
