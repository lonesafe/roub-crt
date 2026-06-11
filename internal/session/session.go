package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ProtocolType string

const (
	ProtocolSSH    ProtocolType = "ssh"
	ProtocolSSH1   ProtocolType = "ssh1"
	ProtocolTelnet ProtocolType = "telnet"
	ProtocolSerial ProtocolType = "serial"
	ProtocolRlogin ProtocolType = "rlogin"
	ProtocolTAPI   ProtocolType = "tapi"
)

type Session struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Folder       string       `json:"folder"`
	Protocol     ProtocolType `json:"protocol"`
	Host         string       `json:"host"`
	Port         int          `json:"port"`
	Username     string       `json:"username"`
	Password     string       `json:"password,omitempty"`
	KeyFile      string       `json:"key_file,omitempty"`
	Encoding     string       `json:"encoding"`
	ColorScheme  string       `json:"color_scheme"`

	SerialConfig *SerialSettings `json:"serial_config,omitempty"`
	TunnelConfig  *TunnelSettings  `json:"tunnel_config,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastUsed  time.Time `json:"last_used,omitempty"`
	UseCount  int       `json:"use_count"`
}

type SerialSettings struct {
	PortName string `json:"port_name"`
	BaudRate int    `json:"baud_rate"`
	DataBits int    `json:"data_bits"`
	StopBits int    `json:"stop_bits"`
	Parity   int    `json:"parity"`
}

type TunnelSettings struct {
	Type     string `json:"type"`
	LocalPort int   `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort int   `json:"remote_port"`
}

type SessionFolder struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Children []*SessionFolder `json:"children,omitempty"`
	Sessions []string        `json:"sessions,omitempty"`
}

type SessionManager struct {
	BaseDir string
	Sessions map[string]*Session
	Folders  map[string]*SessionFolder
}

func NewSessionManager(baseDir string) (*SessionManager, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		baseDir = filepath.Join(homeDir, ".roub-crt", "sessions")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	sm := &SessionManager{
		BaseDir:  baseDir,
		Sessions: make(map[string]*Session),
		Folders:  make(map[string]*SessionFolder),
	}

	sm.Folders["root"] = &SessionFolder{
		Name: "root",
		Path: "/",
	}

	if err := sm.loadAllSessions(); err != nil {
		return sm, nil
	}

	return sm, nil
}

func (sm *SessionManager) loadAllSessions() error {
	files, err := os.ReadDir(sm.BaseDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".json" {
			session, err := sm.LoadSession(filepath.Join(sm.BaseDir, file.Name()))
			if err != nil {
				continue
			}
			sm.Sessions[session.ID] = session
		}
	}

	return nil
}

func (sm *SessionManager) SaveSession(session *Session) error {
	session.UpdatedAt = time.Now()

	if session.ID == "" {
		session.ID = fmt.Sprintf("%d", time.Now().UnixNano())
		session.CreatedAt = time.Now()
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	filename := filepath.Join(sm.BaseDir, session.ID+".json")
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	sm.Sessions[session.ID] = session
	return nil
}

func (sm *SessionManager) LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (sm *SessionManager) DeleteSession(id string) error {
	filename := filepath.Join(sm.BaseDir, id+".json")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	delete(sm.Sessions, id)
	return nil
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	session, ok := sm.Sessions[id]
	return session, ok
}

func (sm *SessionManager) ListSessions() []*Session {
	sessions := make([]*Session, 0, len(sm.Sessions))
	for _, s := range sm.Sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (sm *SessionManager) ListSessionsByFolder(folder string) []*Session {
	sessions := make([]*Session, 0)
	for _, s := range sm.Sessions {
		if s.Folder == folder {
			sessions = append(sessions, s)
		}
	}
	return sessions
}

func (sm *SessionManager) CreateFolder(name, parentPath string) error {
	folder := &SessionFolder{
		Name: name,
		Path: filepath.Join(parentPath, name),
	}

	sm.Folders[folder.Path] = folder
	return nil
}

func (sm *SessionManager) DeleteFolder(path string) error {
	if path == "/" || path == "root" {
		return fmt.Errorf("cannot delete root folder")
	}

	delete(sm.Folders, path)
	return nil
}

func (sm *SessionManager) ImportSessions(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var sessions []*Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return fmt.Errorf("failed to unmarshal sessions: %w", err)
	}

	for _, session := range sessions {
		session.ID = ""
		if err := sm.SaveSession(session); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SessionManager) ExportSessions(ids []string, path string) error {
	sessions := make([]*Session, 0)
	for _, id := range ids {
		if session, ok := sm.Sessions[id]; ok {
			sessions = append(sessions, session)
		}
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (sm *SessionManager) RecordUsage(id string) error {
	session, ok := sm.Sessions[id]
	if !ok {
		return fmt.Errorf("session not found")
	}

	session.UseCount++
	session.LastUsed = time.Now()
	return sm.SaveSession(session)
}

func GetDefaultSessionConfig() *Session {
	return &Session{
		Protocol:    ProtocolSSH,
		Port:        22,
		Encoding:    "UTF-8",
		ColorScheme: "default",
		SerialConfig: &SerialSettings{
			BaudRate: 115200,
			DataBits: 8,
			StopBits: 1,
			Parity:   0,
		},
	}
}
