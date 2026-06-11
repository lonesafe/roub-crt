package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "roub-crt",
	Short: "roub-crt - Professional terminal emulation and file transfer tool",
	Long: `roub-crt is a professional terminal emulation and remote connection management tool.
It supports SSH, Telnet, Serial, Rlogin protocols with advanced features like
file transfer, port forwarding, and secure encrypted connections.`,
	Version: "1.0.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.roub-crt/config.yaml)")
}
