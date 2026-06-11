package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"roub-crt/internal/connection"
)

var connectHost string
var connectPort int
var connectUser string
var connectPassword string
var connectKeyFile string
var connectType string
var connectSerialBaud int
var connectSerialData int
var connectSerialStop int
var connectSerialParity int

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&connectHost, "host", "H", "", "Host to connect to")
	connectCmd.Flags().IntVarP(&connectPort, "port", "p", 0, "Port to connect to (default: 22 for SSH, 23 for Telnet)")
	connectCmd.Flags().StringVarP(&connectUser, "user", "u", "", "Username for authentication")
	connectCmd.Flags().StringVarP(&connectPassword, "password", "P", "", "Password for authentication")
	connectCmd.Flags().StringVarP(&connectKeyFile, "key", "k", "", "Private key file for authentication")
	connectCmd.Flags().StringVarP(&connectType, "type", "t", "ssh", "Connection type (ssh, telnet, serial, rlogin)")
	connectCmd.Flags().IntVar(&connectSerialBaud, "serial-baud", 115200, "Serial baud rate")
	connectCmd.Flags().IntVar(&connectSerialData, "serial-data", 8, "Serial data bits")
	connectCmd.Flags().IntVar(&connectSerialStop, "serial-stop", 1, "Serial stop bits")
	connectCmd.Flags().IntVar(&connectSerialParity, "serial-parity", 0, "Serial parity (0=None, 1=Odd, 2=Even)")
}

var connectCmd = &cobra.Command{
	Use:   "connect [host]",
	Short: "Connect to a remote host",
	Long:  `Establish a terminal connection to a remote host using SSH, Telnet, Serial, or Rlogin protocol.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := args[0]

		if connectPort == 0 {
			switch connectType {
			case "ssh", "ssh1":
				connectPort = 22
			case "telnet":
				connectPort = 23
			case "rlogin":
				connectPort = 513
			default:
				connectPort = 22
			}
		}

		return runConnect(host)
	},
}

func runConnect(host string) error {
	fmt.Printf("Connecting to %s:%d via %s...\n", host, connectPort, connectType)

	switch connectType {
	case "ssh", "ssh1":
		return runSSHConnect(host)
	case "telnet":
		return runTelnetConnect(host)
	case "serial":
		return runSerialConnect(host)
	case "rlogin":
		return runRloginConnect(host)
	default:
		return fmt.Errorf("unsupported connection type: %s", connectType)
	}
}

func runSSHConnect(host string) error {
	sshConn, err := connectSSH(host)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer sshConn.Close()

	session := sshConn.Session
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	TERM := os.Getenv("TERM")
	if TERM == "" {
		TERM = "xterm-256color"
	}

	if err := session.RequestPty(TERM, 80, 40, nil); err != nil {
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	fmt.Println("Connected! Starting shell session...")

	return session.Wait()
}

func runTelnetConnect(host string) error {
	conn, err := connectTelnet(host)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Starting terminal session...")

	go func() {
		io.Copy(os.Stdout, conn.Conn)
	}()
	io.Copy(conn.Conn, os.Stdin)

	return nil
}

func runSerialConnect(host string) error {
	conn, err := connectSerial(host)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Starting terminal session...")

	go func() {
		io.Copy(os.Stdout, conn.RW)
	}()
	io.Copy(conn.RW, os.Stdin)

	return nil
}

func runRloginConnect(host string) error {
	conn, err := connectRlogin(host)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Starting terminal session...")

	go func() {
		io.Copy(os.Stdout, conn.Conn)
	}()
	io.Copy(conn.Conn, os.Stdin)

	return nil
}

func connectSSH(host string) (*connection.SSHConn, error) {
	config := connection.NewSSHConfig(host, connectPort, connectUser, connectPassword)
	config.KeyFile = connectKeyFile

	return connection.ConnectSSH(config)
}

func connectTelnet(host string) (*connection.TelnetConn, error) {
	config := connection.NewTelnetConfig(host, connectPort)
	conn, err := connection.ConnectTelnet(config)
	if err != nil {
		return nil, err
	}

	if connectUser != "" {
		conn.Login(connectUser, connectPassword)
	}

	return conn, nil
}

func connectSerial(port string) (*connection.SerialConn, error) {
	config := connection.NewSerialConfig(port, connectSerialBaud)
	config.DataBits = connectSerialData
	config.StopBits = connectSerialStop
	config.Parity = connectSerialParity

	return connection.ConnectSerial(config)
}

func connectRlogin(host string) (*connection.RloginConn, error) {
	config := connection.NewRloginConfig(host, connectPort, connectUser)
	return connection.ConnectRlogin(config)
}
