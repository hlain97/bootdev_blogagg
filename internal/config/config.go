package config

import (
	"os"
	"encoding/json"
	"io"
	"path/filepath"
)

type Config struct {
	DBURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error){
	homeDir, err := os.UserHomeDir()

	if err!= nil {
		return "", err
	}

	filePath := filepath.Join(homeDir, configFileName)

	return filePath, nil
}

func write(cfg Config) error{

	filePath, err := getConfigFilePath()

	if err != nil {
    return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)

	if err != nil {
		return err
	}

	return nil

}


func (c *Config) SetUser(userName string) error{
	c.CurrentUserName = userName

	return write(*c)
}

func Read() (Config, error){
	filePath, err := getConfigFilePath()

	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(filePath)

	if err != nil {
		return Config{}, err
	}

	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		return Config{}, err
	}

	var cfg Config

	err = json.Unmarshal(data, &cfg)

	if err != nil {
		return Config{}, err
	}


	return cfg, nil


}