package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	App    AppConfig    `json:"app"`
	Server ServerConfig `json:"server"`
}

type AppConfig struct {
	DateFormat   string `json:"date_format"`
	LogDirectory string `json:"log_directory"`
	LogFilePath  string `json:"log_file_path"`
	LogFormat    string `json:"log_format"`
}

type ServerConfig struct {
	Mode string `json:"mode"`
	Port int    `json:"port"`
}

func New(filePath string) *Config {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Can't open config.json: %s", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Can't read config.json: %s", err)
	}

	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		log.Fatalf("Error during config unmarshaling: %s", err)
	}

	return &config
}
