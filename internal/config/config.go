package config

import (
	"os"
	"encoding/json"
	"path/filepath"
)

type Config struct {
	DBURL 			string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	config_file, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(config_file)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	config := Config{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (cfg *Config) SetUser(userName string) error {
	cfg.CurrentUserName = userName
	config_file, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(config_file)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(cfg); err != nil {
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(home, configFileName)
	return fullPath, nil
}
