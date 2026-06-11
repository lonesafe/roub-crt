package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"roub-crt/internal/tunnel"
)

var tunnelType string
var tunnelLocalPort int
var tunnelRemoteHost string
var tunnelRemotePort int
var tunnelSocksPort int

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.AddCommand(tunnelLocalCmd)
	tunnelCmd.AddCommand(tunnelRemoteCmd)
	tunnelCmd.AddCommand(tunnelDynamicCmd)
	tunnelCmd.AddCommand(tunnelListCmd)
	tunnelCmd.AddCommand(tunnelKillCmd)
}

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Port forwarding and tunneling",
	Long:  `Manage SSH tunnels for port forwarding. Supports local, remote, and dynamic (SOCKS5) tunnels.`,
}

var tunnelLocalCmd = &cobra.Command{
	Use:   "local [ssh-args]",
	Short: "Create a local port forward (ssh -L)",
	Long:  `Create a local port forward that redirects traffic from a local port to a remote host:port.

Example: roub-crt tunnel local 8080:localhost:80 user@host`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Local tunnel functionality requires SSH connection")
		fmt.Println("Use 'roub-crt interactive' for full tunnel management")
		return nil
	},
}

var tunnelRemoteCmd = &cobra.Command{
	Use:   "remote [ssh-args]",
	Short: "Create a remote port forward (ssh -R)",
	Long:  `Create a remote port forward that redirects traffic from a remote port to a local host:port.

Example: roub-crt tunnel remote 8080:localhost:80 user@host`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Remote tunnel functionality requires SSH connection")
		fmt.Println("Use 'roub-crt interactive' for full tunnel management")
		return nil
	},
}

var tunnelDynamicCmd = &cobra.Command{
	Use:   "dynamic [ssh-args]",
	Short: "Create a dynamic SOCKS5 proxy (ssh -D)",
	Long:  `Create a dynamic port forward that acts as a SOCKS5 proxy.

Example: roub-crt tunnel dynamic 1080 user@host`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Dynamic tunnel functionality requires SSH connection")
		fmt.Println("Use 'roub-crt interactive' for full tunnel management")
		return nil
	},
}

var tunnelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active tunnels",
	Run: func(cmd *cobra.Command, args []string) {
		tm := tunnel.NewTunnelManager()
		tunnels := tm.ListTunnels()

		if len(tunnels) == 0 {
			fmt.Println("No active tunnels")
			return
		}

		fmt.Printf("\n%-20s %-10s %-25s %-15s\n",
			"Name", "Type", "Local Address", "Status")
		fmt.Println("------------------------------------------------------------")

		for _, t := range tunnels {
			tinfo := t.String()
			fmt.Printf("%-20s %-10s %-25s %-15s\n",
				truncate(tinfo, 20),
				t.Type,
				fmt.Sprintf("%s:%d", t.ListenAddr, t.ListenPort),
				"active")
		}
		fmt.Println()
	},
}

var tunnelKillCmd = &cobra.Command{
	Use:   "kill [name]",
	Short: "Kill an active tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tm := tunnel.NewTunnelManager()

		if err := tm.CloseTunnel(name); err != nil {
			return fmt.Errorf("failed to close tunnel: %w", err)
		}

		fmt.Printf("Tunnel '%s' closed successfully\n", name)
		return nil
	},
}

func parseTunnelSpec(spec string) (localPort int, remoteHost string, remotePort int, err error) {
	fmt.Sscanf(spec, "%d:%s:%d", &localPort, &remoteHost, &remotePort)
	return localPort, remoteHost, remotePort, nil
}
