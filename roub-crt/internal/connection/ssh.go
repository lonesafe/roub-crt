package connection

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConn struct {
	Client  *ssh.Client
	Session *ssh.Session
	Config  *SSHConfig
}

type SSHConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	KeyFile    string
	KeyData    []byte
	Passphrase string
	Timeout    time.Duration
}

func NewSSHConfig(host string, port int, username, password string) *SSHConfig {
	return &SSHConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  30 * time.Second,
	}
}

func (c *SSHConfig) addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func ConnectSSH(config *SSHConfig) (*SSHConn, error) {
	var authMethods []ssh.AuthMethod

	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if config.KeyFile != "" || len(config.KeyData) > 0 {
		var signer ssh.Signer
		var err error

		if config.KeyFile != "" {
			keyData, err := os.ReadFile(config.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read key file: %w", err)
			}
			config.KeyData = keyData
		}

		signer, err = ssh.ParsePrivateKeyWithPassphrase(config.KeyData, []byte(config.Passphrase))
		if err != nil {
			signer, err = ssh.ParsePrivateKey(config.KeyData)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		Timeout:         config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", config.addr(), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", config.addr(), err)
	}

	session, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &SSHConn{
		Client:  conn,
		Session: session,
		Config:  config,
	}, nil
}

func (c *SSHConn) StartPTY(TERM string) error {
	if err := c.Session.RequestPty(TERM, 80, 40, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}); err != nil {
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	if err := c.Session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return nil
}

func (c *SSHConn) Close() error {
	if c.Session != nil {
		c.Session.Close()
	}
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}

func (c *SSHConn) StdinPipe() ( <-chan []byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *SSHConn) Wait() error {
	return c.Session.Wait()
}

func SSHKeyScan(host string, port int) (string, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conf := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	_, _, _, err = ssh.NewServerConn(conn, conf)
	if err != nil {
		return "", fmt.Errorf("failed to handshake: %w", err)
	}

	return "", nil
}
