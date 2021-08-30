package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var _config *Config

type Config struct {
	Server   ServerConfig   `json:"server"`
	Gmail    GmailConfig    `json:"gmail"`
	Rollbar  RollbarConfig  `json:"rollbar"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
}

type ServerConfig struct {
	Mode            string `json:"mode"`
	Port            int    `json:"port"`
	Debug           bool   `json:"debug,omitempty"`
	HostUrl         string `json:"host_url"`
	ApiUrl          string `json:"api_url"`
	JWTSecret       string `json:"jwt_secret"`
	JWTHeader       string `json:"jwt_header"`
	JWTBearerPrefix string `json:"jwt_bearer_prefix"`
}

type GmailConfig struct {
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	ClientUsername string `json:"client_username"`
	ClientPassword string `json:"client_password"`
	RedirectUrl    string `json:"redirect_url"`
}

type RollbarConfig struct {
	Environment string `json:"environment"`
	Token       string `json:"token"`
	ServerRoot  string `json:"server_root"`
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Host     string `json:"port"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

func NewConfig(filePath string) *Config {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("can't open config file: %s", err)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("can't read config file: %s", err)
	}

	var config Config
	if err := json.Unmarshal(byteValue, &config); err != nil {
		log.Fatalf("config file unmarshalling error: %s", err)
	}

	_config = &config
	return &config
}

func GetConfig() *Config {
	return _config
}
