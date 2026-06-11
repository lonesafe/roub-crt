package connection

import (
	"fmt"
	"net"
	"time"
)

type TelnetConn struct {
	Conn    net.Conn
	Timeout time.Duration
}

type TelnetConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

func NewTelnetConfig(host string, port int) *TelnetConfig {
	return &TelnetConfig{
		Host:    host,
		Port:    port,
		Timeout: 30 * time.Second,
	}
}

func ConnectTelnet(config *TelnetConfig) (*TelnetConn, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := net.DialTimeout("tcp", addr, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return &TelnetConn{
		Conn:    conn,
		Timeout: config.Timeout,
	}, nil
}

func (t *TelnetConn) Read(p []byte) (n int, err error) {
	t.Conn.SetReadDeadline(time.Now().Add(t.Timeout))
	return t.Conn.Read(p)
}

func (t *TelnetConn) Write(p []byte) (n int, err error) {
	return t.Conn.Write(p)
}

func (t *TelnetConn) Close() error {
	return t.Conn.Close()
}

func (t *TelnetConn) Login(username, password string) error {
	buf := make([]byte, 1024)

	for {
		n, err := t.Read(buf)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		data := string(buf[:n])

		if contains(data, "login:") {
			t.Write([]byte(username + "\r\n"))
			time.Sleep(100 * time.Millisecond)
		} else if contains(data, "Password:") {
			t.Write([]byte(password + "\r\n"))
			time.Sleep(100 * time.Millisecond)
			break
		}
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
