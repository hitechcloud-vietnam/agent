package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configDir = "/etc/hitechcloudpanel-agent"
const configFile = "config.json"

type Config struct {
	Url         string `json:"url"`
	Secret      string `json:"secret"`
	LogFile     string `json:"log_file"`
	LogToStdout bool   `json:"log_to_stdout"`
	LogLevel    string `json:"log_level"`
	LogMaxSize  int64  `json:"log_max_size_mb"`
	LogMaxFiles int    `json:"log_max_files"`
}

func GetConfig() *Config {
	config := &Config{
		Url:         "",
		Secret:      "",
		LogFile:     "/var/log/hitechcloudpanel-agent/agent.log",
		LogToStdout: true,
		LogLevel:    "info",
		LogMaxSize:  10,
		LogMaxFiles: 5,
	}

	// Create the config directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.Mkdir(configDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			panic(err)
		}
	}

	configPath := fmt.Sprintf("%s/%s", configDir, configFile)

	// Check if the config file exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// Create the config file if it doesn't exist
		createFile, err := os.Create(configPath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			panic(err)
		}
		encoder := json.NewEncoder(createFile)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(config)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			panic(err)
		}
		createFile.Close()
		return config
	}
	if err != nil {
		fmt.Println("Error checking if file exists:", err)
		panic(err)
	}

	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
		panic(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		fmt.Println("Error decoding config:", err)
		panic(err)
	}

	return config
}
