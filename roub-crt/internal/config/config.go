package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Terminal  TerminalConfig  `mapstructure:"terminal"`
	Colorschemes ColorSchemes `mapstructure:"colorschemes"`
	Sessions  SessionsConfig  `mapstructure:"sessions"`
	Connection ConnectionConfig `mapstructure:"connection"`
	Security  SecurityConfig  `mapstructure:"security"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type TerminalConfig struct {
	Font        string `mapstructure:"font"`
	FontSize    int    `mapstructure:"font_size"`
	Scrollback  int    `mapstructure:"scrollback"`
	CursorShape string `mapstructure:"cursor_shape"`
}

type ColorSchemes map[string]ColorScheme

type ColorScheme struct {
	Background  string `mapstructure:"background"`
	Foreground  string `mapstructure:"foreground"`
	Cursor      string `mapstructure:"cursor"`
	Selection   string `mapstructure:"selection"`
	Black       string `mapstructure:"black"`
	Red         string `mapstructure:"red"`
	Green       string `mapstructure:"green"`
	Yellow      string `mapstructure:"yellow"`
	Blue        string `mapstructure:"blue"`
	Magenta     string `mapstructure:"magenta"`
	Cyan        string `mapstructure:"cyan"`
	White       string `mapstructure:"white"`
	BrightBlack string `mapstructure:"bright_black"`
	BrightRed   string `mapstructure:"bright_red"`
	BrightGreen string `mapstructure:"bright_green"`
	BrightYellow string `mapstructure:"bright_yellow"`
	BrightBlue  string `mapstructure:"bright_blue"`
	BrightMagenta string `mapstructure:"bright_magenta"`
	BrightCyan  string `mapstructure:"bright_cyan"`
	BrightWhite string `mapstructure:"bright_white"`
}

type SessionsConfig struct {
	DefaultFolder string `mapstructure:"default_folder"`
	AutoSave      bool   `mapstructure:"auto_save"`
}

type ConnectionConfig struct {
	Timeout        int    `mapstructure:"timeout"`
	Keepalive      int    `mapstructure:"keepalive"`
	DefaultEncoding string `mapstructure:"default_encoding"`
}

type SecurityConfig struct {
	EncryptTransfers    bool `mapstructure:"encrypt_transfers"`
	StrictHostKeyCheck  bool `mapstructure:"strict_host_key_check"`
}

var GlobalConfig *Config

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		configPath = filepath.Join(homeDir, ".roub-crt", "config.yaml")
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	defaultConfig := getDefaultConfig()
	for k, v := range defaultConfig {
		viper.SetDefault(k, v)
	}

	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"app.name":                  "roub-crt",
		"app.version":               "1.0.0",
		"terminal.font":             "Monospace",
		"terminal.font_size":        14,
		"terminal.scrollback":       10000,
		"terminal.cursor_shape":     "block",
		"sessions.default_folder":   "./sessions",
		"sessions.auto_save":        true,
		"connection.timeout":        30,
		"connection.keepalive":      10,
		"connection.default_encoding": "UTF-8",
		"security.encrypt_transfers": true,
		"security.strict_host_key_check": true,
	}
}

func (c *Config) Save(configPath string) error {
	viper.Set("app", c.App)
	viper.Set("terminal", c.Terminal)
	viper.Set("colorschemes", c.Colorschemes)
	viper.Set("sessions", c.Sessions)
	viper.Set("connection", c.Connection)
	viper.Set("security", c.Security)

	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		configPath = filepath.Join(homeDir, ".roub-crt", "config.yaml")
	}

	viper.SetConfigFile(configPath)
	return viper.WriteConfig()
}

func GetDefaultColorSchemes() ColorSchemes {
	return ColorSchemes{
		"default": {
			Background:      "#1E1E1E",
			Foreground:      "#D4D4D4",
			Cursor:          "#FFFFFF",
			Selection:       "#264F78",
			Black:           "#000000",
			Red:             "#CD3131",
			Green:           "#0DBC79",
			Yellow:          "#E5E510",
			Blue:            "#2472C8",
			Magenta:         "#BC3FBC",
			Cyan:            "#11A8CD",
			White:           "#E5E5E5",
			BrightBlack:     "#666666",
			BrightRed:       "#F14C4C",
			BrightGreen:     "#23D18B",
			BrightYellow:    "#F5F543",
			BrightBlue:      "#3B8EEA",
			BrightMagenta:   "#D670D6",
			BrightCyan:      "#29B8DB",
			BrightWhite:     "#FFFFFF",
		},
		"monokai": {
			Background:      "#272822",
			Foreground:      "#F8F8F2",
			Cursor:          "#F8F8F0",
			Selection:       "#49483E",
			Black:           "#272822",
			Red:             "#F92672",
			Green:           "#A6E22E",
			Yellow:          "#F4BF75",
			Blue:            "#66D9EF",
			Magenta:         "#AE81FF",
			Cyan:            "#A1EFE4",
			White:           "#F8F8F2",
			BrightBlack:     "#75715E",
			BrightRed:       "#F92672",
			BrightGreen:     "#A6E22E",
			BrightYellow:    "#F4BF75",
			BrightBlue:      "#66D9EF",
			BrightMagenta:   "#AE81FF",
			BrightCyan:      "#A1EFE4",
			BrightWhite:     "#F9F8F5",
		},
		"solarized_dark": {
			Background:      "#002B36",
			Foreground:      "#839496",
			Cursor:          "#D33682",
			Selection:       "#274642",
			Black:           "#073642",
			Red:             "#DC322F",
			Green:           "#859900",
			Yellow:          "#B58900",
			Blue:            "#268BD2",
			Magenta:         "#D33682",
			Cyan:            "#2AA198",
			White:           "#EEE8D5",
			BrightBlack:     "#586E75",
			BrightRed:       "#CB4B16",
			BrightGreen:     "#586E75",
			BrightYellow:    "#B58900",
			BrightBlue:      "#268BD2",
			BrightMagenta:   "#6C71C4",
			BrightCyan:      "#2AA198",
			BrightWhite:     "#FDF6E3",
		},
		"dracula": {
			Background:      "#282A36",
			Foreground:      "#F8F8F2",
			Cursor:          "#F8F8F2",
			Selection:       "#44475A",
			Black:           "#21222C",
			Red:             "#FF5555",
			Green:           "#50FA7B",
			Yellow:          "#F1FA8C",
			Blue:            "#BD93F9",
			Magenta:         "#FF79C6",
			Cyan:            "#8BE9FD",
			White:           "#F8F8F2",
			BrightBlack:     "#6272A4",
			BrightRed:       "#FF6E6E",
			BrightGreen:     "#69FF94",
			BrightYellow:    "#FFFFA5",
			BrightBlue:      "#D6ACFF",
			BrightMagenta:   "#FF92DF",
			BrightCyan:      "#A4FFFF",
			BrightWhite:     "#FFFFFF",
		},
		"nord": {
			Background:      "#2E3440",
			Foreground:      "#D8DEE9",
			Cursor:          "#D8DEE9",
			Selection:       "#434C5E",
			Black:           "#3B4252",
			Red:             "#BF616A",
			Green:           "#A3BE8C",
			Yellow:          "#EBCB8B",
			Blue:            "#81A1C1",
			Magenta:         "#B48EAD",
			Cyan:            "#88C0D0",
			White:           "#E5E9F0",
			BrightBlack:     "#4C566A",
			BrightRed:       "#BF616A",
			BrightGreen:     "#A3BE8C",
			BrightYellow:    "#EBCB8B",
			BrightBlue:      "#81A1C1",
			BrightMagenta:   "#B48EAD",
			BrightCyan:      "#8FBCBB",
			BrightWhite:     "#ECEFF4",
		},
	}
}
