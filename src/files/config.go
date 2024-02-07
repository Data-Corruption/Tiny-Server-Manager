package files

import (
	"encoding/json"
	"log"
	"os"
)

var (
	Config     ConfigInterface = ConfigInterface{}
	configPath string
)

type ConfigInterface struct {
	GameExePath      string `json:"game_exe_path"`
	GameSavePath     string `json:"game_save_path"`
	UpdateCommand    string `json:"update_command"`
	DashboardTitle   string `json:"dashboard_title"`
	Port             int    `json:"port"`
	Host             string `json:"host"`
	TrustProxy       bool   `json:"trust_proxy"`
	TLSKeyPath       string `json:"tls_key_path"`
	TLSCertPath      string `json:"tls_cert_path"`
	AdminPassword    string `json:"admin_password"`
	BanDurationHours int    `json:"ban_dur_hours"`
	SessionDurMins   int    `json:"session_dur_mins"`
	LogLevel         string `json:"log_level"`
}

func setDefaultConfigValues() {
	Config = ConfigInterface{}
	Config.TrustProxy = true
	Config.BanDurationHours = 1
	Config.SessionDurMins = 15
	Config.LogLevel = "warn"
}

// LoadConfig loads the configuration from database, or creates a new one if it doesn't exist.
func LoadConfig() {
	configPath = "config.json"

	// If the config file doesn't exist, create it and exit
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Set default config values and save to file
		setDefaultConfigValues()
		SaveConfig()
		log.Printf("Generated config.json at `%s`.\n", configPath)
		os.Exit(0)
	}

	// Read the config file
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %s\n", err)
	}

	// Parse the config file
	err = json.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalf("Error parsing config file: %s\n", err)
	}
}

// SaveConfig saves the configuration to the database.
func SaveConfig() {
	// Encode the config
	data, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		log.Fatalf("Error encoding config: %s\n", err)
	}
	// Write the config to the file
	err = os.WriteFile(configPath, data, 0644) // 0644 - Read/Write for owner, Read for everybody else
	if err != nil {
		log.Fatalf("Error writing config file: %s\n", err)
	}
}
